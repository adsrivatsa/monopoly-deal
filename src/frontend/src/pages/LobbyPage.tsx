import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import {
  createRoom,
  CreateRoomParams,
  getRoom,
  joinRoom,
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

const PAGE_SIZE = 10;
const SEARCH_DEBOUNCE_MS = 350;

const gameFilterOptions: Array<{ value: Game; label: string }> =
  supportedGames.map((game) => {
    return {
      value: game,
      label: getGameDisplayName(game),
    };
  });

const LobbyPage = () => {
  const navigate = useNavigate();
  const [rooms, setRooms] = useState<RoomListItem[]>([]);
  const [totalCount, setTotalCount] = useState(0);
  const [offset, setOffset] = useState(0);
  const [searchInput, setSearchInput] = useState("");
  const [debouncedSearch, setDebouncedSearch] = useState<string | null>(null);
  const [gameFilter, setGameFilter] = useState<Game | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [apiError, setApiError] = useState<ApiErrorPayload | null>(null);
  const [isCreateRoomOpen, setIsCreateRoomOpen] = useState(false);
  const [activeRoom, setActiveRoom] = useState<{
    roomId: string;
    displayName: string;
  } | null>(null);

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

    const fetchRooms = async () => {
      setIsLoading(true);

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
        if (active) {
          setIsLoading(false);
        }
      }
    };

    void fetchRooms();

    return () => {
      active = false;
    };
  }, [debouncedSearch, gameFilter, navigate, offset]);

  const currentPage = Math.floor(offset / PAGE_SIZE) + 1;
  const totalPages = Math.max(1, Math.ceil(totalCount / PAGE_SIZE));
  const canGoPrevious = offset > 0;
  const canGoNext = offset + PAGE_SIZE < totalCount;

  const handleJoinRoom = async (roomId: string) => {
    const result = await joinRoom(roomId);

    if (!result.ok) {
      if (result.isTokenError) {
        navigate("/login", { replace: true });
        return;
      }

      setApiError(
        result.error ?? {
          message: "Could not join room.",
          status: 500,
          code: "UNKNOWN",
        },
      );
      return;
    }

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

    setIsCreateRoomOpen(false);
    navigate(`/room/${result.data.room_id}`);
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
  };

  return (
    <main className="page">
      {activeRoom ? (
        <div className="reconnect-banner-wrap" role="status" aria-live="polite">
          <div className="reconnect-banner">
            <span>
              You are still in <strong>{activeRoom.displayName}</strong>.
            </span>

            <div className="reconnect-banner-actions">
              <Button
                size="sm"
                onClick={() => {
                  navigate(`/room/${activeRoom.roomId}`);
                }}
              >
                Reconnect
              </Button>
              <Button size="sm" variant="outline" onClick={() => {
                void handleLeaveActiveRoom();
              }}>
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
          <Button
            size="sm"
            onClick={() => setIsCreateRoomOpen(true)}
            disabled={Boolean(activeRoom)}
          >
            Create room
          </Button>
        </CardHeader>
        <CardContent>
          {activeRoom ? (
            <p className="lobby-table-lock-note">
              You are already in a room. Reconnect above to continue.
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

                      void handleJoinRoom(room.room_id);
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
    </main>
  );
};

export default LobbyPage;
