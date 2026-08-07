// Phase 4b game page for "Monopoly Deal: No Mercy".
//
// This page owns the game socket lifecycle, the decoded + optimistically-updated
// GameState, the action-history log, timers, sounds, and the wiring of every
// send-* function to a board callback. It mirrors MonopolyDealGamePage but is
// factored into smaller pieces: the optimistic action fan-out lives in
// dealNoMercyReducer.ts, the sidebar and game-end overlay are their own
// components.
//
// Protocol reminder (service/deal-no-mercy): on connect/reconnect the server
// sends one authoritative GameState snapshot plus a replay of ActionHistory
// (masked, silent). Each subsequent play broadcasts a single Action, which this
// page fans out into local GameState mutations. wonGame arrives as its own
// message. The board renders purely from the GameState we hold here.
import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import {
  connectDealNoMercyGameSocket,
  decodeDealNoMercyServerMessage,
  sendDealNoMercyChatMessage,
  sendDealNoMercyCompleteTurnMessage,
  sendDealNoMercyComplyBankCardDemandMessage,
  sendDealNoMercyComplyBankSwapDemandMessage,
  sendDealNoMercyComplyColorPropertiesDemandMessage,
  sendDealNoMercyComplyDebtTrapDemandMessage,
  sendDealNoMercyComplyPaymentDemandMessage,
  sendDealNoMercyComplyPickpocketDemandMessage,
  sendDealNoMercyComplyPropertyDemandMessage,
  sendDealNoMercyComplyPropertySetDemandMessage,
  sendDealNoMercyComplyRepoManDemandMessage,
  sendDealNoMercyComplyTaxDayDemandMessage,
  sendDealNoMercyDenyDemandMessage,
  sendDealNoMercyDiscardCardsMessage,
  sendDealNoMercyPlayBankSwapMessage,
  sendDealNoMercyPlayBigPaydayMessage,
  sendDealNoMercyPlayDebtTrapMessage,
  sendDealNoMercyPlayGoAgainMessage,
  sendDealNoMercyPlayHeistMessage,
  sendDealNoMercyPlayMarketCrashMessage,
  sendDealNoMercyPlayMoneyMessage,
  sendDealNoMercyPlayPickpocketMessage,
  sendDealNoMercyPlayPropertyMessage,
  sendDealNoMercyPlayPropertyRaidMessage,
  sendDealNoMercyPlayRentMessage,
  sendDealNoMercyPlayRepoManMessage,
  sendDealNoMercyPlaySetSnatcherMessage,
  sendDealNoMercyPlayShackMessage,
  sendDealNoMercyPlayTaxDayMessage,
  sendDealNoMercyPlayYoinkMessage,
  sendDealNoMercyRearrangeCardMessage,
  sendDealNoMercySettleDebtMessage,
} from "../../../api/dealNoMercySocket";
import { getPlayer } from "../../../api/player";
import ErrorToastStack, {
  type ErrorToastNotice,
} from "../../../components/ui/error-toast-stack";
import {
  ActionKind,
  AssetImageKind,
  StealCategory,
  actionKindToJSON,
  type Action,
  type AssetImage,
  type Color,
  type DistributionAssignment,
  type GameState,
  type Player,
  type PropertyPick,
  type ServerMessage,
} from "../../../generated/deal_no_mercy";
import {
  installAudioUnlock,
  playCardTock,
  playChatTock,
  playDemandKnock,
  playLossTocks,
  playWinTocks,
  playYourTurnTocks,
} from "../../shared/sound";
import DealNoMercyGameMount, {
  type DealNoMercyBoardViewMode,
} from "../DealNoMercyGameMount";
import DealNoMercyGameEndOverlay from "./DealNoMercyGameEndOverlay";
import DealNoMercyGameSidebar from "./DealNoMercyGameSidebar";
import {
  applyActionToGameState,
  setReducerSelfPlayerId,
} from "./dealNoMercyReducer";
import type {
  ActionHistoryEntry,
  GameChatMessage,
  WonGameResult,
} from "./dealNoMercyPageTypes";

// Builds the assetKey -> large-image-URL lookup the board renders cards with.
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

const VIEW_MODE_STORAGE_KEY = "deal-no-mercy:view-mode";

const getInitialBoardViewMode = (): DealNoMercyBoardViewMode => {
  try {
    const stored = window.localStorage.getItem(VIEW_MODE_STORAGE_KEY);
    if (stored === "compact" || stored === "expanded") {
      return stored;
    }
  } catch {
    // Ignore storage access errors (private mode, etc.).
  }
  return "expanded";
};

const randomId = (prefix: string): string =>
  `${prefix}-${Date.now()}-${Math.random().toString(16).slice(2)}`;

// ---------------------------------------------------------------------------
// Action-history entry construction
// ---------------------------------------------------------------------------

const stealCategoryLabel = (category: StealCategory): string => {
  switch (category) {
    case StealCategory.STEAL_CATEGORY_MONEY:
      return "money";
    case StealCategory.STEAL_CATEGORY_ACTION:
      return "action";
    case StealCategory.STEAL_CATEGORY_PROPERTY:
      return "property";
    default:
      return "cards";
  }
};

// buildHistoryEntry converts one Action into a display entry for the sidebar
// log, or null when the action carries nothing worth logging (e.g. an empty
// timeout auto-discard). nameOf resolves a player id to a display label.
const buildHistoryEntry = (
  action: Action,
  nameOf: (playerId: string) => string,
): ActionHistoryEntry | null => {
  const playerId = action.playerId;
  const actor = playerId ? nameOf(playerId) : "Someone";
  const id = `${action.seqNum}-${actionKindToJSON(action.kind)}-${randomId("h")}`;

  switch (action.kind) {
    case ActionKind.ACTION_KIND_START_TURN: {
      const full = action.actionStartTurn;
      const masked = action.maskedActionStartTurn;
      const drawCount = full?.cards.length ?? masked?.numCards ?? 0;
      return {
        id,
        kind: full ? "startTurn" : "maskedStartTurn",
        text: `${actor} started their turn (drew ${drawCount})`,
        playerId,
        drawCount,
      };
    }
    case ActionKind.ACTION_KIND_PLAY_MONEY:
      return {
        id,
        kind: "playMoney",
        text: `${actor} banked money`,
        playerId,
        cardAssetKey: action.actionPlayMoney?.card?.assetKey,
      };
    case ActionKind.ACTION_KIND_PLAY_PROPERTY:
      return {
        id,
        kind: "playProperty",
        text: `${actor} played a property`,
        playerId,
        cardAssetKeys: action.actionPlayProperty?.propertySet?.cards.map(
          (card) => card.assetKey,
        ),
      };
    case ActionKind.ACTION_KIND_PLAY_SHACK:
      return {
        id,
        kind: "playShack",
        text: `${actor} added a Shack`,
        playerId,
        cardAssetKey: action.actionPlayShack?.card?.assetKey,
      };
    case ActionKind.ACTION_KIND_PLAY_BIG_PAYDAY: {
      const drawCount =
        action.actionPlayBigPayday?.cards.length ??
        action.maskedActionPlayBigPayday?.numCards ??
        0;
      return {
        id,
        kind: "bigPayday",
        text: `${actor} played Big Payday (drew ${drawCount})`,
        playerId,
        cardAssetKey:
          action.actionPlayBigPayday?.lastPlayedCard?.assetKey ??
          action.maskedActionPlayBigPayday?.lastPlayedCard?.assetKey,
        drawCount,
      };
    }
    case ActionKind.ACTION_KIND_PLAY_GO_AGAIN:
      return {
        id,
        kind: "goAgain",
        text: `${actor} played Go Again`,
        playerId,
        cardAssetKey: action.actionPlayGoAgain?.lastPlayedCard?.assetKey,
      };
    case ActionKind.ACTION_KIND_REARRANGE_CARD:
      return {
        id,
        kind: "generic",
        text: `${actor} rearranged a property`,
        playerId,
      };
    case ActionKind.ACTION_KIND_DEMAND_CREATED: {
      const created = action.actionDemandsCreated;
      if (created?.deniedDemand) {
        return {
          id,
          kind: "denied",
          text: `${actor} played NAH! to deny a demand`,
          playerId,
        };
      }
      return {
        id,
        kind: "demandsCreated",
        text: `${actor} played ${created?.demands.length ?? 0 > 1 ? "cards demanding a response" : "a card demanding a response"}`,
        playerId,
        cardAssetKey: created?.lastPlayedCard?.assetKey,
      };
    }
    case ActionKind.ACTION_KIND_DEMAND_COMPLIED: {
      const complied = action.actionDemandComplied;
      if (complied?.transferCards) {
        const tc = complied.transferCards;
        const keys = [
          ...tc.cards.map((card) => card.assetKey),
          ...tc.propertySets.flatMap((set) => set.cards.map((card) => card.assetKey)),
        ];
        return {
          id,
          kind: "paymentComplied",
          text: `${nameOf(tc.sourceId)} paid ${nameOf(tc.targetId)}${tc.debt ? " (debt issued)" : ""}`,
          playerId,
          cardAssetKeys: keys,
          sourcePlayerId: tc.sourceId,
          targetPlayerId: tc.targetId,
        };
      }
      if (complied?.transferProperty) {
        const tp = complied.transferProperty;
        return {
          id,
          kind: "propertySteal",
          text: `${nameOf(tp.targetId)} took a property from ${nameOf(tp.sourceId)}`,
          playerId,
          cardAssetKeys: tp.card ? [tp.card.assetKey] : [],
          sourcePlayerId: tp.sourceId,
          targetPlayerId: tp.targetId,
        };
      }
      if (complied?.transferPropertySets) {
        const ts = complied.transferPropertySets;
        return {
          id,
          kind: "propertySetSteal",
          text: `${nameOf(ts.targetId)} took a property set from ${nameOf(ts.sourceId)}`,
          playerId,
          cardAssetKeys: ts.propertySets.flatMap((set) =>
            set.cards.map((card) => card.assetKey),
          ),
          sourcePlayerId: ts.sourceId,
          targetPlayerId: ts.targetId,
        };
      }
      if (complied?.transferBankSwap) {
        const tb = complied.transferBankSwap;
        return {
          id,
          kind: "bankSwap",
          text: `${nameOf(tb.sourceId)} swapped banks with ${nameOf(tb.targetId)}`,
          playerId,
          sourcePlayerId: tb.sourceId,
          targetPlayerId: tb.targetId,
        };
      }
      if (complied?.transferDistribution) {
        const td = complied.transferDistribution;
        return {
          id,
          kind: "distribution",
          text: `${nameOf(td.sourceId)} distributed cards`,
          playerId,
          cardAssetKeys: td.entries
            .map((entry) => entry.card?.assetKey)
            .filter((key): key is number => typeof key === "number"),
          sourcePlayerId: td.sourceId,
        };
      }
      if (complied?.transferPickpocket) {
        const tp = complied.transferPickpocket;
        return {
          id,
          kind: "pickpocket",
          text: `${nameOf(tp.targetId)} pickpocketed ${stealCategoryLabel(tp.category)} from ${nameOf(tp.sourceId)}`,
          playerId,
          cardAssetKeys: tp.cards.map((card) => card.assetKey),
          sourcePlayerId: tp.sourceId,
          targetPlayerId: tp.targetId,
        };
      }
      if (complied?.maskedTransferPickpocket) {
        const mp = complied.maskedTransferPickpocket;
        return {
          id,
          kind: "pickpocket",
          text: `${nameOf(mp.targetId)} pickpocketed ${mp.numCards} ${stealCategoryLabel(mp.category)} card(s) from ${nameOf(mp.sourceId)}`,
          playerId,
          drawCount: mp.numCards,
          sourcePlayerId: mp.sourceId,
          targetPlayerId: mp.targetId,
        };
      }
      if (complied?.transferDebtChip?.debt) {
        const debt = complied.transferDebtChip.debt;
        return {
          id,
          kind: "debtTrap",
          text: `${nameOf(debt.debtorId)} was hit by a Debt Trap`,
          playerId,
          sourcePlayerId: debt.debtorId,
          targetPlayerId: debt.creditorId,
        };
      }
      return {
        id,
        kind: "generic",
        text: `${actor} resolved a demand`,
        playerId,
      };
    }
    case ActionKind.ACTION_KIND_DEBT_SETTLED: {
      const settled = action.actionDebtSettled;
      return {
        id,
        kind: "debtSettled",
        text: settled?.forgiven
          ? `${actor} had a debt forgiven (empty hand)`
          : `${actor} settled a debt`,
        playerId,
        cardAssetKey: settled?.card?.assetKey,
      };
    }
    case ActionKind.ACTION_KIND_DISCARD_CARDS: {
      const discarded = action.actionDiscardCards?.cards;
      const maskedCount = action.maskedActionDiscardCards?.numCards;
      if ((discarded?.length ?? 0) === 0 && (maskedCount ?? 0) === 0) {
        // Empty auto-discard on timeout — nothing worth logging.
        return null;
      }
      if (discarded && discarded.length > 0) {
        return {
          id,
          kind: "discardCards",
          text: `${actor} discarded ${discarded.length} card(s)`,
          playerId,
          cardAssetKeys: discarded.map((card) => card.assetKey),
        };
      }
      return {
        id,
        kind: "maskedDiscardCards",
        text: `${actor} discarded ${maskedCount ?? 0} card(s)`,
        playerId,
        drawCount: maskedCount ?? 0,
      };
    }
    default:
      return null;
  }
};

const DealNoMercyGamePage = () => {
  const { game_id: gameId } = useParams();
  const navigate = useNavigate();

  const socketRef = useRef<WebSocket | null>(null);
  const reconnectTimeoutRef = useRef<number | null>(null);
  const reconnectAttemptRef = useRef(0);
  const selfPlayerIdRef = useRef<string | null>(null);
  const currentTurnPlayerIdRef = useRef<string | null>(null);
  const actionHistoryListRef = useRef<HTMLDivElement | null>(null);
  // Track demands already knocked-for so we don't re-knock on re-render.
  const knockedDemandIdsRef = useRef<Set<string>>(new Set());

  const [gameState, setGameState] = useState<GameState | null>(null);
  const [assetImageByKey, setAssetImageByKey] = useState<Record<number, string>>(
    {},
  );
  const [selfPlayerId, setSelfPlayerId] = useState<string | null>(null);
  const [playerNameById, setPlayerNameById] = useState<Record<string, string>>(
    {},
  );
  const [chatMessages, setChatMessages] = useState<GameChatMessage[]>([]);
  const [actionHistoryEntries, setActionHistoryEntries] = useState<
    ActionHistoryEntry[]
  >([]);
  const [isSidebarCollapsed, setIsSidebarCollapsed] = useState(false);
  const [errorNotices, setErrorNotices] = useState<ErrorToastNotice[]>([]);
  const [selectedDiscardCardIds, setSelectedDiscardCardIds] = useState<
    Set<string>
  >(() => new Set());
  const [wonGame, setWonGame] = useState<WonGameResult | null>(null);
  const [demandDeadlineMsById, setDemandDeadlineMsById] = useState<
    Record<string, number>
  >({});
  const [turnDeadlineMs, setTurnDeadlineMs] = useState(0);
  const [boardViewMode, setBoardViewMode] = useState<DealNoMercyBoardViewMode>(
    getInitialBoardViewMode,
  );

  const pushErrorNotice = useCallback(
    (message: string, code: string, eyebrow = "Game server") => {
      setErrorNotices((current) => {
        const notice: ErrorToastNotice = {
          id: randomId("error"),
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

  const nameFor = useCallback(
    (playerId: string): string => {
      if (playerId === selfPlayerIdRef.current) {
        return "You";
      }
      return playerNameById[playerId] ?? playerId.slice(0, 8);
    },
    [playerNameById],
  );

  useEffect(() => {
    return installAudioUnlock();
  }, []);

  // Keep the reducer's notion of "self" in sync so it can maintain yourHand.
  useEffect(() => {
    setReducerSelfPlayerId(selfPlayerId);
  }, [selfPlayerId]);

  // Auto-scroll the action history to the newest entry.
  useEffect(() => {
    const list = actionHistoryListRef.current;
    if (list) {
      list.scrollTop = list.scrollHeight;
    }
  }, [actionHistoryEntries]);

  // Resolve the current player's id for "You" labels and turn/win sounds.
  useEffect(() => {
    let active = true;
    void (async () => {
      try {
        const response = await getPlayer();
        if (!active) {
          return;
        }
        if (!response.ok) {
          pushErrorNotice(
            "Could not load current player details.",
            "PLAYER_LOOKUP_FAILED",
            "Player error",
          );
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

  // Merge demand deadlines carried in a GameState snapshot; the turn deadline is
  // the entry without a demandId.
  const ingestDeadlines = useCallback((state: GameState) => {
    const incoming: Record<string, number> = {};
    let nextTurnDeadline = 0;
    for (const deadline of state.deadlines ?? []) {
      if (deadline.demandId && deadline.deadlineMs > 0) {
        incoming[deadline.demandId] = deadline.deadlineMs;
      } else if (!deadline.demandId && deadline.deadlineMs > 0) {
        nextTurnDeadline = deadline.deadlineMs;
      }
    }
    setDemandDeadlineMsById((current) => ({ ...current, ...incoming }));
    if (nextTurnDeadline > 0) {
      setTurnDeadlineMs(nextTurnDeadline);
    }
  }, []);

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

    // handleServerMessage applies one decoded ServerMessage. isReplay is true
    // for ActionHistory (reconnect replay), which must never trigger sounds or
    // optimistic mutations of the snapshot (the snapshot is already current).
    const handleServerMessage = (message: ServerMessage, isReplay: boolean) => {
      if (message.error) {
        pushErrorNotice(
          message.error.message || "Server error.",
          message.error.code || "SERVER_ERROR",
        );
      }

      if (message.gameState) {
        const nextState = message.gameState;
        const nextAssetImages = nextState.assetImages ?? [];
        if (nextAssetImages.length > 0) {
          setAssetImageByKey((current) => ({
            ...current,
            ...toLargeAssetImageByKey(nextAssetImages),
          }));
        }
        currentTurnPlayerIdRef.current = nextState.currentPlayerId;
        setPlayerNameById((current) => {
          const incoming = Object.fromEntries(
            nextState.players.map((player) => [
              player.playerId,
              player.displayName,
            ]),
          );
          return { ...current, ...incoming };
        });
        ingestDeadlines(nextState);
        // Authoritative snapshot: it fully replaces optimistic state.
        setGameState(nextState);
      }

      if (message.chatReceived) {
        const chat = message.chatReceived;
        setChatMessages((current) => [
          ...current,
          {
            id: randomId("chat"),
            playerId: chat.playerId ?? "",
            text: chat.payload ?? "",
          },
        ]);
        if (!isReplay && chat.playerId && chat.playerId !== selfPlayerIdRef.current) {
          playChatTock();
        }
      }

      if (message.wonGame) {
        const won = message.wonGame;
        setWonGame((current) => current ?? { ...won });
        if (!isReplay) {
          if (won.playerId === selfPlayerIdRef.current) {
            playWinTocks();
          } else {
            playLossTocks();
          }
        }
      }

      const action = message.action ?? message.actionHistory?.action;
      if (action) {
        applyAction(action, isReplay);
      }
    };

    // applyAction fans one Action into optimistic state, appends a history
    // entry, tracks demand/turn deadlines, and (for live actions only) plays
    // the appropriate sound.
    const applyAction = (action: Action, isReplay: boolean) => {
      const actorId = action.playerId;
      const self = selfPlayerIdRef.current;

      // Optimistic state fan-out (skipped for replay — snapshot is current).
      if (!isReplay) {
        setGameState((current) => {
          if (!current) {
            return current;
          }
          const result = applyActionToGameState(current, action);
          if (action.kind === ActionKind.ACTION_KIND_START_TURN && actorId) {
            currentTurnPlayerIdRef.current = actorId;
          }
          return result.state;
        });
      }

      // Track the turn deadline carried by every live action.
      if (!isReplay && typeof action.turnDeadlineMs === "number" && action.turnDeadlineMs > 0) {
        setTurnDeadlineMs(action.turnDeadlineMs);
      }

      // New demands inherit the action's turn deadline for their countdown.
      const created = action.actionDemandsCreated;
      if (created && created.demands.length > 0 && action.turnDeadlineMs > 0) {
        setDemandDeadlineMsById((current) => {
          const nextMap = { ...current };
          for (const demand of created.demands) {
            if (!nextMap[demand.id]) {
              nextMap[demand.id] = action.turnDeadlineMs;
            }
          }
          return nextMap;
        });
      }

      // Drop deadlines for demands that were just resolved.
      if (action.kind === ActionKind.ACTION_KIND_DEMAND_COMPLIED) {
        const resolvedId = action.actionDemandComplied?.demandId;
        if (resolvedId) {
          setDemandDeadlineMsById((current) => {
            if (!(resolvedId in current)) {
              return current;
            }
            const nextMap = { ...current };
            delete nextMap[resolvedId];
            return nextMap;
          });
        }
      }

      // Action-history log (both live and replay build entries; only live plays
      // sounds). We resolve names via the ref-backed lookup captured here.
      const entry = buildHistoryEntry(action, (playerId) =>
        playerId === self
          ? "You"
          : (playerNameByIdRef.current[playerId] ?? playerId.slice(0, 8)),
      );
      if (entry) {
        setActionHistoryEntries((current) => [...current.slice(-199), entry]);
      }

      // Sounds — live events only.
      if (isReplay) {
        return;
      }
      // A demand that targets you replaces the card tock with a knock.
      const demandsForSelf =
        created?.demands.filter(
          (demand) => demand.playerId === self && demand.isActive,
        ) ?? [];
      const newSelfDemand = demandsForSelf.find(
        (demand) => !knockedDemandIdsRef.current.has(demand.id),
      );
      if (newSelfDemand) {
        for (const demand of demandsForSelf) {
          knockedDemandIdsRef.current.add(demand.id);
        }
        playDemandKnock();
        return;
      }

      if (action.kind === ActionKind.ACTION_KIND_START_TURN && actorId === self) {
        playYourTurnTocks();
        return;
      }

      // Any other real play makes the card tock (empty auto-discards excluded).
      const isEmptyDiscard =
        action.kind === ActionKind.ACTION_KIND_DISCARD_CARDS &&
        (action.actionDiscardCards?.cards.length ?? 0) === 0 &&
        (action.maskedActionDiscardCards?.numCards ?? 0) === 0;
      if (
        action.kind !== ActionKind.ACTION_KIND_START_TURN &&
        !isEmptyDiscard
      ) {
        playCardTock();
      }
    };

    const connectSocket = () => {
      if (didUnmount) {
        return;
      }
      const socket = connectDealNoMercyGameSocket();
      socketRef.current = socket;

      socket.onopen = () => {
        reconnectAttemptRef.current = 0;
        clearReconnectTimer();
      };

      socket.onmessage = (event) => {
        void (async () => {
          try {
            const decoded = await decodeDealNoMercyServerMessage(event.data);
            if (!decoded) {
              return;
            }
            const inner = decoded.error ??
              decoded.chatReceived ??
              decoded.gameState ??
              decoded.wonGame ??
              decoded.action ??
              decoded.actionHistory;
            if (inner === undefined) {
              return;
            }
            // actionHistory frames are the reconnect replay: silent, no fan-out.
            handleServerMessage(decoded, decoded.actionHistory !== undefined);
          } catch (error) {
            console.warn("[no-mercy-ws] failed to handle message", error);
          }
        })();
      };

      socket.onerror = () => {
        // onclose schedules the reconnect.
      };

      socket.onclose = () => {
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
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [gameId, pushErrorNotice, ingestDeadlines]);

  // playerNameById via a ref so the socket effect (which does not depend on it)
  // can resolve fresh names without re-subscribing the socket.
  const playerNameByIdRef = useRef<Record<string, string>>({});
  useEffect(() => {
    playerNameByIdRef.current = playerNameById;
  }, [playerNameById]);

  // -------------------------------------------------------------------------
  // Derived turn state (for discard / pass gating)
  // -------------------------------------------------------------------------

  const handCards = gameState?.yourHand?.cards ?? [];
  const maxHandSize = gameState?.maxHandSize ?? 0;
  const isSelfTurn =
    !!selfPlayerId && selfPlayerId === gameState?.currentPlayerId;
  const movesLeft = gameState?.movesLeft ?? 0;
  const discardRequiredCount = Math.max(0, handCards.length - maxHandSize);
  const isDiscardRequired =
    isSelfTurn && movesLeft <= 0 && discardRequiredCount > 0;

  const players = useMemo<Player[]>(
    () => gameState?.players ?? [],
    [gameState?.players],
  );

  // -------------------------------------------------------------------------
  // Socket send helpers
  // -------------------------------------------------------------------------

  const withSocket = useCallback(
    (action: (socket: WebSocket) => void) => {
      const socket = socketRef.current;
      if (!socket || socket.readyState !== WebSocket.OPEN) {
        pushErrorNotice(
          "Not connected to the game server.",
          "SOCKET_UNAVAILABLE",
          "Connection",
        );
        return;
      }
      action(socket);
    },
    [pushErrorNotice],
  );

  const handleSendChat = useCallback(
    (payload: string) => {
      const trimmed = payload.trim();
      if (!trimmed) {
        return;
      }
      withSocket((socket) => sendDealNoMercyChatMessage(socket, trimmed));
    },
    [withSocket],
  );

  // Turn lifecycle -----------------------------------------------------------

  const handlePassTurn = useCallback(() => {
    if (discardRequiredCount > 0) {
      pushErrorNotice(
        `You must discard ${discardRequiredCount} card(s) before passing.`,
        "DISCARD_REQUIRED_BEFORE_PASS",
        "Cannot pass turn",
      );
      return;
    }
    withSocket((socket) => sendDealNoMercyCompleteTurnMessage(socket));
  }, [discardRequiredCount, pushErrorNotice, withSocket]);

  const handleToggleDiscardCard = useCallback((cardId: string) => {
    setSelectedDiscardCardIds((current) => {
      const next = new Set(current);
      if (next.has(cardId)) {
        next.delete(cardId);
      } else {
        next.add(cardId);
      }
      return next;
    });
  }, []);

  const handleSubmitDiscard = useCallback(() => {
    const cardIds = Array.from(selectedDiscardCardIds);
    if (cardIds.length === 0) {
      return;
    }
    withSocket((socket) => sendDealNoMercyDiscardCardsMessage(socket, cardIds));
    setSelectedDiscardCardIds(new Set());
  }, [selectedDiscardCardIds, withSocket]);

  // Plays --------------------------------------------------------------------

  const handlePlayMoney = useCallback(
    (cardId: string) =>
      withSocket((socket) => sendDealNoMercyPlayMoneyMessage(socket, cardId)),
    [withSocket],
  );

  const handlePlayProperty = useCallback(
    (cardId: string, propertySetId?: string, activeColor?: Color) =>
      withSocket((socket) =>
        sendDealNoMercyPlayPropertyMessage(socket, {
          cardId,
          propertySetId,
          activeColor,
        }),
      ),
    [withSocket],
  );

  const handleRearrangeCard = useCallback(
    (cardId: string, propertySetId?: string, color?: Color) =>
      withSocket((socket) =>
        sendDealNoMercyRearrangeCardMessage(socket, {
          cardId,
          propertySetId,
          color,
        }),
      ),
    [withSocket],
  );

  const handlePlayShack = useCallback(
    (cardId: string, propertySetId: string) =>
      withSocket((socket) =>
        sendDealNoMercyPlayShackMessage(socket, { cardId, propertySetId }),
      ),
    [withSocket],
  );

  const handlePlayBigPayday = useCallback(
    (cardId: string) =>
      withSocket((socket) => sendDealNoMercyPlayBigPaydayMessage(socket, cardId)),
    [withSocket],
  );

  const handlePlayGoAgain = useCallback(
    (cardId: string) =>
      withSocket((socket) => sendDealNoMercyPlayGoAgainMessage(socket, cardId)),
    [withSocket],
  );

  const handlePlaySetSnatcher = useCallback(
    (cardId: string, targetId: string, propertySetId: string) =>
      withSocket((socket) =>
        sendDealNoMercyPlaySetSnatcherMessage(socket, {
          cardId,
          targetId,
          propertySetId,
        }),
      ),
    [withSocket],
  );

  const handlePlayDebtTrap = useCallback(
    (cardId: string, targetId: string) =>
      withSocket((socket) =>
        sendDealNoMercyPlayDebtTrapMessage(socket, { cardId, targetId }),
      ),
    [withSocket],
  );

  const handlePlayYoink = useCallback(
    (cardId: string, targetId: string) =>
      withSocket((socket) =>
        sendDealNoMercyPlayYoinkMessage(socket, { cardId, targetId }),
      ),
    [withSocket],
  );

  const handlePlayBankSwap = useCallback(
    (cardId: string, targetId: string) =>
      withSocket((socket) =>
        sendDealNoMercyPlayBankSwapMessage(socket, { cardId, targetId }),
      ),
    [withSocket],
  );

  const handlePlayRepoMan = useCallback(
    (cardId: string, targetId: string) =>
      withSocket((socket) =>
        sendDealNoMercyPlayRepoManMessage(socket, { cardId, targetId }),
      ),
    [withSocket],
  );

  const handlePlayTaxDay = useCallback(
    (cardId: string, targetId: string) =>
      withSocket((socket) =>
        sendDealNoMercyPlayTaxDayMessage(socket, { cardId, targetId }),
      ),
    [withSocket],
  );

  const handlePlayPickpocket = useCallback(
    (cardId: string, targetId: string, category: StealCategory) =>
      withSocket((socket) =>
        sendDealNoMercyPlayPickpocketMessage(socket, {
          cardId,
          targetId,
          category,
        }),
      ),
    [withSocket],
  );

  const handlePlayPropertyRaid = useCallback(
    (cardId: string, color: Color) =>
      withSocket((socket) =>
        sendDealNoMercyPlayPropertyRaidMessage(socket, { cardId, color }),
      ),
    [withSocket],
  );

  const handlePlayRent = useCallback(
    (cardId: string, color: Color) =>
      withSocket((socket) =>
        sendDealNoMercyPlayRentMessage(socket, { cardId, color }),
      ),
    [withSocket],
  );

  const handlePlayHeist = useCallback(
    (cardId: string, picks: PropertyPick[]) =>
      withSocket((socket) =>
        sendDealNoMercyPlayHeistMessage(socket, { cardId, picks }),
      ),
    [withSocket],
  );

  const handlePlayMarketCrash = useCallback(
    (cardId: string, picks: PropertyPick[]) =>
      withSocket((socket) =>
        sendDealNoMercyPlayMarketCrashMessage(socket, { cardId, picks }),
      ),
    [withSocket],
  );

  // Demand resolution --------------------------------------------------------

  const handleComplyPaymentDemand = useCallback(
    (demandId: string, cardIds: string[]) =>
      withSocket((socket) =>
        sendDealNoMercyComplyPaymentDemandMessage(socket, { demandId, cardIds }),
      ),
    [withSocket],
  );

  const handleComplyPropertyDemand = useCallback(
    (demandId: string) =>
      withSocket((socket) =>
        sendDealNoMercyComplyPropertyDemandMessage(socket, demandId),
      ),
    [withSocket],
  );

  const handleComplyPropertySetDemand = useCallback(
    (demandId: string) =>
      withSocket((socket) =>
        sendDealNoMercyComplyPropertySetDemandMessage(socket, demandId),
      ),
    [withSocket],
  );

  const handleComplyColorPropertiesDemand = useCallback(
    (demandId: string) =>
      withSocket((socket) =>
        sendDealNoMercyComplyColorPropertiesDemandMessage(socket, demandId),
      ),
    [withSocket],
  );

  const handleComplyBankCardDemand = useCallback(
    (demandId: string) =>
      withSocket((socket) =>
        sendDealNoMercyComplyBankCardDemandMessage(socket, demandId),
      ),
    [withSocket],
  );

  const handleComplyBankSwapDemand = useCallback(
    (demandId: string) =>
      withSocket((socket) =>
        sendDealNoMercyComplyBankSwapDemandMessage(socket, demandId),
      ),
    [withSocket],
  );

  const handleComplyDebtTrapDemand = useCallback(
    (demandId: string) =>
      withSocket((socket) =>
        sendDealNoMercyComplyDebtTrapDemandMessage(socket, demandId),
      ),
    [withSocket],
  );

  const handleComplyPickpocketDemand = useCallback(
    (demandId: string) =>
      withSocket((socket) =>
        sendDealNoMercyComplyPickpocketDemandMessage(socket, demandId),
      ),
    [withSocket],
  );

  const handleComplyRepoManDemand = useCallback(
    (demandId: string, keepCardId: string, distribution: DistributionAssignment[]) =>
      withSocket((socket) =>
        sendDealNoMercyComplyRepoManDemandMessage(socket, {
          demandId,
          keepCardId,
          distribution,
        }),
      ),
    [withSocket],
  );

  const handleComplyTaxDayDemand = useCallback(
    (demandId: string, keepCardId: string, distribution: DistributionAssignment[]) =>
      withSocket((socket) =>
        sendDealNoMercyComplyTaxDayDemandMessage(socket, {
          demandId,
          keepCardId,
          distribution,
        }),
      ),
    [withSocket],
  );

  const handleDenyDemand = useCallback(
    (demandId: string, cardId: string) =>
      withSocket((socket) =>
        sendDealNoMercyDenyDemandMessage(socket, { demandId, cardId }),
      ),
    [withSocket],
  );

  const handleSettleDebt = useCallback(
    (debtId: string, cardId: string) =>
      withSocket((socket) =>
        sendDealNoMercySettleDebtMessage(socket, { debtId, cardId }),
      ),
    [withSocket],
  );

  const handleBoardViewModeChange = useCallback(
    (mode: DealNoMercyBoardViewMode) => {
      setBoardViewMode(mode);
      try {
        window.localStorage.setItem(VIEW_MODE_STORAGE_KEY, mode);
      } catch {
        // Ignore storage write failures.
      }
    },
    [],
  );

  // -------------------------------------------------------------------------
  // Game-end overlay derived labels
  // -------------------------------------------------------------------------

  const winnerName = wonGame ? nameFor(wonGame.playerId) : "";
  const winnerPlayer = wonGame
    ? players.find((player) => player.playerId === wonGame.playerId)
    : undefined;
  const winnerAvatarUrl = winnerPlayer?.avatarUrl ?? "";
  const winnerInitial = (winnerName || "?").slice(0, 1).toUpperCase();

  const gamePageClassName = [
    "page",
    "game-page",
    `game-page--${boardViewMode}`,
    wonGame ? "game-page--ended" : "",
    isSidebarCollapsed ? "game-page--sidebar-collapsed" : "",
  ]
    .filter(Boolean)
    .join(" ");

  return (
    <>
      <main className={gamePageClassName}>
        <section className="game-page__board">
          <DealNoMercyGameMount
            gameState={gameState}
            assetImageByKey={assetImageByKey}
            viewMode={boardViewMode}
            selfPlayerId={selfPlayerId}
            demandDeadlineMsById={demandDeadlineMsById}
            turnDeadlineMs={turnDeadlineMs}
            onPassTurn={handlePassTurn}
            onSubmitDiscard={handleSubmitDiscard}
            isDiscardRequired={isDiscardRequired}
            requiredDiscardCount={discardRequiredCount}
            selectedDiscardCardIds={selectedDiscardCardIds}
            onToggleDiscardCard={handleToggleDiscardCard}
            onPlayMoney={handlePlayMoney}
            onPlayProperty={handlePlayProperty}
            onRearrangeCard={handleRearrangeCard}
            onPlayShack={handlePlayShack}
            onPlayBigPayday={handlePlayBigPayday}
            onPlayGoAgain={handlePlayGoAgain}
            onPlaySetSnatcher={handlePlaySetSnatcher}
            onPlayDebtTrap={handlePlayDebtTrap}
            onPlayYoink={handlePlayYoink}
            onPlayBankSwap={handlePlayBankSwap}
            onPlayRepoMan={handlePlayRepoMan}
            onPlayTaxDay={handlePlayTaxDay}
            onPlayPickpocket={handlePlayPickpocket}
            onPlayPropertyRaid={handlePlayPropertyRaid}
            onPlayRent={handlePlayRent}
            onPlayHeist={handlePlayHeist}
            onPlayMarketCrash={handlePlayMarketCrash}
            onComplyPaymentDemand={handleComplyPaymentDemand}
            onComplyPropertyDemand={handleComplyPropertyDemand}
            onComplyPropertySetDemand={handleComplyPropertySetDemand}
            onComplyColorPropertiesDemand={handleComplyColorPropertiesDemand}
            onComplyBankCardDemand={handleComplyBankCardDemand}
            onComplyBankSwapDemand={handleComplyBankSwapDemand}
            onComplyDebtTrapDemand={handleComplyDebtTrapDemand}
            onComplyPickpocketDemand={handleComplyPickpocketDemand}
            onComplyRepoManDemand={handleComplyRepoManDemand}
            onComplyTaxDayDemand={handleComplyTaxDayDemand}
            onDenyDemand={handleDenyDemand}
            onSettleDebt={handleSettleDebt}
            onGameError={(error) => {
              pushErrorNotice(error.message, error.code, "Cannot play card");
            }}
          />
        </section>

        <DealNoMercyGameSidebar
          isSidebarCollapsed={isSidebarCollapsed}
          onToggleSidebar={() => setIsSidebarCollapsed((current) => !current)}
          actionHistoryListRef={actionHistoryListRef}
          actionHistoryEntries={actionHistoryEntries}
          assetImageByKey={assetImageByKey}
          players={players}
          playerNameById={playerNameById}
          selfPlayerId={selfPlayerId}
          chatMessages={chatMessages}
          onSendChatMessage={handleSendChat}
          boardViewMode={boardViewMode}
          onBoardViewModeChange={handleBoardViewModeChange}
        />

        {wonGame ? (
          <DealNoMercyGameEndOverlay
            wonGame={wonGame}
            winnerName={winnerName}
            winnerAvatarUrl={winnerAvatarUrl}
            winnerInitial={winnerInitial}
            onGoHome={() => navigate("/lobby")}
          />
        ) : null}
      </main>

      <ErrorToastStack notices={errorNotices} onDismiss={dismissErrorNotice} />
    </>
  );
};

export default DealNoMercyGamePage;
