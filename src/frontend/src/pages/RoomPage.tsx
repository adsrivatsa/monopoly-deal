import { FormEvent, useEffect, useMemo, useRef, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import {
  Game,
  getCapacityOptions,
  getCapacityRangeForGame,
  getGameDisplayName,
  getGameSettingSelectValues,
  getDefaultSettingsForGame,
  parseGame,
  stringifyGameSettings,
  supportedGames,
  type GameSettingSelectValue,
  type ShortPlayer,
} from "../api/models";
import {
  getRoom,
  leaveRoom,
  readyRoom,
  updateRoomSettings,
  type UpdateRoomSettingsParams,
} from "../api/room";
import { getPlayer } from "../api/player";
import {
  connectRoomSocket,
  decodeRoomServerMessage,
  sendRoomChatMessage,
} from "../api/roomSocket";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "../components/ui/card";
import Button from "../components/ui/button";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "../components/ui/table";

const RoomPage = () => {
  const navigate = useNavigate();
  const { room_id: roomId } = useParams();
  const socketRef = useRef<WebSocket | null>(null);
  const [roomGame, setRoomGame] = useState<Game | null>(null);
  const [roomCapacity, setRoomCapacity] = useState<number | null>(null);
  const [roomSettingSelectValues, setRoomSettingSelectValues] = useState<
    GameSettingSelectValue[]
  >([]);
  const [chatValue, setChatValue] = useState("");
  const [players, setPlayers] = useState<ShortPlayer[]>([]);
  const [currentPlayerId, setCurrentPlayerId] = useState<string | null>(null);
  const [chatMessages, setChatMessages] = useState<
    (
      | {
          id: string;
          kind: "system";
          text: string;
          playerName: string;
          playerImageUrl?: string;
        }
      | {
          id: string;
          kind: "chat";
          text: string;
          playerId: string;
        }
    )[]
  >([]);

  const roomSettingsJson = useMemo(() => {
    if (!roomGame) {
      return "{}";
    }

    if (roomGame === Game.MonopolyDeal) {
      const deckSetting = roomSettingSelectValues.find(
        (setting) => setting.key === "num_decks",
      );
      const parsedDecks = Number.parseInt(deckSetting?.value ?? "", 10);
      return stringifyGameSettings(roomGame, {
        num_decks: Number.isNaN(parsedDecks)
          ? getDefaultSettingsForGame(roomGame).num_decks
          : parsedDecks,
      });
    }

    return stringifyGameSettings(roomGame, getDefaultSettingsForGame(roomGame));
  }, [roomGame, roomSettingSelectValues]);

  const capacityRange = useMemo(() => {
    if (!roomGame) {
      return { min: 2, max: 5 };
    }

    return getCapacityRangeForGame(roomGame, roomSettingsJson);
  }, [roomGame, roomSettingsJson]);

  useEffect(() => {
    if (roomCapacity === null) {
      return;
    }

    if (roomCapacity < capacityRange.min || roomCapacity > capacityRange.max) {
      setRoomCapacity(capacityRange.min);
    }
  }, [capacityRange.max, capacityRange.min, roomCapacity]);

  useEffect(() => {
    let active = true;
    const socket = connectRoomSocket();
    socketRef.current = socket;

    void (async () => {
      const [roomResult, playerResponse] = await Promise.all([
        getRoom(),
        getPlayer(),
      ]);

      if (!active) {
        return;
      }

      if (playerResponse.ok) {
        const currentPlayer = (await playerResponse.json()) as {
          player_id: string;
        };
        if (active) {
          setCurrentPlayerId(currentPlayer.player_id);
        }
      }

      if (!roomResult.ok) {
        return;
      }

      const nextPlayers: ShortPlayer[] = roomResult.data.players.map(
        (player) => {
          return {
            id: player.player_id,
            name: player.display_name,
            imageUrl: player.image_url,
            isHost: player.is_host,
            isReady: player.is_ready,
          };
        },
      );

      const parsedRoomGame = parseGame(roomResult.data.game);
      setRoomGame(parsedRoomGame);
      setRoomCapacity(roomResult.data.capacity);
      setRoomSettingSelectValues(
        parsedRoomGame
          ? getGameSettingSelectValues(parsedRoomGame, roomResult.data.settings)
          : [],
      );
      setPlayers(nextPlayers);
    })();

    console.log("[room-ws] connecting", socket.url);

    socket.onopen = () => {
      console.log("[room-ws] open");
    };

    socket.onmessage = (event) => {
      void (async () => {
        const message = await decodeRoomServerMessage(event.data);

        const joinedPlayer = message?.roomMessage?.playerJoinedRoom?.player;
        if (joinedPlayer) {
          setPlayers((currentPlayers) => {
            const nextPlayer: ShortPlayer = {
              id: joinedPlayer.playerId,
              name: joinedPlayer.displayName,
              imageUrl: joinedPlayer.avatarUrl,
              isHost: joinedPlayer.isHost,
              isReady: joinedPlayer.isReady,
            };

            const existingIndex = currentPlayers.findIndex(
              (player) => player.id === nextPlayer.id,
            );
            if (existingIndex === -1) {
              return [...currentPlayers, nextPlayer];
            }

            const updatedPlayers = [...currentPlayers];
            updatedPlayers[existingIndex] = nextPlayer;
            return updatedPlayers;
          });

          setChatMessages((currentMessages) => {
            return [
              ...currentMessages,
              {
                id: `join-${joinedPlayer.playerId}-${Date.now()}`,
                kind: "system",
                text: `${joinedPlayer.displayName} joined the room`,
                playerName: joinedPlayer.displayName,
                playerImageUrl: joinedPlayer.avatarUrl,
              },
            ];
          });
        }

        const leftPlayer = message?.roomMessage?.playerLeftRoom;
        if (leftPlayer) {
          const leavingPlayer = players.find(
            (player) => player.id === leftPlayer.playedId,
          );

          setPlayers((currentPlayers) => {
            const remainingPlayers = currentPlayers.filter(
              (player) => player.id !== leftPlayer.playedId,
            );

            if (!leftPlayer.newHostPlayerId) {
              return remainingPlayers;
            }

            return remainingPlayers.map((player) => {
              return {
                ...player,
                isHost: player.id === leftPlayer.newHostPlayerId,
              };
            });
          });

          setChatMessages((currentMessages) => {
            return [
              ...currentMessages,
              {
                id: `leave-${leftPlayer.playedId}-${Date.now()}`,
                kind: "system",
                text: `${leavingPlayer?.name ?? "A player"} left the room`,
                playerName: leavingPlayer?.name ?? "Player",
                playerImageUrl: leavingPlayer?.imageUrl,
              },
            ];
          });
        }

        const chatReceived = message?.roomMessage?.chatReceived;
        if (chatReceived) {
          setChatMessages((currentMessages) => {
            return [
              ...currentMessages,
              {
                id: `chat-${chatReceived.playerId}-${Date.now()}`,
                kind: "chat",
                text: chatReceived.payload,
                playerId: chatReceived.playerId,
              },
            ];
          });
        }

        const playerToggledReady = message?.roomMessage?.playerToggledReady;
        if (playerToggledReady) {
          setPlayers((currentPlayers) => {
            return currentPlayers.map((player) => {
              if (player.id !== playerToggledReady.playerId) {
                return player;
              }

              return {
                ...player,
                isReady: playerToggledReady.isReady,
              };
            });
          });
        }

        const settingsUpdated = message?.roomMessage?.settingsUpdated;
        if (settingsUpdated) {
          const nextGame =
            settingsUpdated.game === 0 ? Game.MonopolyDeal : null;

          if (nextGame) {
            setRoomGame(nextGame);
            setRoomCapacity(settingsUpdated.capacity);
            setRoomSettingSelectValues(
              getGameSettingSelectValues(nextGame, settingsUpdated.settings),
            );
          }
        }

        console.log("[room-ws] message", message ?? event.data);
      })();
    };

    socket.onerror = (event) => {
      console.log("[room-ws] error", event);
    };

    socket.onclose = (event) => {
      console.log("[room-ws] close", {
        code: event.code,
        reason: event.reason,
        wasClean: event.wasClean,
      });
    };

    return () => {
      active = false;
      socketRef.current = null;
      socket.close();
    };
  }, [roomId]);

  const handleSendMessage = (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();

    const payload = chatValue.trim();
    if (!payload) {
      return;
    }

    const socket = socketRef.current;
    if (!socket || socket.readyState !== WebSocket.OPEN) {
      return;
    }

    sendRoomChatMessage(socket, payload);
    setChatValue("");
  };

  const handleLeaveRoom = async () => {
    const result = await leaveRoom();

    if (!result.ok) {
      if (result.isTokenError) {
        navigate("/login", { replace: true });
      }
      return;
    }

    navigate("/lobby");
  };

  const handleReadyUp = async () => {
    const result = await readyRoom();

    if (!result.ok && result.isTokenError) {
      navigate("/login", { replace: true });
    }
  };

  const buildSettingsPayload = (
    selectedGame: Game,
    settingValues: GameSettingSelectValue[],
  ): string => {
    if (selectedGame === Game.MonopolyDeal) {
      const deckSetting = settingValues.find((setting) => setting.key === "num_decks");
      const parsedDecks = Number.parseInt(deckSetting?.value ?? "", 10);
      return stringifyGameSettings(selectedGame, {
        num_decks: Number.isNaN(parsedDecks)
          ? getDefaultSettingsForGame(selectedGame).num_decks
          : parsedDecks,
      });
    }

    return stringifyGameSettings(
      selectedGame,
      getDefaultSettingsForGame(selectedGame),
    );
  };

  const persistRoomSettings = async (
    params: UpdateRoomSettingsParams,
  ): Promise<void> => {
    const result = await updateRoomSettings(params);

    if (!result.ok && result.isTokenError) {
      navigate("/login", { replace: true });
    }
  };

  const currentPlayer = players.find((player) => player.id === currentPlayerId);
  const canEditSettings = currentPlayer?.isHost ?? false;

  return (
    <main className="page room-page">
      <section className="room-layout">
        <div className="room-left-panel">
          <Card className="room-left-card room-settings-card">
            <CardHeader>
              <CardTitle>Room settings</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="room-settings-grid">
                <div className="room-settings-row">
                  <label className="room-setting-label" htmlFor="room-game-setting">
                    Game
                  </label>
                  <select
                    id="room-game-setting"
                    className="field-input room-setting-input"
                    value={roomGame ?? ""}
                    disabled={!canEditSettings}
                    onChange={(event) => {
                      if (!canEditSettings) {
                        return;
                      }

                      const nextGame = parseGame(event.target.value);
                      if (!nextGame) {
                        return;
                      }

                      const nextSettings = getGameSettingSelectValues(nextGame, "{}");
                      const nextCapacity = getCapacityRangeForGame(nextGame, "{}").min;

                      setRoomGame(nextGame);
                      setRoomSettingSelectValues(nextSettings);
                      setRoomCapacity(nextCapacity);

                      void persistRoomSettings({
                        capacity: nextCapacity,
                        game: nextGame,
                        settings: buildSettingsPayload(nextGame, nextSettings),
                      });
                    }}
                  >
                    {supportedGames.map((game) => {
                      return (
                        <option key={game} value={game}>
                          {getGameDisplayName(game)}
                        </option>
                      );
                    })}
                  </select>
                </div>

                <div className="room-settings-row">
                  <label className="room-setting-label" htmlFor="room-capacity-setting">
                    Capacity
                  </label>
                  <select
                    id="room-capacity-setting"
                    className="field-input room-setting-input"
                    value={roomCapacity !== null ? String(roomCapacity) : ""}
                    disabled={!canEditSettings}
                    onChange={(event) => {
                      if (!canEditSettings) {
                        return;
                      }

                      const nextCapacity = Number.parseInt(event.target.value, 10);
                      if (Number.isNaN(nextCapacity) || !roomGame) {
                        return;
                      }

                      setRoomCapacity(nextCapacity);

                      void persistRoomSettings({
                        capacity: nextCapacity,
                        game: roomGame,
                        settings: buildSettingsPayload(roomGame, roomSettingSelectValues),
                      });
                    }}
                  >
                    {getCapacityOptions(capacityRange.min, capacityRange.max).map(
                      (option) => {
                        return (
                          <option key={option.value} value={option.value}>
                            {option.label}
                          </option>
                        );
                      },
                    )}
                  </select>
                </div>

                {roomSettingSelectValues.map((setting) => {
                  return (
                    <div key={setting.key} className="room-settings-row">
                      <label className="room-setting-label" htmlFor={`room-setting-${setting.key}`}>
                        {setting.label}
                      </label>
                      <select
                        id={`room-setting-${setting.key}`}
                        className="field-input room-setting-input"
                        value={setting.value}
                        disabled={!canEditSettings}
                        onChange={(event) => {
                          if (!canEditSettings) {
                            return;
                          }

                          const nextValue = event.target.value;
                          setRoomSettingSelectValues((currentSettings) => {
                            return currentSettings.map((currentSetting) => {
                              if (currentSetting.key !== setting.key) {
                                return currentSetting;
                              }

                              return {
                                ...currentSetting,
                                value: nextValue,
                              };
                            });
                          });

                          if (!roomGame) {
                            return;
                          }

                          const nextSettings = roomSettingSelectValues.map((currentSetting) => {
                            if (currentSetting.key !== setting.key) {
                              return currentSetting;
                            }

                            return {
                              ...currentSetting,
                              value: nextValue,
                            };
                          });

                          const nextSettingsPayload = buildSettingsPayload(
                            roomGame,
                            nextSettings,
                          );
                          const nextCapacityRange = getCapacityRangeForGame(
                            roomGame,
                            nextSettingsPayload,
                          );

                          let nextCapacity = roomCapacity;
                          if (
                            nextCapacity === null ||
                            nextCapacity < nextCapacityRange.min ||
                            nextCapacity > nextCapacityRange.max
                          ) {
                            nextCapacity = nextCapacityRange.min;
                            setRoomCapacity(nextCapacityRange.min);
                          }

                          if (nextCapacity !== null) {
                            void persistRoomSettings({
                              capacity: nextCapacity,
                              game: roomGame,
                              settings: nextSettingsPayload,
                            });
                          }
                        }}
                      >
                        {setting.options.map((option) => {
                          return (
                            <option key={option.value} value={option.value}>
                              {option.label}
                            </option>
                          );
                        })}
                      </select>
                    </div>
                  );
                })}
              </div>

              {!canEditSettings ? (
                <p className="room-settings-readonly-note">
                  Only the host can change room settings.
                </p>
              ) : null}
            </CardContent>
          </Card>

          <Card className="room-left-card room-players-card">
            <CardHeader>
              <CardTitle>Current players</CardTitle>
            </CardHeader>
            <CardContent className="room-players-content">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead className="room-players-header-player">
                      Player
                    </TableHead>
                    <TableHead className="room-players-header-ready">
                      Ready
                    </TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {players.map((player) => (
                    <TableRow key={player.id}>
                      <TableCell>
                        <div className="host-cell">
                          <span
                            className="player-host-badge"
                            aria-hidden="true"
                          >
                            {player.isHost ? "👑" : ""}
                          </span>
                          <img
                            src={player.imageUrl}
                            alt={player.name}
                            className="host-avatar"
                            loading="lazy"
                            referrerPolicy="no-referrer"
                          />
                          <span>{player.name}</span>
                        </div>
                      </TableCell>
                      <TableCell>
                        <div className="player-ready-cell">
                          <span
                            className={`player-ready-status${
                              player.isReady ? " is-ready" : " is-not-ready"
                            }`}
                          >
                            {player.isReady ? "Ready" : "Not ready"}
                          </span>
                        </div>
                      </TableCell>
                    </TableRow>
                  ))}
                  {players.length === 0 ? (
                    <TableRow>
                      <TableCell>Loading players...</TableCell>
                      <TableCell>-</TableCell>
                    </TableRow>
                  ) : null}
                </TableBody>
              </Table>

              <Button
                size="lg"
                className="room-ready-button"
                onClick={() => {
                  void handleReadyUp();
                }}
                disabled={!currentPlayer}
              >
                {currentPlayer?.isReady ? "Unready" : "Ready up"}
              </Button>
            </CardContent>
          </Card>
        </div>

        <div className="room-chat-column">
          <Card className="room-chat-panel">
            <CardHeader>
              <CardTitle>Room chat</CardTitle>
            </CardHeader>
            <CardContent className="room-chat-content">
              <div className="chat-log">
                {chatMessages.length === 0 ? (
                  <p className="chat-message chat-message--empty">
                    No new events.
                  </p>
                ) : (
                  chatMessages.map((chatMessage) => {
                    if (chatMessage.kind === "system") {
                      return (
                        <div key={chatMessage.id} className="chat-event-join">
                          {chatMessage.playerImageUrl ? (
                            <img
                              src={chatMessage.playerImageUrl}
                              alt={chatMessage.playerName}
                              className="host-avatar"
                              loading="lazy"
                              referrerPolicy="no-referrer"
                            />
                          ) : null}
                          <p className="chat-message chat-message--system">
                            {chatMessage.text}
                          </p>
                        </div>
                      );
                    }

                    const chatPlayer = players.find(
                      (player) => player.id === chatMessage.playerId,
                    );

                    return (
                      <p key={chatMessage.id} className="chat-message">
                        <span className="chat-message__author">
                          {chatPlayer?.name ?? "Player"}:
                        </span>{" "}
                        {chatMessage.text}
                      </p>
                    );
                  })
                )}
              </div>

              <form className="chat-input-row" onSubmit={handleSendMessage}>
                <input
                  className="field-input"
                  placeholder="Type a message..."
                  value={chatValue}
                  onChange={(event) => setChatValue(event.target.value)}
                />
              </form>
            </CardContent>
          </Card>

          <Button variant="outline" onClick={handleLeaveRoom}>
            Leave room
          </Button>
        </div>
      </section>
    </main>
  );
};

export default RoomPage;
