import { useEffect, useMemo, useRef, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import {
  Game,
  getCapacityOptions,
  getCapacityRangeForGame,
  getGameDisplayName,
  getGameSettingSelectValues,
  getDefaultSettingsForGame,
  parseGame,
  parseGameSettings,
  stringifyGameSettings,
  supportedGames,
  type GameSettingSelectValue,
  type ShortPlayer,
} from "../api/models";
import {
  getRoom,
  leaveRoom,
  readyRoom,
  startGame,
  updateRoomSettings,
  type UpdateRoomSettingsParams,
} from "../api/room";
import type { ApiErrorPayload } from "../api/client";
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
import ErrorModal from "../components/ui/error-modal";
import ChatBox from "../components/chat/ChatBox";

const RoomPage = () => {
  const navigate = useNavigate();
  const { room_id: roomId } = useParams();
  const socketRef = useRef<WebSocket | null>(null);
  const [roomGame, setRoomGame] = useState<Game | null>(null);
  const [roomCapacity, setRoomCapacity] = useState<number | null>(null);
  const [roomSettingSelectValues, setRoomSettingSelectValues] = useState<
    GameSettingSelectValue[]
  >([]);
  const [players, setPlayers] = useState<ShortPlayer[]>([]);
  const [currentPlayerId, setCurrentPlayerId] = useState<string | null>(null);
  const [modalError, setModalError] = useState<ApiErrorPayload | null>(null);
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

  const getDefaultSettingsPayload = (selectedGame: Game): Uint8Array => {
    return stringifyGameSettings(
      selectedGame,
      getDefaultSettingsForGame(selectedGame),
    );
  };

  const buildSettingsPayload = (
    selectedGame: Game,
    settingValues: GameSettingSelectValue[],
  ): Uint8Array => {
    const candidateSettings = Object.fromEntries(
      settingValues.map((setting) => {
        const parsedValue = Number.parseInt(setting.value, 10);
        return [
          setting.key,
          Number.isNaN(parsedValue)
            ? Number.parseInt(setting.options[0]?.value ?? "0", 10)
            : parsedValue,
        ];
      }),
    );

    return stringifyGameSettings(
      selectedGame,
      parseGameSettings(selectedGame, JSON.stringify(candidateSettings)),
    );
  };

  const roomSettingsPayload = useMemo(() => {
    if (!roomGame) {
      return new Uint8Array(0);
    }

    return buildSettingsPayload(roomGame, roomSettingSelectValues);
  }, [roomGame, roomSettingSelectValues]);

  const capacityRange = useMemo(() => {
    if (!roomGame) {
      return { min: 2, max: 5 };
    }

    return getCapacityRangeForGame(roomGame, roomSettingsPayload);
  }, [roomGame, roomSettingsPayload]);

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
            settingsUpdated.game === 0 ? Game.MonopolyDeal : roomGame;

          if (nextGame) {
            setRoomGame(nextGame);
            setRoomCapacity(settingsUpdated.capacity);
            setRoomSettingSelectValues(
              getGameSettingSelectValues(nextGame, settingsUpdated.settings),
            );
          }
        }

        const gameStarted = message?.roomMessage?.gameStarted;
        if (gameStarted?.gameId) {
          navigate(`/game/${gameStarted.gameId}`);
          return;
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
  }, [navigate, roomId]);

  const handleSendMessage = (payload: string) => {
    if (!payload) {
      return;
    }

    const socket = socketRef.current;
    if (!socket || socket.readyState !== WebSocket.OPEN) {
      return;
    }

    sendRoomChatMessage(socket, payload);
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

  const handleStartGame = async (): Promise<void> => {
    if (!everyoneReady) {
      setModalError({
        message: "Not every player has readied up.",
        status: 400,
        code: "PLAYERS_NOT_READY",
      });
      return;
    }

    await startGame();

    return;
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
  const everyoneReady =
    players.length > 0 &&
    players.every((player) => player.isHost || player.isReady);

  return (
    <main className="page room-page">
      <section className="room-layout">
        <div className="room-left-panel">
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
                              player.isHost || player.isReady
                                ? " is-ready"
                                : " is-not-ready"
                            }`}
                          >
                            {player.isHost || player.isReady
                              ? "Ready"
                              : "Not ready"}
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

              <div className="room-action-buttons">
                {!currentPlayer?.isHost ? (
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
                ) : null}

                {currentPlayer?.isHost ? (
                  <Button
                    size="lg"
                    className="room-start-button"
                    disabled={!everyoneReady}
                    onClick={() => {
                      void handleStartGame();
                    }}
                  >
                    Start game
                  </Button>
                ) : null}
              </div>
            </CardContent>
          </Card>

          <Card className="room-left-card room-settings-card">
            <CardHeader>
              <CardTitle>Room settings</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="room-settings-grid">
                <div className="room-settings-row">
                  <label
                    className="room-setting-label"
                    htmlFor="room-game-setting"
                  >
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

                      const defaultSettingsPayload =
                        getDefaultSettingsPayload(nextGame);
                      const nextSettings = getGameSettingSelectValues(
                        nextGame,
                        defaultSettingsPayload,
                      );
                      const nextCapacity = getCapacityRangeForGame(
                        nextGame,
                        defaultSettingsPayload,
                      ).min;

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
                  <label
                    className="room-setting-label"
                    htmlFor="room-capacity-setting"
                  >
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

                      const nextCapacity = Number.parseInt(
                        event.target.value,
                        10,
                      );
                      if (Number.isNaN(nextCapacity) || !roomGame) {
                        return;
                      }

                      setRoomCapacity(nextCapacity);

                      void persistRoomSettings({
                        capacity: nextCapacity,
                        game: roomGame,
                        settings: buildSettingsPayload(
                          roomGame,
                          roomSettingSelectValues,
                        ),
                      });
                    }}
                  >
                    {getCapacityOptions(
                      capacityRange.min,
                      capacityRange.max,
                    ).map((option) => {
                      return (
                        <option key={option.value} value={option.value}>
                          {option.label}
                        </option>
                      );
                    })}
                  </select>
                </div>

                {roomSettingSelectValues.map((setting) => {
                  return (
                    <div key={setting.key} className="room-settings-row">
                      <label
                        className="room-setting-label"
                        htmlFor={`room-setting-${setting.key}`}
                      >
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

                          const nextSettings = roomSettingSelectValues.map(
                            (currentSetting) => {
                              if (currentSetting.key !== setting.key) {
                                return currentSetting;
                              }

                              return {
                                ...currentSetting,
                                value: nextValue,
                              };
                            },
                          );

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
        </div>

        <div className="room-chat-column">
          <ChatBox
            title="Room chat"
            messages={chatMessages}
            onSendMessage={handleSendMessage}
            getMessageKey={(message) => message.id}
            renderMessage={(chatMessage) => {
              if (chatMessage.kind === "system") {
                return (
                  <div className="chat-event-join">
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
                <p className="chat-message">
                  <span className="chat-message__author">
                    {chatPlayer?.name ?? "Player"}:
                  </span>{" "}
                  {chatMessage.text}
                </p>
              );
            }}
          />

          <Button variant="outline" onClick={handleLeaveRoom}>
            Leave room
          </Button>
        </div>
      </section>

      {modalError ? (
        <ErrorModal error={modalError} onClose={() => setModalError(null)} />
      ) : null}
    </main>
  );
};

export default RoomPage;
