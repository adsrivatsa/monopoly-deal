// Phase 4a scaffold page for "Monopoly Deal: No Mercy".
//
// This page owns the game socket lifecycle and the decoded game state, then
// hands public state to DealNoMercyGameMount for display. It is intentionally
// minimal: it connects, tracks the latest GameState snapshot, shows chat and
// error toasts, and never crashes on an unrecognized server message (it logs
// unhandled ones instead). The classic MonopolyDealGamePage is the reference
// for the far richer per-action reducer Phase 4b will need.
//
// Phase 4b: this page should grow an action-history sidebar, per-action
// optimistic state updates, demand overlays, a game-end overlay, reconnect
// deadline handling, and it should wire the send-* functions from
// api/dealNoMercySocket.ts to interactive board callbacks.
import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { useParams } from "react-router-dom";
import {
  connectDealNoMercyGameSocket,
  decodeDealNoMercyServerMessage,
  sendDealNoMercyChatMessage,
  toDealNoMercyServerMessageJson,
} from "../../../api/dealNoMercySocket";
import { getPlayer } from "../../../api/player";
import Button from "../../../components/ui/button";
import ErrorToastStack, {
  type ErrorToastNotice,
} from "../../../components/ui/error-toast-stack";
import {
  AssetImageKind,
  type AssetImage,
  type GameState,
} from "../../../generated/deal_no_mercy";
import {
  playChatTock,
  playLossTocks,
  playWinTocks,
  playYourTurnTocks,
} from "../../shared/sound";
import DealNoMercyGameMount from "../DealNoMercyGameMount";

type ChatMessage = {
  id: string;
  playerId: string;
  text: string;
};

// Builds the assetKey -> large-image-URL lookup the mount renders cards with.
const toLargeAssetImageByKey = (
  assetImages: AssetImage[],
): Record<number, string> => {
  const lookup: Record<number, string> = {};
  for (const assetImage of assetImages) {
    if (
      assetImage.imageUrl &&
      assetImage.kind === AssetImageKind.ASSET_IMAGE_KIND_LARGE
    ) {
      lookup[assetImage.assetKey] = assetImage.imageUrl;
    }
  }
  return lookup;
};

const DealNoMercyGamePage = () => {
  const { game_id: gameId } = useParams();
  const socketRef = useRef<WebSocket | null>(null);
  const reconnectTimeoutRef = useRef<number | null>(null);
  const reconnectAttemptRef = useRef(0);
  const selfPlayerIdRef = useRef<string | null>(null);
  const currentPlayerIdRef = useRef<string | null>(null);

  const [gameState, setGameState] = useState<GameState | null>(null);
  const [assetImageByKey, setAssetImageByKey] = useState<
    Record<number, string>
  >({});
  const [selfPlayerId, setSelfPlayerId] = useState<string | null>(null);
  const [playerNameById, setPlayerNameById] = useState<Record<string, string>>(
    {},
  );
  const [chatMessages, setChatMessages] = useState<ChatMessage[]>([]);
  const [chatInput, setChatInput] = useState("");
  const [errorNotices, setErrorNotices] = useState<ErrorToastNotice[]>([]);

  const pushErrorNotice = useCallback(
    (message: string, code: string, eyebrow = "Game server") => {
      setErrorNotices((current) => {
        const notice: ErrorToastNotice = {
          id: `error-${Date.now()}-${Math.random().toString(16).slice(2)}`,
          message,
          code,
          eyebrow,
        };
        return [...current.slice(-4), notice];
      });
    },
    [],
  );

  const dismissErrorNotice = useCallback((id: string) => {
    setErrorNotices((current) => current.filter((notice) => notice.id !== id));
  }, []);

  // Resolve the current player's id so we can label "You" and pick turn/win
  // sounds.
  useEffect(() => {
    let active = true;
    void (async () => {
      try {
        const response = await getPlayer();
        if (!active || !response.ok) {
          if (active && !response.ok) {
            pushErrorNotice(
              "Could not load current player details.",
              "PLAYER_LOOKUP_FAILED",
              "Player error",
            );
          }
          return;
        }
        const data = (await response.json()) as { player_id: string | null };
        if (active && data.player_id) {
          selfPlayerIdRef.current = data.player_id;
          setSelfPlayerId(data.player_id);
        }
      } catch {
        if (active) {
          pushErrorNotice(
            "Could not load current player details.",
            "PLAYER_LOOKUP_CRASH",
            "Player error",
          );
        }
      }
    })();
    return () => {
      active = false;
    };
  }, [pushErrorNotice]);

  useEffect(() => {
    let didUnmount = false;

    const clearReconnectTimer = () => {
      if (reconnectTimeoutRef.current !== null) {
        window.clearTimeout(reconnectTimeoutRef.current);
        reconnectTimeoutRef.current = null;
      }
    };

    const scheduleReconnect = () => {
      if (didUnmount) {
        return;
      }
      clearReconnectTimer();
      const attempt = reconnectAttemptRef.current;
      const delayMs = Math.min(5000, 500 * 2 ** attempt);
      reconnectAttemptRef.current = attempt + 1;
      reconnectTimeoutRef.current = window.setTimeout(() => {
        reconnectTimeoutRef.current = null;
        connectSocket();
      }, delayMs);
    };

    // handleServerMessage applies one decoded ServerMessage. Every branch is
    // defensive: an unrecognized or empty payload is logged, never thrown, so
    // the shell survives any message Phase 4b has not taught it to handle yet.
    const handleServerMessage = (json: unknown) => {
      const message = json as {
        dealNoMercyMessage?: {
          error?: { message?: string; code?: string };
          chatReceived?: { playerId?: string; payload?: string };
          gameState?: GameState;
          wonGame?: { playerId?: string };
          action?: { playerId?: string; actionStartTurn?: unknown };
          actionHistory?: unknown;
        };
      };

      const payload = message.dealNoMercyMessage;
      if (!payload) {
        console.warn("[no-mercy-ws] unhandled message", json);
        return;
      }

      if (payload.error) {
        pushErrorNotice(
          payload.error.message ?? "Server error.",
          payload.error.code ?? "SERVER_ERROR",
        );
      }

      if (payload.gameState) {
        const nextState = payload.gameState;
        const nextAssetImages = nextState.assetImages ?? [];
        if (nextAssetImages.length > 0) {
          setAssetImageByKey((current) => ({
            ...current,
            ...toLargeAssetImageByKey(nextAssetImages),
          }));
        }
        currentPlayerIdRef.current = nextState.currentPlayerId;
        setPlayerNameById((current) => {
          const incoming = Object.fromEntries(
            nextState.players.map((player) => [
              player.playerId,
              player.displayName,
            ]),
          );
          return { ...current, ...incoming };
        });
        setGameState(nextState);
      }

      if (payload.chatReceived) {
        const chat = payload.chatReceived;
        setChatMessages((current) => [
          ...current,
          {
            id: `chat-${chat.playerId ?? "?"}-${Date.now()}-${current.length}`,
            playerId: chat.playerId ?? "",
            text: chat.payload ?? "",
          },
        ]);
        if (chat.playerId && chat.playerId !== selfPlayerIdRef.current) {
          playChatTock();
        }
      }

      if (payload.wonGame) {
        if (payload.wonGame.playerId === selfPlayerIdRef.current) {
          playWinTocks();
        } else {
          playLossTocks();
        }
      }

      // Phase 4b: the classic page fans an Action out into optimistic state
      // mutations (money played, properties moved, demands, distributions).
      // For now we only detect a start-of-turn to sound the turn cue; the next
      // authoritative gameState snapshot carries the real state.
      const action = payload.action;
      if (
        action?.actionStartTurn &&
        action.playerId &&
        action.playerId === selfPlayerIdRef.current
      ) {
        playYourTurnTocks();
      }

      // Anything with no recognized field is logged so Phase 4b can see gaps.
      if (
        !payload.error &&
        !payload.gameState &&
        !payload.chatReceived &&
        !payload.wonGame &&
        !payload.action &&
        !payload.actionHistory
      ) {
        console.warn("[no-mercy-ws] unhandled dealNoMercyMessage", payload);
      }
    };

    const connectSocket = () => {
      if (didUnmount) {
        return;
      }

      const socket = connectDealNoMercyGameSocket();
      socketRef.current = socket;
      console.log("[no-mercy-ws] connecting", socket.url);

      socket.onopen = () => {
        reconnectAttemptRef.current = 0;
        clearReconnectTimer();
        console.log("[no-mercy-ws] open", { gameId });
      };

      socket.onmessage = (event) => {
        void (async () => {
          try {
            const decoded = await decodeDealNoMercyServerMessage(event.data);
            if (!decoded) {
              console.warn("[no-mercy-ws] non-binary message", event.data);
              return;
            }
            handleServerMessage(toDealNoMercyServerMessageJson(decoded));
          } catch (error) {
            console.warn("[no-mercy-ws] failed to handle message", error);
          }
        })();
      };

      socket.onerror = (event) => {
        console.warn("[no-mercy-ws] error", event);
      };

      socket.onclose = (event) => {
        console.log("[no-mercy-ws] close", {
          code: event.code,
          reason: event.reason,
          wasClean: event.wasClean,
        });
        if (socketRef.current === socket) {
          socketRef.current = null;
        }
        scheduleReconnect();
      };
    };

    connectSocket();

    return () => {
      didUnmount = true;
      clearReconnectTimer();
      const socket = socketRef.current;
      socketRef.current = null;
      if (socket) {
        socket.onopen = null;
        socket.onmessage = null;
        socket.onerror = null;
        socket.onclose = null;
        socket.close();
      }
    };
  }, [gameId, pushErrorNotice]);

  const handleSendChat = useCallback(() => {
    const trimmed = chatInput.trim();
    const socket = socketRef.current;
    if (!trimmed || !socket || socket.readyState !== WebSocket.OPEN) {
      return;
    }
    sendDealNoMercyChatMessage(socket, trimmed);
    setChatInput("");
  }, [chatInput]);

  const nameFor = useCallback(
    (playerId: string): string => {
      if (playerId === selfPlayerId) {
        return "You";
      }
      return playerNameById[playerId] ?? playerId.slice(0, 8);
    },
    [playerNameById, selfPlayerId],
  );

  const orderedChat = useMemo(() => chatMessages.slice(-50), [chatMessages]);

  return (
    <main className="page deal-no-mercy-page">
      <header className="deal-no-mercy-page__header">
        <h1>Deal No Mercy</h1>
        <p className="deal-no-mercy-page__subtitle">
          {gameState
            ? `Turn: ${nameFor(gameState.currentPlayerId)} · moves left ${gameState.movesLeft}`
            : "Connecting..."}
        </p>
      </header>

      <div className="deal-no-mercy-page__body">
        <DealNoMercyGameMount
          gameState={gameState}
          assetImageByKey={assetImageByKey}
          selfPlayerId={selfPlayerId}
          playerNameById={playerNameById}
        />

        {/* Phase 4b: replace this bare chat box with the full sidebar (chat +
            action history) the classic game uses. */}
        <aside className="deal-no-mercy-page__chat">
          <h2 className="deal-no-mercy-shell__heading">Chat</h2>
          <div className="deal-no-mercy-page__chat-log">
            {orderedChat.map((chat) => (
              <p key={chat.id} className="deal-no-mercy-page__chat-line">
                <strong>{nameFor(chat.playerId)}:</strong> {chat.text}
              </p>
            ))}
            {orderedChat.length === 0 ? (
              <p className="deal-no-mercy-shell__empty">No messages yet.</p>
            ) : null}
          </div>
          <form
            className="deal-no-mercy-page__chat-form"
            onSubmit={(event) => {
              event.preventDefault();
              handleSendChat();
            }}
          >
            <input
              type="text"
              className="field-input"
              placeholder="Say something"
              value={chatInput}
              onChange={(event) => setChatInput(event.target.value)}
              aria-label="Chat message"
            />
            <Button type="submit" size="sm">
              Send
            </Button>
          </form>
        </aside>
      </div>

      <ErrorToastStack notices={errorNotices} onDismiss={dismissErrorNotice} />
    </main>
  );
};

export default DealNoMercyGamePage;
