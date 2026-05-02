import { useEffect, useMemo, useRef, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import {
  Game,
  getCapacityRangeForGame,
  getGameDisplayName,
  getGameSettingDefinition,
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
  joinRoom,
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

const formatSettingsChangeMessage = (changes: string[]): string => {
  return changes.join("\n");
};

const GROUP_ORDER = [
  "Game Speed",
  "Deal Rules",
  "Card Rules",
  "Win Conditions",
  "Deck Rules",
];

const RoomPage = () => {
  const navigate = useNavigate();
  const { room_id: roomId } = useParams();
  const socketRef = useRef<WebSocket | null>(null);
  const roomGameRef = useRef<Game | null>(null);
  const roomCapacityRef = useRef<number | null>(null);
  const roomSettingSelectValuesRef = useRef<GameSettingSelectValue[]>([]);
  const playersRef = useRef<ShortPlayer[]>([]);
  const [roomGame, setRoomGame] = useState<Game | null>(null);
  const [roomCapacity, setRoomCapacity] = useState<number | null>(null);
  const [capacityInputText, setCapacityInputText] = useState<string | null>(null);
  const [settingInputTexts, setSettingInputTexts] = useState<Record<string, string>>({});
  const [roomSettingSelectValues, setRoomSettingSelectValues] = useState<
    GameSettingSelectValue[]
  >([]);
  const [players, setPlayers] = useState<ShortPlayer[]>([]);
  const [currentPlayerId, setCurrentPlayerId] = useState<string | null>(null);
  const [modalError, setModalError] = useState<ApiErrorPayload | null>(null);
  const [isBootstrappingRoom, setIsBootstrappingRoom] = useState(true);
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
          kind: "player-event";
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
    roomGameRef.current = roomGame;
  }, [roomGame]);

  useEffect(() => {
    roomCapacityRef.current = roomCapacity;
  }, [roomCapacity]);

  useEffect(() => {
    roomSettingSelectValuesRef.current = roomSettingSelectValues;
  }, [roomSettingSelectValues]);

  useEffect(() => {
    playersRef.current = players;
  }, [players]);

  const groupedRoomSettings = useMemo(() => {
    const groups: Map<string, typeof roomSettingSelectValues> = new Map();
    for (const groupName of GROUP_ORDER) {
      groups.set(groupName, []);
    }

    for (const setting of roomSettingSelectValues) {
      const groupName = setting.group ?? "Other";
      if (!groups.has(groupName)) {
        groups.set(groupName, []);
      }
      groups.get(groupName)!.push(setting);
    }

    return [...groups.entries()].filter(([, settings]) => settings.length > 0);
  }, [roomSettingSelectValues]);

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
    let socket: WebSocket | null = null;

    if (!roomId) {
      navigate("/lobby", { replace: true });
      return () => {
        active = false;
      };
    }

    setIsBootstrappingRoom(true);

    void (async () => {
      const joinResult = await joinRoom(roomId);
      if (!active) {
        return;
      }

      if (!joinResult.ok) {
        if (joinResult.error?.code === "SER002") {
          // Already joined this room; continue loading the room page.
        } else if (joinResult.isTokenError) {
          navigate("/login", { replace: true });
          return;
        } else {
          setModalError(
            joinResult.error ?? {
              message: "Could not join room.",
              status: 500,
              code: "UNKNOWN",
            },
          );
          setIsBootstrappingRoom(false);
          return;
        }
      }

      socket = connectRoomSocket();
      socketRef.current = socket;
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
                  kind: "player-event" as const,
                  text: `joined the room`,
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
                  kind: "player-event" as const,
                  text: `left the room`,
                  playerName: leavingPlayer?.name ?? "A player",
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
              settingsUpdated.game === 0
                ? (roomGameRef.current ?? Game.MonopolyDeal)
                : Game.MonopolyDeal;

            if (nextGame) {
              const nextSettingSelectValues = getGameSettingSelectValues(
                nextGame,
                settingsUpdated.settings,
              );

              const changes: string[] = [];

              if (roomGameRef.current && roomGameRef.current !== nextGame) {
                changes.push(
                  `Game: ${getGameDisplayName(roomGameRef.current)}, ${getGameDisplayName(nextGame)}`,
                );
              }

              if (
                roomCapacityRef.current !== null &&
                roomCapacityRef.current !== settingsUpdated.capacity
              ) {
                changes.push(`Capacity: ${roomCapacityRef.current} -> ${settingsUpdated.capacity}`);
              }

              for (const nextSetting of nextSettingSelectValues) {
                const previousSetting = roomSettingSelectValuesRef.current.find((setting) => {
                  return setting.key === nextSetting.key;
                });
                if (!previousSetting || previousSetting.value === nextSetting.value) {
                  continue;
                }

                const settingDefinition = getGameSettingDefinition(nextGame, nextSetting.key);
                const previousLabel =
                  previousSetting.options.find((option) => option.value === previousSetting.value)
                    ?.label ?? previousSetting.value;
                const nextLabel =
                  nextSetting.options.find((option) => option.value === nextSetting.value)
                    ?.label ?? nextSetting.value;

                changes.push(
                  `${settingDefinition?.label ?? nextSetting.label}: ${previousLabel} -> ${nextLabel}`,
                );
              }

              setRoomGame(nextGame);
              setRoomCapacity(settingsUpdated.capacity);
              setRoomSettingSelectValues(nextSettingSelectValues);

              if (changes.length > 0) {
                const hostPlayer = playersRef.current.find((player) => player.isHost);
                setChatMessages((currentMessages) => {
                  return [
                    ...currentMessages,
                    {
                      id: `settings-updated-${Date.now()}`,
                      kind: "system",
                      text: formatSettingsChangeMessage(changes),
                      playerName: hostPlayer?.name ?? "Host",
                      playerImageUrl: hostPlayer?.imageUrl,
                    },
                  ];
                });
              }
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
        setModalError(
          roomResult.error ?? {
            message: "Could not load room details.",
            status: 500,
            code: "UNKNOWN",
          },
        );
        setIsBootstrappingRoom(false);
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
      setIsBootstrappingRoom(false);
    })();

    return () => {
      active = false;
      socketRef.current = null;
      socket?.close();
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

    if (result.ok) {
      const nextSettingSelectValues = getGameSettingSelectValues(
        params.game,
        params.settings,
      );
      const changes: string[] = [];

      if (roomGameRef.current && roomGameRef.current !== params.game) {
        changes.push(
          `Game: ${getGameDisplayName(roomGameRef.current)}, ${getGameDisplayName(params.game)}`,
        );
      }

      if (
        roomCapacityRef.current !== null &&
        roomCapacityRef.current !== params.capacity
      ) {
        changes.push(`Capacity: ${roomCapacityRef.current} -> ${params.capacity}`);
      }

      for (const nextSetting of nextSettingSelectValues) {
        const previousSetting = roomSettingSelectValuesRef.current.find((setting) => {
          return setting.key === nextSetting.key;
        });
        if (!previousSetting || previousSetting.value === nextSetting.value) {
          continue;
        }

        const settingDefinition = getGameSettingDefinition(params.game, nextSetting.key);
        const previousLabel =
          previousSetting.options.find((option) => option.value === previousSetting.value)
            ?.label ?? previousSetting.value;
        const nextLabel =
          nextSetting.options.find((option) => option.value === nextSetting.value)
            ?.label ?? nextSetting.value;

        changes.push(
          `${settingDefinition?.label ?? nextSetting.label}: ${previousLabel} -> ${nextLabel}`,
        );
      }

      if (changes.length > 0) {
        const hostPlayer = playersRef.current.find((player) => player.isHost);
        setChatMessages((currentMessages) => {
          return [
            ...currentMessages,
            {
              id: `settings-updated-local-${Date.now()}`,
              kind: "system",
              text: formatSettingsChangeMessage(changes),
              playerName: hostPlayer?.name ?? "Host",
              playerImageUrl: hostPlayer?.imageUrl,
            },
          ];
        });
      }
    }

    if (!result.ok && result.isTokenError) {
      navigate("/login", { replace: true });
    }
  };

  const currentPlayer = players.find((player) => player.id === currentPlayerId);
  const canEditSettings = currentPlayer?.isHost ?? false;
  const everyoneReady =
    players.length > 0 &&
    players.every((player) => player.isHost || player.isReady);

  if (isBootstrappingRoom) {
    return (
      <main className="page room-page">
        <section className="room-layout">
          <Card className="room-left-card">
            <CardHeader>
              <CardTitle>Joining room...</CardTitle>
            </CardHeader>
            <CardContent>
              <p>Please wait while we connect you to this room.</p>
            </CardContent>
          </Card>
        </section>
      </main>
    );
  }

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
                  <div className="room-slider-control">
                    <span className="room-slider-min">{capacityRange.min}</span>
                    <input
                      id="room-capacity-setting"
                      type="range"
                      className="room-slider"
                      min={capacityRange.min}
                      max={capacityRange.max}
                      step={1}
                      value={roomCapacity ?? capacityRange.min}
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
                    />
                    <span className="room-slider-max">{capacityRange.max}</span>
                    <input
                      type="number"
                      className="room-slider-value-input"
                      disabled={!canEditSettings}
                      value={capacityInputText ?? String(roomCapacity ?? capacityRange.min)}
                      min={capacityRange.min}
                      max={capacityRange.max}
                      onChange={(event) => {
                        setCapacityInputText(event.target.value);
                      }}
                      onBlur={() => {
                        const parsed = Number.parseInt(capacityInputText ?? "", 10);
                        const clamped = Number.isNaN(parsed)
                          ? (roomCapacity ?? capacityRange.min)
                          : Math.min(capacityRange.max, Math.max(capacityRange.min, parsed));
                        setCapacityInputText(null);
                        if (clamped === roomCapacity || !roomGame) return;
                        setRoomCapacity(clamped);
                        void persistRoomSettings({
                          capacity: clamped,
                          game: roomGame,
                          settings: buildSettingsPayload(roomGame, roomSettingSelectValues),
                        });
                      }}
                      onKeyDown={(event) => {
                        if (event.key === "Enter") {
                          event.currentTarget.blur();
                        }
                      }}
                    />
                  </div>
                </div>

                {groupedRoomSettings.map(([groupName, groupSettings]) => (
                  <div key={groupName} className="room-settings-group">
                    <p className="room-settings-group-label">{groupName}</p>
                    {groupSettings.map((setting) => {
                      const isSpeed = setting.key === "speed";
                      const settingMin = Number(setting.options[0]?.value ?? 0);
                      const settingMax = Number(setting.options[setting.options.length - 1]?.value ?? settingMin);

                      const commitSettingValue = (rawValue: string, key: string) => {
                        const parsed = Number.parseInt(rawValue, 10);
                        const def = getGameSettingDefinition(roomGame!, key);
                        const min = def?.min ?? settingMin;
                        const max = def?.max ?? settingMax;
                        const currentNumeric = Number.parseInt(
                          roomSettingSelectValues.find((s) => s.key === key)?.value ?? "",
                          10,
                        );
                        const clamped = Number.isNaN(parsed)
                          ? currentNumeric
                          : Math.min(max, Math.max(min, parsed));
                        const nextValue = String(clamped);

                        setSettingInputTexts((prev) => {
                          const next = { ...prev };
                          delete next[key];
                          return next;
                        });

                        if (nextValue === String(currentNumeric) || !roomGame) return;

                        const nextSettings = roomSettingSelectValues.map((s) =>
                          s.key !== key ? s : { ...s, value: nextValue },
                        );
                        setRoomSettingSelectValues(nextSettings);

                        const nextSettingsPayload = buildSettingsPayload(roomGame, nextSettings);
                        const nextCapacityRange = getCapacityRangeForGame(roomGame, nextSettingsPayload);
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
                      };

                      const applySelectChange = (nextValue: string) => {
                        if (!canEditSettings || !roomGame) return;
                        const nextSettings = roomSettingSelectValues.map((s) =>
                          s.key !== setting.key ? s : { ...s, value: nextValue },
                        );
                        setRoomSettingSelectValues(nextSettings);
                        const nextSettingsPayload = buildSettingsPayload(roomGame, nextSettings);
                        const nextCapacityRange = getCapacityRangeForGame(roomGame, nextSettingsPayload);
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
                      };

                      return (
                        <div key={setting.key} className="room-settings-row">
                          <label
                            className="room-setting-label"
                            htmlFor={`room-setting-${setting.key}`}
                          >
                            {setting.label}
                          </label>

                          {isSpeed ? (
                            <div
                              className="room-speed-group"
                              role="group"
                              aria-label="Speed"
                            >
                              {setting.options.map((option) => (
                                <button
                                  key={option.value}
                                  type="button"
                                  className={`room-speed-option${setting.value === option.value ? " is-active" : ""}`}
                                  disabled={!canEditSettings}
                                  onClick={() => applySelectChange(option.value)}
                                >
                                  {option.label}
                                </button>
                              ))}
                            </div>
                          ) : (
                            <div className="room-slider-control">
                              <span className="room-slider-min">{settingMin}</span>
                              <input
                                id={`room-setting-${setting.key}`}
                                type="range"
                                className="room-slider"
                                min={settingMin}
                                max={settingMax}
                                step={1}
                                value={Number.parseInt(setting.value, 10) || settingMin}
                                disabled={!canEditSettings}
                                onChange={(event) => {
                                  if (!canEditSettings) return;
                                  applySelectChange(event.target.value);
                                }}
                              />
                              <span className="room-slider-max">{settingMax}</span>
                              <input
                                type="number"
                                className="room-slider-value-input"
                                disabled={!canEditSettings}
                                value={
                                  settingInputTexts[setting.key] ?? setting.value
                                }
                                min={settingMin}
                                max={settingMax}
                                onChange={(event) => {
                                  setSettingInputTexts((prev) => ({
                                    ...prev,
                                    [setting.key]: event.target.value,
                                  }));
                                }}
                                onBlur={() => {
                                  commitSettingValue(
                                    settingInputTexts[setting.key] ?? setting.value,
                                    setting.key,
                                  );
                                }}
                                onKeyDown={(event) => {
                                  if (event.key === "Enter") {
                                    event.currentTarget.blur();
                                  }
                                }}
                              />
                            </div>
                          )}
                        </div>
                      );
                    })}
                  </div>
                ))}
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
if (chatMessage.kind === "player-event") {
                return (
                  <article className="room-chat-event-line">
                    {chatMessage.playerImageUrl ? (
                      <img
                        src={chatMessage.playerImageUrl}
                        alt={chatMessage.playerName}
                        className="game-chat-line__avatar"
                        loading="lazy"
                        referrerPolicy="no-referrer"
                      />
                    ) : null}
                    <span className="game-chat-line__author">
                      {chatMessage.playerName}
                    </span>
                    <span className="room-chat-event-line__action">
                      {chatMessage.text}
                    </span>
                  </article>
                );
              }

              if (chatMessage.kind === "system") {
                const isMultiChange = chatMessage.text.split("\n").length > 1;
                return (
                  <article className="room-chat-settings-line">
                    <div className="room-chat-settings-line__header">
                      {chatMessage.playerImageUrl ? (
                        <img
                          src={chatMessage.playerImageUrl}
                          alt={chatMessage.playerName}
                          className="game-chat-line__avatar"
                          loading="lazy"
                          referrerPolicy="no-referrer"
                        />
                      ) : null}
                      <span className="game-chat-line__author">
                        {chatMessage.playerName}
                      </span>
                      <span className="room-chat-settings-line__action">
                        {" changed setting"}{isMultiChange ? "s" : ""}:
                      </span>
                    </div>
                    <p className="room-chat-settings-line__details">
                      {chatMessage.text}
                    </p>
                  </article>
                );
              }

              const chatPlayer = players.find(
                (player) => player.id === chatMessage.playerId,
              );

              return (
                <article className="game-chat-line">
                  {chatPlayer?.imageUrl ? (
                    <img
                      src={chatPlayer.imageUrl}
                      alt={chatPlayer.name}
                      className="game-chat-line__avatar"
                      loading="lazy"
                      referrerPolicy="no-referrer"
                    />
                  ) : null}
                  <p className="chat-message game-chat-line__message">
                    <span className="game-chat-line__author">
                      {chatPlayer?.name ?? "Player"}:
                    </span>{" "}
                    {chatMessage.text}
                  </p>
                </article>
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
