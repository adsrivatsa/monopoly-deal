import { useEffect, useRef, useState } from "react";
import { useNavigate } from "react-router-dom";
import {
  createRoom,
  CreateRoomParams,
  getGame,
  getRoom,
  leaveRoom,
  listRooms,
  RoomListItem,
} from "../api/room";
import {
  Game,
  getGameDisplayName,
  parseGame,
  supportedGames,
} from "../api/models";
import Button from "../components/ui/button";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "../components/ui/card";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "../components/ui/table";
import ErrorModal from "../components/ui/error-modal";
import { ApiErrorPayload } from "../api/client";
import CreateRoomModal from "../components/ui/create-room-modal";
import {
  connectQuickPlayRoomSocket,
  decodeRoomServerMessage,
} from "../api/roomSocket";

const PAGE_SIZE = 10;
const SEARCH_DEBOUNCE_MS = 350;
const ROOMS_POLL_INTERVAL_MS = 10000;

const copyTextToClipboard = async (text: string): Promise<boolean> => {
  if (navigator.clipboard?.writeText) {
    try {
      await navigator.clipboard.writeText(text);
      return true;
    } catch {
      // Fall back for browsers that block the async clipboard API.
    }
  }

  const textarea = document.createElement("textarea");
  textarea.value = text;
  textarea.setAttribute("readonly", "");
  textarea.style.position = "fixed";
  textarea.style.opacity = "0";
  document.body.appendChild(textarea);
  textarea.select();

  try {
    return document.execCommand("copy");
  } finally {
    document.body.removeChild(textarea);
  }
};

const gameFilterOptions: Array<{ value: Game; label: string }> =
  supportedGames.map((game) => {
    return {
      value: game,
      label: getGameDisplayName(game),
    };
  });

type QuickPlayPlayer = {
  id: string;
  displayName: string;
  avatarUrl: string;
};

type QuickPlayRoom = {
  roomId: string;
  displayName: string;
  capacity: number;
  players: QuickPlayPlayer[];
};

const LobbyPage = () => {
  const navigate = useNavigate();
  const isMountedRef = useRef(true);
  const quickPlayCloseExpectedRef = useRef(false);
  const quickPlaySocketRef = useRef<WebSocket | null>(null);
  const [rooms, setRooms] = useState<RoomListItem[]>([]);
  const [totalCount, setTotalCount] = useState(0);
  const [offset, setOffset] = useState(0);
  const [searchInput, setSearchInput] = useState("");
  const [debouncedSearch, setDebouncedSearch] = useState<string | null>(null);
  const [gameFilter, setGameFilter] = useState<Game | null>(null);
  const [quickPlayGame, setQuickPlayGame] = useState<Game>(
    supportedGames[0] ?? Game.MonopolyDeal,
  );
  const [isLoading, setIsLoading] = useState(true);
  const [apiError, setApiError] = useState<ApiErrorPayload | null>(null);
  const [isCreateRoomOpen, setIsCreateRoomOpen] = useState(false);
  const [isQuickPlayWaiting, setIsQuickPlayWaiting] = useState(false);
  const [quickPlayRoom, setQuickPlayRoom] = useState<QuickPlayRoom | null>(
    null,
  );
  const [refreshCycle, setRefreshCycle] = useState(0);
  const [manualRefreshCount, setManualRefreshCount] = useState(0);
  const [activeRoom, setActiveRoom] = useState<{
    roomId: string;
    displayName: string;
  } | null>(null);

  useEffect(() => {
    isMountedRef.current = true;
    quickPlayCloseExpectedRef.current = false;

    return () => {
      isMountedRef.current = false;
      quickPlayCloseExpectedRef.current = true;
      quickPlaySocketRef.current?.close();
      quickPlaySocketRef.current = null;
    };
  }, []);

  useEffect(() => {
    const timeoutId = window.setTimeout(() => {
      const normalizedSearch = searchInput.trim();
      setDebouncedSearch(normalizedSearch.length > 0 ? normalizedSearch : null);
    }, SEARCH_DEBOUNCE_MS);

    return () => {
      window.clearTimeout(timeoutId);
    };
  }, [searchInput]);

  useEffect(() => {
    setOffset(0);
  }, [debouncedSearch, gameFilter]);

  useEffect(() => {
    let active = true;

    const fetchActiveGame = async () => {
      const result = await getGame();

      if (!active) {
        return;
      }

      if (!result.ok) {
        if (result.isTokenError) {
          navigate("/login", { replace: true });
        }
        return;
      }

      navigate(`/game/${result.data.game_id}`, { replace: true });
    };

    void fetchActiveGame();

    return () => {
      active = false;
    };
  }, [navigate]);

  useEffect(() => {
    let active = true;

    const fetchActiveRoom = async () => {
      const result = await getRoom();

      if (!active) {
        return;
      }

      if (!result.ok) {
        if (result.isTokenError) {
          navigate("/login", { replace: true });
        }
        return;
      }

      setActiveRoom({
        roomId: result.data.room_id,
        displayName: result.data.display_name,
      });
    };

    void fetchActiveRoom();

    return () => {
      active = false;
    };
  }, [navigate]);

  useEffect(() => {
    let active = true;
    let isFetching = false;
    let pollTimeout: number | null = null;

    const scheduleNextPoll = () => {
      setRefreshCycle((value) => value + 1);
      pollTimeout = window.setTimeout(() => {
        void fetchRooms(false);
      }, ROOMS_POLL_INTERVAL_MS);
    };

    const fetchRooms = async (withLoadingState: boolean) => {
      if (isFetching) {
        return;
      }

      isFetching = true;

      if (withLoadingState) {
        setIsLoading(true);
      }

      try {
        const result = await listRooms({
          limit: PAGE_SIZE,
          offset,
          search: debouncedSearch,
          game: gameFilter,
        });

        if (!active) {
          return;
        }

        if (!result.ok) {
          if (result.isTokenError) {
            navigate("/login", { replace: true });
            return;
          }

          setApiError(
            result.error ?? {
              message: "Unknown error while loading rooms.",
              status: 500,
              code: "UNKNOWN",
            },
          );
          return;
        }

        setApiError(null);
        setRooms(result.data.rooms);
        setTotalCount(result.data.total_count);
      } finally {
        if (active && withLoadingState) {
          setIsLoading(false);
        }

        isFetching = false;

        if (active) {
          scheduleNextPoll();
        }
      }
    };

    void fetchRooms(true);

    return () => {
      active = false;
      if (pollTimeout !== null) {
        window.clearTimeout(pollTimeout);
      }
    };
  }, [debouncedSearch, gameFilter, manualRefreshCount, navigate, offset]);

  const currentPage = Math.floor(offset / PAGE_SIZE) + 1;
  const totalPages = Math.max(1, Math.ceil(totalCount / PAGE_SIZE));
  const canGoPrevious = offset > 0;
  const canGoNext = offset + PAGE_SIZE < totalCount;
  const quickPlayPlayersJoined = quickPlayRoom?.players.length ?? 0;
  const quickPlayCapacity = quickPlayRoom?.capacity ?? 0;

  const handleJoinRoom = (roomId: string) => {
    navigate(`/room/${roomId}`);
  };

  const handleCreateRoom = async (values: CreateRoomParams) => {
    const result = await createRoom(values);

    if (!result.ok) {
      if (result.isTokenError) {
        navigate("/login", { replace: true });
        return;
      }

      setApiError(
        result.error ?? {
          message: "Could not create room.",
          status: 500,
          code: "UNKNOWN",
        },
      );
      return;
    }

    const roomPath = `/room/${result.data.room_id}`;
    const roomUrl = new URL(roomPath, window.location.origin).toString();
    const didCopyRoomUrl = await copyTextToClipboard(roomUrl);

    setIsCreateRoomOpen(false);
    navigate(roomPath, {
      state: didCopyRoomUrl ? { roomUrlCopied: true } : undefined,
    });
  };

  const handleQuickPlay = () => {
    if (quickPlaySocketRef.current) {
      return;
    }

    setApiError(null);
    setIsQuickPlayWaiting(true);
    setQuickPlayRoom(null);
    quickPlayCloseExpectedRef.current = false;

    const socket = connectQuickPlayRoomSocket(quickPlayGame);
    quickPlaySocketRef.current = socket;

    console.log("[quick-play-ws] connecting", socket.url);

    socket.onopen = () => {
      console.log("[quick-play-ws] open");
    };

    socket.onmessage = (event) => {
      void (async () => {
        const message = await decodeRoomServerMessage(event.data);
        console.log("[quick-play-ws] message", message ?? event.data);

        if (!isMountedRef.current) {
          return;
        }

        const gameStarted = message?.roomMessage?.gameStarted;
        if (gameStarted?.gameId) {
          quickPlayCloseExpectedRef.current = true;
          navigate(`/game/${gameStarted.gameId}`);
          return;
        }

        const room = message?.roomMessage?.roomJoined?.room;
        if (room) {
          const players = room.players.map((player) => {
            return {
              id: player.playerId,
              displayName: player.displayName,
              avatarUrl: player.avatarUrl,
            };
          });

          setActiveRoom({
            roomId: room.roomId,
            displayName: room.displayName,
          });
          setQuickPlayRoom({
            roomId: room.roomId,
            displayName: room.displayName,
            capacity: room.capacity,
            players,
          });
        }

        const joinedPlayer = message?.roomMessage?.playerJoinedRoom?.player;
        if (joinedPlayer) {
          setQuickPlayRoom((currentRoom) => {
            if (!currentRoom) {
              return currentRoom;
            }

            const nextPlayer: QuickPlayPlayer = {
              id: joinedPlayer.playerId,
              displayName: joinedPlayer.displayName,
              avatarUrl: joinedPlayer.avatarUrl,
            };
            const existingIndex = currentRoom.players.findIndex((player) => {
              return player.id === nextPlayer.id;
            });

            if (existingIndex === -1) {
              return {
                ...currentRoom,
                players: [...currentRoom.players, nextPlayer],
              };
            }

            const players = [...currentRoom.players];
            players[existingIndex] = nextPlayer;
            return {
              ...currentRoom,
              players,
            };
          });
        }

        const leftPlayer = message?.roomMessage?.playerLeftRoom;
        if (leftPlayer) {
          setQuickPlayRoom((currentRoom) => {
            if (!currentRoom) {
              return currentRoom;
            }

            return {
              ...currentRoom,
              players: currentRoom.players.filter((player) => {
                return player.id !== leftPlayer.playedId;
              }),
            };
          });
        }
      })();
    };

    socket.onerror = (event) => {
      console.log("[quick-play-ws] error", event);

      if (!isMountedRef.current || quickPlayCloseExpectedRef.current) {
        return;
      }

      setApiError({
        message: "Could not start quick play.",
        status: 500,
        code: "UNKNOWN",
      });
    };

    socket.onclose = (event) => {
      console.log("[quick-play-ws] close", {
        code: event.code,
        reason: event.reason,
        wasClean: event.wasClean,
      });

      if (!isMountedRef.current) {
        return;
      }

      if (quickPlaySocketRef.current === socket) {
        // The server removes the player from the quick play room on
        // disconnect, so the local room state is stale either way.
        quickPlaySocketRef.current = null;
        setIsQuickPlayWaiting(false);
        setQuickPlayRoom(null);
        setActiveRoom(null);
      }

      if (!quickPlayCloseExpectedRef.current) {
        setApiError({
          message: "Quick play disconnected.",
          status: 500,
          code: "UNKNOWN",
        });
      }
    };
  };

  const handleLeaveQuickPlay = () => {
    quickPlayCloseExpectedRef.current = true;
    quickPlaySocketRef.current?.close();
    quickPlaySocketRef.current = null;
    setActiveRoom(null);
    setQuickPlayRoom(null);
    setIsQuickPlayWaiting(false);
  };

  const handleLeaveActiveRoom = async () => {
    const result = await leaveRoom();

    if (!result.ok) {
      if (result.isTokenError) {
        navigate("/login", { replace: true });
        return;
      }

      setApiError(
        result.error ?? {
          message: "Could not leave room.",
          status: 500,
          code: "UNKNOWN",
        },
      );
      return;
    }

    setActiveRoom(null);
    setQuickPlayRoom(null);
    quickPlayCloseExpectedRef.current = true;
    quickPlaySocketRef.current?.close();
    quickPlaySocketRef.current = null;
    setIsQuickPlayWaiting(false);
  };

  return (
    <main className="page">
      {activeRoom ? (
        <div className="reconnect-banner-wrap" role="status" aria-live="polite">
          <div className="reconnect-banner">
            <span>
              {isQuickPlayWaiting ? (
                <>
                  Finding a Quick Play match in{" "}
                  <strong>{activeRoom.displayName}</strong>.
                </>
              ) : (
                <>
                  You are still in <strong>{activeRoom.displayName}</strong>.
                </>
              )}
            </span>

            <div className="reconnect-banner-actions">
              {isQuickPlayWaiting ? null : (
                <Button
                  size="sm"
                  onClick={() => {
                    navigate(`/room/${activeRoom.roomId}`);
                  }}
                >
                  Reconnect
                </Button>
              )}
              <Button
                size="sm"
                variant="outline"
                onClick={() => {
                  if (isQuickPlayWaiting) {
                    handleLeaveQuickPlay();
                    return;
                  }

                  void handleLeaveActiveRoom();
                }}
              >
                Leave room
              </Button>
            </div>
          </div>
        </div>
      ) : null}

      <section className="page-header"></section>

      <Card
        className={
          activeRoom ? "lobby-table-card is-locked" : "lobby-table-card"
        }
        aria-disabled={activeRoom ? true : undefined}
      >
        <CardHeader className="table-card-header">
          <CardTitle>Active Rooms</CardTitle>
          <div className="lobby-header-actions">
            <select
              className="field-input"
              aria-label="Quick Play game"
              value={quickPlayGame}
              disabled={Boolean(activeRoom) || isQuickPlayWaiting}
              onChange={(event) => {
                const nextGame = parseGame(event.target.value);
                if (nextGame) {
                  setQuickPlayGame(nextGame);
                }
              }}
            >
              {gameFilterOptions.map((option) => {
                return (
                  <option key={option.value} value={option.value}>
                    {option.label}
                  </option>
                );
              })}
            </select>
            <Button
              size="sm"
              variant="outline"
              onClick={handleQuickPlay}
              disabled={Boolean(activeRoom) || isQuickPlayWaiting}
            >
              {isQuickPlayWaiting ? "Finding game..." : "Quick Play"}
            </Button>
            <Button
              size="sm"
              onClick={() => setIsCreateRoomOpen(true)}
              disabled={Boolean(activeRoom) || isQuickPlayWaiting}
            >
              Create room
            </Button>
          </div>
        </CardHeader>
        <CardContent>
          {activeRoom ? (
            <p className="lobby-table-lock-note">
              {isQuickPlayWaiting
                ? "Waiting for enough Quick Play players to start the game."
                : "You are already in a room. Reconnect above to continue."}
            </p>
          ) : null}

          <div className="table-filters" role="search">
            <input
              type="search"
              className="field-input"
              placeholder="Search rooms"
              value={searchInput}
              onChange={(event) => setSearchInput(event.target.value)}
              aria-label="Search rooms"
              autoComplete="off"
            />
            <select
              className="field-input"
              aria-label="Filter by game"
              value={gameFilter ?? ""}
              onChange={(event) => {
                const nextValue = event.target.value;
                setGameFilter(nextValue ? parseGame(nextValue) : null);
              }}
            >
              <option value="">All games</option>
              {gameFilterOptions.map((option) => {
                return (
                  <option key={option.value} value={option.value}>
                    {option.label}
                  </option>
                );
              })}
            </select>
          </div>

          <div
            className="lobby-refresh-indicator"
            role="status"
            aria-live="polite"
          >
            <div className="lobby-refresh-row">
              <div className="lobby-refresh-track" aria-hidden="true">
                <span key={refreshCycle} className="lobby-refresh-bar" />
              </div>
              <Button
                size="sm"
                variant="outline"
                onClick={() => setManualRefreshCount((value) => value + 1)}
              >
                Refresh
              </Button>
            </div>
          </div>

          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Room</TableHead>
                <TableHead>Game</TableHead>
                <TableHead>Host</TableHead>
                <TableHead>Players</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {rooms.map((room) => {
                const host = room.players[0];

                return (
                  <TableRow
                    key={room.room_id}
                    className={
                      activeRoom ? undefined : "ui-table__row--clickable"
                    }
                    onClick={() => {
                      if (activeRoom) {
                        return;
                      }

                      handleJoinRoom(room.room_id);
                    }}
                  >
                    <TableCell>{room.display_name}</TableCell>
                    <TableCell>{getGameDisplayName(room.game)}</TableCell>
                    <TableCell>
                      {host ? (
                        <div className="host-cell">
                          <img
                            src={host.image_url}
                            alt={host.display_name}
                            className="host-avatar"
                            loading="lazy"
                            referrerPolicy="no-referrer"
                          />
                          <span>{host.display_name}</span>
                        </div>
                      ) : (
                        <span>-</span>
                      )}
                    </TableCell>
                    <TableCell>
                      {room.occupied}/{room.capacity}
                    </TableCell>
                  </TableRow>
                );
              })}
              {!isLoading && rooms.length === 0 ? (
                <TableRow>
                  <TableCell>No rooms found</TableCell>
                  <TableCell>-</TableCell>
                  <TableCell>-</TableCell>
                  <TableCell>-</TableCell>
                </TableRow>
              ) : null}
            </TableBody>
          </Table>

          <div className="table-toolbar">
            <div className="table-pagination">
              <Button
                size="sm"
                variant="outline"
                disabled={Boolean(activeRoom) || isLoading || !canGoPrevious}
                onClick={() =>
                  setOffset((value) => Math.max(0, value - PAGE_SIZE))
                }
              >
                Previous
              </Button>
              <span className="table-page-label">
                Page {currentPage} of {totalPages}
              </span>
              <Button
                size="sm"
                variant="outline"
                disabled={Boolean(activeRoom) || isLoading || !canGoNext}
                onClick={() => setOffset((value) => value + PAGE_SIZE)}
              >
                Next
              </Button>
            </div>
          </div>
        </CardContent>
      </Card>

      {apiError ? (
        <ErrorModal error={apiError} onClose={() => setApiError(null)} />
      ) : null}

      {isCreateRoomOpen ? (
        <CreateRoomModal
          onClose={() => setIsCreateRoomOpen(false)}
          onCreate={handleCreateRoom}
        />
      ) : null}

      {isQuickPlayWaiting && quickPlayRoom ? (
        <div
          className="overlay-backdrop"
          onClick={handleLeaveQuickPlay}
          role="presentation"
        >
          <div
            className="overlay-card quick-play-modal"
            onClick={(event) => event.stopPropagation()}
            role="dialog"
            aria-modal="true"
            aria-labelledby="quick-play-title"
          >
            <p className="eyebrow">Quick Play</p>
            <h2 id="quick-play-title" className="overlay-title">
              Finding players
              <span className="quick-play-ellipsis">...</span>
            </h2>
            <p className="quick-play-modal-message">
              {quickPlayPlayersJoined}/{quickPlayCapacity} players joined
            </p>

            <div className="quick-play-player-meter" aria-hidden="true">
              {Array.from({ length: quickPlayCapacity }, (_, index) => {
                return (
                  <span
                    key={index}
                    className={
                      index < quickPlayPlayersJoined
                        ? "quick-play-player-dot is-filled"
                        : "quick-play-player-dot"
                    }
                  />
                );
              })}
            </div>

            <ul className="quick-play-player-list">
              {quickPlayRoom.players.map((player) => {
                return (
                  <li key={player.id} className="quick-play-player-item">
                    <img
                      src={player.avatarUrl}
                      alt={player.displayName}
                      className="quick-play-player-avatar"
                      loading="lazy"
                      referrerPolicy="no-referrer"
                    />
                    <span>{player.displayName}</span>
                  </li>
                );
              })}
            </ul>

            <div className="overlay-actions">
              <Button
                variant="outline"
                onClick={handleLeaveQuickPlay}
              >
                Leave Quick Play
              </Button>
            </div>
          </div>
        </div>
      ) : null}
    </main>
  );
};

export default LobbyPage;
