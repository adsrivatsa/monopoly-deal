// DealNoMercyBoard renders the interactive board for "Monopoly Deal: No Mercy".
// It mirrors the classic MonopolyDealBoard architecture (expanded + compact
// layouts, compact pan/zoom, per-player money/property/hand areas, a
// pending-play state machine driven by the tapped hand card, demand overlays,
// and a debt-settlement flow) but targets the No Mercy generated types and the
// expanded No Mercy action roster.
//
// Interaction model (mirrors classic):
//   - Tap a hand card to select it, then tap a dropzone (money / property set /
//     new set) OR the Actions pile to commit the play.
//   - Money/property plays resolve immediately (property wilds pop a color
//     picker + can target a set). Action cards open the picker their kind needs
//     (target player, color, set, per-opponent multi-pick, etc.).
//   - Demands targeting you show a comply/deny overlay; NAH! denies when held.
//   - Debt at your turn start blocks all plays until each chip is settled.
import {
  useCallback,
  useEffect,
  useMemo,
  useRef,
  useState,
  type CSSProperties,
  type PointerEvent as ReactPointerEvent,
} from "react";
import {
  AssetKey,
  Category,
  Color,
  DemandKind,
  StealCategory,
  colorToJSON,
  type Card,
  type DebtChip,
  type Demand,
  type DistributionAssignment,
  type GameState,
  type Player,
  type PropertyPick,
  type PropertySet,
} from "../../../generated/deal_no_mercy";
import DealNoMercyCardStackBox from "./DealNoMercyCardStackBox";
import { installAudioUnlock, playTimerTick } from "../../shared/sound";
import "./deal-no-mercy-board.css";

export type DealNoMercyBoardLayoutMode = "expanded" | "compact";

// One settlement decision the debtor makes for an outstanding chip.
export type DebtSettlementInput = {
  debtId: string;
  cardId: string;
};

export type DealNoMercyBoardProps = {
  gameState: GameState | null;
  assetImageByKey: Record<number, string>;
  layoutMode?: DealNoMercyBoardLayoutMode;
  selfPlayerId?: string;
  // demandId -> absolute epoch ms deadline, plus the turn deadline under the
  // reserved "" key.
  demandDeadlineMsById?: Record<string, number>;
  turnDeadlineMs?: number;

  // Turn lifecycle.
  onPassTurn: () => void;
  onSubmitDiscard: () => void;
  isDiscardRequired: boolean;
  requiredDiscardCount: number;
  selectedDiscardCardIds: ReadonlySet<string>;
  onToggleDiscardCard: (cardId: string) => void;

  // Simple / property plays.
  onPlayMoney: (cardId: string) => void;
  onPlayProperty: (
    cardId: string,
    propertySetId?: string,
    activeColor?: Color,
  ) => void;
  onRearrangeCard: (
    cardId: string,
    propertySetId?: string,
    color?: Color,
  ) => void;
  onPlayShack: (cardId: string, propertySetId: string) => void;
  onPlayBigPayday: (cardId: string) => void;
  onPlayGoAgain: (cardId: string) => void;

  // Targeted / choice action plays.
  onPlaySetSnatcher: (
    cardId: string,
    targetId: string,
    propertySetId: string,
  ) => void;
  onPlayDebtTrap: (cardId: string, targetId: string) => void;
  onPlayYoink: (cardId: string, targetId: string) => void;
  onPlayBankSwap: (cardId: string, targetId: string) => void;
  onPlayRepoMan: (cardId: string, targetId: string) => void;
  onPlayTaxDay: (cardId: string, targetId: string) => void;
  onPlayPickpocket: (
    cardId: string,
    targetId: string,
    category: StealCategory,
  ) => void;
  onPlayPropertyRaid: (cardId: string, color: Color) => void;
  onPlayRent: (cardId: string, color: Color) => void;
  onPlayHeist: (cardId: string, picks: PropertyPick[]) => void;
  onPlayMarketCrash: (cardId: string, picks: PropertyPick[]) => void;

  // Demand resolution.
  onComplyPaymentDemand: (demandId: string, cardIds: string[]) => void;
  onComplyPropertyDemand: (demandId: string) => void;
  onComplyPropertySetDemand: (demandId: string) => void;
  onComplyColorPropertiesDemand: (demandId: string) => void;
  onComplyBankCardDemand: (demandId: string) => void;
  onComplyBankSwapDemand: (demandId: string) => void;
  onComplyDebtTrapDemand: (demandId: string) => void;
  onComplyPickpocketDemand: (demandId: string) => void;
  onComplyRepoManDemand: (
    demandId: string,
    keepCardId: string,
    distribution: DistributionAssignment[],
  ) => void;
  onComplyTaxDayDemand: (
    demandId: string,
    keepCardId: string,
    distribution: DistributionAssignment[],
  ) => void;
  onDenyDemand: (demandId: string, cardId: string) => void;

  // Debt settlement (one card per outstanding chip owed by you).
  onSettleDebt: (debtId: string, cardId: string) => void;

  onGameError?: (error: { message: string; code: string }) => void;
};

type Size = { width: number; height: number };
type Pan = { x: number; y: number };

const MIN_ZOOM = 0.15;
const MAX_ZOOM = 2.2;
const BOARD_COLUMN_GAP_REM = 0.7;
const TIMER_TICK_WINDOW_SECONDS = 5;

// ---------------------------------------------------------------------------
// Card / color helpers
// ---------------------------------------------------------------------------

const isPropertyCard = (card: Card): boolean =>
  card.category === Category.CATEGORY_PURE_PROPERTY ||
  card.category === Category.CATEGORY_WILD_PROPERTY;

const isMoneyCard = (card: Card): boolean =>
  card.category === Category.CATEGORY_MONEY;

const isActionCard = (card: Card): boolean =>
  card.category === Category.CATEGORY_ACTION;

const RENT_ASSET_KEYS = new Set<AssetKey>([
  AssetKey.ASSET_KEY_RENT_WILD,
  AssetKey.ASSET_KEY_RENT_BROWN_SKY,
  AssetKey.ASSET_KEY_RENT_PINK_ORANGE,
  AssetKey.ASSET_KEY_RENT_RED_YELLOW,
  AssetKey.ASSET_KEY_RENT_GREEN_BLUE,
  AssetKey.ASSET_KEY_RENT_UTILITY_RAILROAD,
  AssetKey.ASSET_KEY_DOUBLE_RENT_WILD,
  AssetKey.ASSET_KEY_DOUBLE_RENT_BROWN_SKY,
  AssetKey.ASSET_KEY_DOUBLE_RENT_PINK_ORANGE,
  AssetKey.ASSET_KEY_DOUBLE_RENT_RED_YELLOW,
  AssetKey.ASSET_KEY_DOUBLE_RENT_GREEN_BLUE,
  AssetKey.ASSET_KEY_DOUBLE_RENT_UTILITY_RAILROAD,
]);

const isRentAssetKey = (assetKey: AssetKey): boolean =>
  RENT_ASSET_KEYS.has(assetKey);

const ALL_COLORS: Color[] = [
  Color.COLOR_BROWN,
  Color.COLOR_SKY,
  Color.COLOR_PINK,
  Color.COLOR_ORANGE,
  Color.COLOR_RED,
  Color.COLOR_YELLOW,
  Color.COLOR_GREEN,
  Color.COLOR_BLUE,
  Color.COLOR_UTILITY,
  Color.COLOR_RAILROAD,
];

const toSelectableColors = (card: Card): Color[] => {
  const unique = new Set<Color>();
  for (const color of card.colors) {
    if (color === Color.COLOR_UNSPECIFIED || color === Color.UNRECOGNIZED) {
      continue;
    }
    unique.add(color);
  }
  return Array.from(unique);
};

const colorName = (color: Color): string => {
  switch (color) {
    case Color.COLOR_BROWN:
      return "Brown";
    case Color.COLOR_SKY:
      return "Sky";
    case Color.COLOR_PINK:
      return "Pink";
    case Color.COLOR_ORANGE:
      return "Orange";
    case Color.COLOR_RED:
      return "Red";
    case Color.COLOR_YELLOW:
      return "Yellow";
    case Color.COLOR_GREEN:
      return "Green";
    case Color.COLOR_BLUE:
      return "Blue";
    case Color.COLOR_UTILITY:
      return "Utility";
    case Color.COLOR_RAILROAD:
      return "Railroad";
    default:
      return "Unspecified";
  }
};

const colorHex = (color: Color): string => {
  switch (color) {
    case Color.COLOR_BROWN:
      return "#5a2f2e";
    case Color.COLOR_SKY:
      return "#bdd8ff";
    case Color.COLOR_PINK:
      return "#ce3f8d";
    case Color.COLOR_ORANGE:
      return "#ff9f47";
    case Color.COLOR_RED:
      return "#ed453d";
    case Color.COLOR_YELLOW:
      return "#f4e954";
    case Color.COLOR_GREEN:
      return "#2c7241";
    case Color.COLOR_BLUE:
      return "#2b56ba";
    case Color.COLOR_UTILITY:
      return "#bdc9a4";
    case Color.COLOR_RAILROAD:
      return "#222";
    default:
      return "#4e6073";
  }
};

const minPropertyCountForCompleteSet = (color: Color): number => {
  switch (color) {
    case Color.COLOR_BROWN:
    case Color.COLOR_BLUE:
    case Color.COLOR_UTILITY:
      return 2;
    case Color.COLOR_SKY:
    case Color.COLOR_PINK:
    case Color.COLOR_ORANGE:
    case Color.COLOR_RED:
    case Color.COLOR_YELLOW:
    case Color.COLOR_GREEN:
      return 3;
    case Color.COLOR_RAILROAD:
      return 4;
    default:
      return Number.POSITIVE_INFINITY;
  }
};

const countPropertyCards = (propertySet: PropertySet): number =>
  propertySet.cards.filter((card) => isPropertyCard(card)).length;

const isPropertySetComplete = (propertySet: PropertySet): boolean => {
  const required = minPropertyCountForCompleteSet(propertySet.color);
  if (!Number.isFinite(required)) {
    return false;
  }
  return countPropertyCards(propertySet) >= required;
};

const formatClock = (totalSeconds: number): string => {
  const minutes = Math.floor(totalSeconds / 60);
  const seconds = totalSeconds % 60;
  return `${String(minutes).padStart(2, "0")}:${String(seconds).padStart(2, "0")}`;
};

const computeBoardColumns = (playerCount: number): number => {
  if (playerCount <= 2) {
    return 1;
  }
  return Math.ceil(Math.sqrt(playerCount));
};

const clampPan = (
  pan: Pan,
  zoom: number,
  viewportSize: Size,
  contentSize: Size,
): Pan => {
  const scaledWidth = contentSize.width * zoom;
  const scaledHeight = contentSize.height * zoom;
  let nextX = pan.x;
  let nextY = pan.y;
  const bufferX = viewportSize.width * 0.2;
  const bufferY = viewportSize.height * 0.2;

  if (scaledWidth <= viewportSize.width) {
    const cx = (viewportSize.width - scaledWidth) / 2;
    nextX = Math.min(cx + bufferX, Math.max(cx - bufferX, nextX));
  } else {
    const minX = viewportSize.width - scaledWidth - bufferX;
    nextX = Math.min(bufferX, Math.max(minX, nextX));
  }

  if (scaledHeight <= viewportSize.height) {
    const cy = (viewportSize.height - scaledHeight) / 2;
    nextY = Math.min(cy + bufferY, Math.max(cy - bufferY, nextY));
  } else {
    const minY = viewportSize.height - scaledHeight - bufferY;
    nextY = Math.min(bufferY, Math.max(minY, nextY));
  }

  return { x: nextX, y: nextY };
};

const distanceBetween = (
  a: { x: number; y: number },
  b: { x: number; y: number },
): number => Math.hypot(b.x - a.x, b.y - a.y);

const midpointBetween = (
  a: { x: number; y: number },
  b: { x: number; y: number },
): { x: number; y: number } => ({
  x: (a.x + b.x) / 2,
  y: (a.y + b.y) / 2,
});

// ---------------------------------------------------------------------------
// Pending-play state machine types
// ---------------------------------------------------------------------------

// A color picker for a wild-property play/rearrange or a color-choosing action.
type ColorPickerState =
  | {
      mode: "property_play";
      cardId: string;
      propertySetId?: string;
      colors: Color[];
    }
  | { mode: "property_rearrange"; cardId: string; colors: Color[] }
  | {
      mode: "action_color";
      cardId: string;
      action: "property_raid";
      colors: Color[];
    };

// A single-target player picker for target-only action cards.
type TargetPickerState = {
  cardId: string;
  action:
    | "yoink"
    | "bank_swap"
    | "debt_trap"
    | "repo_man"
    | "tax_day"
    | "set_snatcher"
    | "pickpocket";
};

// Set Snatcher: after picking the target, pick one of their complete sets.
type SetSnatcherState = {
  cardId: string;
  targetId: string;
};

// Pickpocket: after picking the target, pick the category to steal.
type PickpocketState = {
  cardId: string;
  targetId: string;
};

// Heist / Market Crash: one pick per eligible opponent, single confirm.
type MultiPickState = {
  cardId: string;
  action: "heist" | "market_crash";
  // targetId -> chosen cardId
  picks: Record<string, string>;
};

// Shack: pick one of your own sets (that has no shack yet).
type ShackState = { cardId: string };

// Distribution (Repo Man / Tax Day comply): keep one card, hand the rest out.
type DistributionState = {
  demandId: string;
  kind: "repo_man" | "tax_day";
  keepCardId?: string;
  // cardId -> recipientId
  assignments: Record<string, string>;
};

// NAH! deny: pick which held NAH! card to use.
type DenyState = { demandId: string; nahCardIds: string[] };

// ---------------------------------------------------------------------------
// Component
// ---------------------------------------------------------------------------

const DealNoMercyBoard = ({
  gameState,
  assetImageByKey,
  layoutMode = "expanded",
  selfPlayerId,
  demandDeadlineMsById,
  turnDeadlineMs = 0,
  onPassTurn,
  onSubmitDiscard,
  isDiscardRequired,
  requiredDiscardCount,
  selectedDiscardCardIds,
  onToggleDiscardCard,
  onPlayMoney,
  onPlayProperty,
  onRearrangeCard,
  onPlayShack,
  onPlayBigPayday,
  onPlayGoAgain,
  onPlaySetSnatcher,
  onPlayDebtTrap,
  onPlayYoink,
  onPlayBankSwap,
  onPlayRepoMan,
  onPlayTaxDay,
  onPlayPickpocket,
  onPlayPropertyRaid,
  onPlayRent,
  onPlayHeist,
  onPlayMarketCrash,
  onComplyPaymentDemand,
  onComplyPropertyDemand,
  onComplyPropertySetDemand,
  onComplyColorPropertiesDemand,
  onComplyBankCardDemand,
  onComplyBankSwapDemand,
  onComplyDebtTrapDemand,
  onComplyPickpocketDemand,
  onComplyRepoManDemand,
  onComplyTaxDayDemand,
  onDenyDemand,
  onSettleDebt,
  onGameError,
}: DealNoMercyBoardProps) => {
  const viewportRef = useRef<HTMLDivElement | null>(null);
  const contentRef = useRef<HTMLDivElement | null>(null);

  // Pan / zoom (compact only, but state is layout-agnostic).
  const [zoom, setZoom] = useState(1);
  const hasInitializedZoomRef = useRef(false);
  const hasUserAdjustedBoardRef = useRef(false);
  const [pan, setPan] = useState<Pan>({ x: 0, y: 0 });
  const [viewportSize, setViewportSize] = useState<Size>({ width: 1, height: 1 });
  const [contentSize, setContentSize] = useState<Size>({ width: 1, height: 1 });
  const [isPanning, setIsPanning] = useState(false);
  const activeTouchPointsRef = useRef<Map<number, { x: number; y: number }>>(
    new Map(),
  );
  const pinchStartDistanceRef = useRef<number | null>(null);
  const pinchStartZoomRef = useRef(1);
  const pinchWorldCenterRef = useRef<{ x: number; y: number } | null>(null);
  const panPointerIdRef = useRef<number | null>(null);
  const panStartPointRef = useRef<{ x: number; y: number } | null>(null);
  const didPanDuringPointerRef = useRef(false);

  // Selection / pending-play state.
  const [selectedHandCardId, setSelectedHandCardId] = useState<string | null>(
    null,
  );
  const [colorPicker, setColorPicker] = useState<ColorPickerState | null>(null);
  const [targetPicker, setTargetPicker] = useState<TargetPickerState | null>(
    null,
  );
  const [setSnatcher, setSetSnatcher] = useState<SetSnatcherState | null>(null);
  const [pickpocket, setPickpocket] = useState<PickpocketState | null>(null);
  const [multiPick, setMultiPick] = useState<MultiPickState | null>(null);
  const [shackPick, setShackPick] = useState<ShackState | null>(null);
  const [rearrangeCardId, setRearrangeCardId] = useState<{
    cardId: string;
    sourceSetId: string;
  } | null>(null);

  // Payment-demand selection.
  const [activePaymentDemandId, setActivePaymentDemandId] = useState<
    string | null
  >(null);
  const [selectedPaymentCardIds, setSelectedPaymentCardIds] = useState<
    Set<string>
  >(() => new Set());

  // Repo Man / Tax Day distribution.
  const [distribution, setDistribution] = useState<DistributionState | null>(
    null,
  );

  // NAH! deny.
  const [denyState, setDenyState] = useState<DenyState | null>(null);

  // Debt settlement.
  const [settleDebtId, setSettleDebtId] = useState<string | null>(null);

  // Clock ticker for timers.
  const [nowMs, setNowMs] = useState(() => Date.now());
  const lastTimerTickSecondRef = useRef<number | null>(null);

  useEffect(() => {
    return installAudioUnlock();
  }, []);

  // -------------------------------------------------------------------------
  // Derived game state
  // -------------------------------------------------------------------------

  const players = useMemo<Player[]>(
    () => gameState?.players ?? [],
    [gameState?.players],
  );

  const currentPlayerId = gameState?.currentPlayerId ?? "";
  const movesLeft = gameState?.movesLeft ?? 0;
  const isSelfTurn = !!selfPlayerId && selfPlayerId === currentPlayerId;
  const hasMovesLeft = movesLeft > 0;

  const yourHand = useMemo<Card[]>(
    () => gameState?.yourHand?.cards ?? [],
    [gameState?.yourHand],
  );

  const moneyByPlayer = useMemo<Record<string, Card[]>>(() => {
    const map: Record<string, Card[]> = {};
    for (const pile of gameState?.money ?? []) {
      map[pile.playerId] = pile.cards;
    }
    return map;
  }, [gameState?.money]);

  const propertySetsByPlayer = useMemo<Record<string, PropertySet[]>>(() => {
    const map: Record<string, PropertySet[]> = {};
    for (const set of gameState?.properties ?? []) {
      (map[set.playerId] ??= []).push(set);
    }
    return map;
  }, [gameState?.properties]);

  // Debt chips grouped by debtor / creditor for the badges.
  const debtChips = useMemo<DebtChip[]>(
    () => gameState?.debtChips ?? [],
    [gameState?.debtChips],
  );
  const debtOwedByPlayer = useMemo<Record<string, number>>(() => {
    const map: Record<string, number> = {};
    for (const chip of debtChips) {
      map[chip.debtorId] = (map[chip.debtorId] ?? 0) + 1;
    }
    return map;
  }, [debtChips]);
  const debtHeldByPlayer = useMemo<Record<string, number>>(() => {
    const map: Record<string, number> = {};
    for (const chip of debtChips) {
      map[chip.creditorId] = (map[chip.creditorId] ?? 0) + 1;
    }
    return map;
  }, [debtChips]);

  // Outstanding debts you owe — these gate all other plays at your turn start.
  const selfDebts = useMemo<DebtChip[]>(() => {
    if (!selfPlayerId) {
      return [];
    }
    return debtChips.filter((chip) => chip.debtorId === selfPlayerId);
  }, [debtChips, selfPlayerId]);
  const hasOutstandingDebt = isSelfTurn && selfDebts.length > 0;

  const selfMoneyCards = useMemo<Card[]>(
    () => (selfPlayerId ? moneyByPlayer[selfPlayerId] ?? [] : []),
    [moneyByPlayer, selfPlayerId],
  );
  const selfPropertySets = useMemo<PropertySet[]>(
    () => (selfPlayerId ? propertySetsByPlayer[selfPlayerId] ?? [] : []),
    [propertySetsByPlayer, selfPlayerId],
  );

  // Every card by id (hand + all banks + all sets), for lookups.
  const cardById = useMemo<Map<string, Card>>(() => {
    const map = new Map<string, Card>();
    for (const card of yourHand) {
      map.set(card.cardId, card);
    }
    for (const pile of gameState?.money ?? []) {
      for (const card of pile.cards) {
        map.set(card.cardId, card);
      }
    }
    for (const set of gameState?.properties ?? []) {
      for (const card of set.cards) {
        map.set(card.cardId, card);
      }
      if (set.shack) {
        map.set(set.shack.cardId, set.shack);
      }
    }
    return map;
  }, [yourHand, gameState?.money, gameState?.properties]);

  const getCardById = useCallback(
    (cardId: string): Card | undefined => cardById.get(cardId),
    [cardById],
  );

  // Held NAH! card ids (for demand denial).
  const nahCardIds = useMemo<string[]>(
    () =>
      yourHand
        .filter((card) => card.assetKey === AssetKey.ASSET_KEY_NAH)
        .map((card) => card.cardId),
    [yourHand],
  );
  const hasNah = nahCardIds.length > 0;

  // Ordered players: self first, so your board reads at a stable position.
  const orderedPlayers = useMemo<Player[]>(() => {
    if (!selfPlayerId) {
      return players;
    }
    const self = players.filter((p) => p.playerId === selfPlayerId);
    const others = players.filter((p) => p.playerId !== selfPlayerId);
    return [...self, ...others];
  }, [players, selfPlayerId]);

  // Active demands targeting you (need a response). Only these gate the board.
  const selfActiveDemands = useMemo<Demand[]>(() => {
    if (!selfPlayerId) {
      return [];
    }
    return (gameState?.demands ?? []).filter(
      (demand) => demand.playerId === selfPlayerId && demand.isActive,
    );
  }, [gameState?.demands, selfPlayerId]);

  // Demands shown in the overlay (self-active first, else spectator view).
  const visibleDemands = useMemo<Demand[]>(() => {
    const demands = gameState?.demands ?? [];
    if (selfActiveDemands.length > 0) {
      return selfActiveDemands.slice(0, 3);
    }
    const sorted = [...demands].sort((a, b) => {
      if (a.isActive === b.isActive) {
        return 0;
      }
      return a.isActive ? -1 : 1;
    });
    return sorted.slice(0, 2);
  }, [gameState?.demands, selfActiveDemands.length]);

  const hasAnyDemand = visibleDemands.length > 0;

  // The board is interactive for plays only when it's your turn, there are no
  // demands to answer, and you have no outstanding debt to settle first.
  const canPlay =
    isSelfTurn &&
    hasMovesLeft &&
    !isDiscardRequired &&
    selfActiveDemands.length === 0 &&
    !hasOutstandingDebt;

  const canSelectHandCards =
    canPlay &&
    !colorPicker &&
    !targetPicker &&
    !setSnatcher &&
    !pickpocket &&
    !multiPick &&
    !shackPick &&
    !rearrangeCardId;

  // Payment selection: your money/property cards are selectable.
  const isSelectingPayment = activePaymentDemandId !== null;
  const selectablePaymentCardById = useMemo<Map<string, Card>>(() => {
    const map = new Map<string, Card>();
    if (!selfPlayerId) {
      return map;
    }
    for (const card of selfMoneyCards) {
      map.set(card.cardId, card);
    }
    for (const set of selfPropertySets) {
      for (const card of set.cards) {
        map.set(card.cardId, card);
      }
    }
    return map;
  }, [selfMoneyCards, selfPropertySets, selfPlayerId]);

  const selectedPaymentTotal = useMemo(() => {
    let total = 0;
    for (const cardId of selectedPaymentCardIds) {
      total += selectablePaymentCardById.get(cardId)?.value ?? 0;
    }
    return total;
  }, [selectablePaymentCardById, selectedPaymentCardIds]);

  const activePaymentDemand = useMemo<Demand | undefined>(
    () =>
      activePaymentDemandId
        ? visibleDemands.find((d) => d.id === activePaymentDemandId)
        : undefined,
    [activePaymentDemandId, visibleDemands],
  );
  const paymentAmount = activePaymentDemand?.paymentDemand?.amount ?? 0;
  const totalPayableValue = useMemo(() => {
    let total = 0;
    for (const card of selectablePaymentCardById.values()) {
      total += card.value ?? 0;
    }
    return total;
  }, [selectablePaymentCardById]);
  // You can confirm if you've selected enough, or you've selected everything
  // payable (a shortfall then issues a debt chip server-side).
  const canConfirmPayment =
    selectedPaymentTotal >= paymentAmount ||
    selectedPaymentCardIds.size >= selectablePaymentCardById.size ||
    selectedPaymentTotal >= totalPayableValue;

  // Empty payment auto-comply (nothing payable, no NAH! held) is now decided
  // server-side: the engine completes such a demand — issuing the debt chip on
  // the shortfall — before it ever reaches this client, so no client-side
  // auto-comply effect is needed. The payment picker still appears when the
  // player can pay or holds a NAH! (a real choice remains).

  // -------------------------------------------------------------------------
  // Timers
  // -------------------------------------------------------------------------

  const demandTimerSecondsById = useMemo<Record<string, number>>(() => {
    const result: Record<string, number> = {};
    if (!demandDeadlineMsById) {
      return result;
    }
    for (const demand of visibleDemands) {
      const deadlineMs = demandDeadlineMsById[demand.id];
      if (typeof deadlineMs === "number" && deadlineMs > 0) {
        result[demand.id] = Math.max(0, Math.ceil((deadlineMs - nowMs) / 1000));
      }
    }
    return result;
  }, [demandDeadlineMsById, nowMs, visibleDemands]);

  const remainingTurnSeconds = useMemo(() => {
    if (turnDeadlineMs <= 0) {
      return 0;
    }
    return Math.max(0, Math.ceil((turnDeadlineMs - nowMs) / 1000));
  }, [turnDeadlineMs, nowMs]);

  const hasAnyDemandDeadline = Object.keys(demandTimerSecondsById).length > 0;
  const shouldShowTurnTimer = isSelfTurn && turnDeadlineMs > 0;

  useEffect(() => {
    if (turnDeadlineMs <= 0 && !hasAnyDemandDeadline) {
      return;
    }
    const timerId = window.setInterval(() => setNowMs(Date.now()), 250);
    return () => window.clearInterval(timerId);
  }, [hasAnyDemandDeadline, turnDeadlineMs]);

  // Tick only during the final seconds of your own turn.
  useEffect(() => {
    if (
      !shouldShowTurnTimer ||
      remainingTurnSeconds <= 0 ||
      remainingTurnSeconds > TIMER_TICK_WINDOW_SECONDS
    ) {
      lastTimerTickSecondRef.current = null;
      return;
    }
    if (lastTimerTickSecondRef.current === remainingTurnSeconds) {
      return;
    }
    lastTimerTickSecondRef.current = remainingTurnSeconds;
    const urgency =
      (TIMER_TICK_WINDOW_SECONDS - remainingTurnSeconds) /
      TIMER_TICK_WINDOW_SECONDS;
    playTimerTick(urgency);
  }, [remainingTurnSeconds, shouldShowTurnTimer]);

  // -------------------------------------------------------------------------
  // Pan / zoom (compact)
  // -------------------------------------------------------------------------

  const columns = useMemo(
    () =>
      layoutMode === "compact" ? 1 : computeBoardColumns(orderedPlayers.length),
    [layoutMode, orderedPlayers.length],
  );

  const playerColumns = useMemo<Player[][]>(() => {
    const next: Player[][] = Array.from({ length: columns }, () => []);
    const rows = Math.max(1, Math.ceil(orderedPlayers.length / columns));
    orderedPlayers.forEach((player, index) => {
      const columnIndex = Math.min(columns - 1, Math.floor(index / rows));
      next[columnIndex].push(player);
    });
    return next;
  }, [columns, orderedPlayers]);

  const boardStyle = useMemo<CSSProperties>(() => {
    // Width must derive from the content's width, never its height: this width
    // is applied to the content element itself, and a height-based value
    // reflows the wrapped rows, changing the height it was derived from — an
    // oscillating resize loop that makes the board jitter. (Classic lesson.)
    const compactBoardWidth =
      layoutMode === "compact"
        ? `${Math.max(viewportSize.width, contentSize.width)}px`
        : undefined;

    return {
      transform: `translate(${pan.x}px, ${pan.y}px) scale(${zoom})`,
      display: "flex",
      alignItems: "flex-start",
      gap: `${BOARD_COLUMN_GAP_REM}rem`,
      ["--dnm-compact-board-width" as string]: compactBoardWidth,
    };
  }, [contentSize.width, layoutMode, pan.x, pan.y, viewportSize.width, zoom]);

  const dynamicMinZoom = useMemo(() => {
    if (viewportSize.width <= 1 || contentSize.width <= 1) {
      return MIN_ZOOM;
    }
    const fitX = viewportSize.width / contentSize.width;
    const fitY = viewportSize.height / contentSize.height;
    return Math.max(MIN_ZOOM, Math.min(fitX, fitY, 1));
  }, [viewportSize, contentSize]);

  const initialBoardZoom = useMemo(() => {
    if (
      viewportSize.width <= 1 ||
      viewportSize.height <= 1 ||
      contentSize.width <= 1 ||
      contentSize.height <= 1
    ) {
      return dynamicMinZoom;
    }
    if (layoutMode === "expanded") {
      return dynamicMinZoom;
    }
    const fitX = viewportSize.width / contentSize.width;
    return Math.min(MAX_ZOOM, Math.max(MIN_ZOOM, fitX));
  }, [contentSize, dynamicMinZoom, layoutMode, viewportSize]);

  useEffect(() => {
    const viewport = viewportRef.current;
    const content = contentRef.current;
    if (!viewport || !content) {
      return;
    }
    const updateSizes = () => {
      setViewportSize({
        width: Math.max(viewport.clientWidth, 1),
        height: Math.max(viewport.clientHeight, 1),
      });
      setContentSize({
        width: Math.max(content.scrollWidth, 1),
        height: Math.max(content.scrollHeight, 1),
      });
    };
    const viewportObserver = new ResizeObserver(updateSizes);
    const contentObserver = new ResizeObserver(updateSizes);
    viewportObserver.observe(viewport);
    contentObserver.observe(content);
    updateSizes();
    return () => {
      viewportObserver.disconnect();
      contentObserver.disconnect();
    };
  }, [orderedPlayers.length, layoutMode]);

  // Reset the initial zoom/pan once sizes are known (until the user adjusts).
  useEffect(() => {
    if (hasUserAdjustedBoardRef.current) {
      return;
    }
    if (viewportSize.width <= 1 || contentSize.width <= 1) {
      return;
    }
    if (!hasInitializedZoomRef.current) {
      hasInitializedZoomRef.current = true;
    }
    const nextZoom = initialBoardZoom;
    setZoom(nextZoom);
    setPan(
      clampPan(
        {
          x: (viewportSize.width - contentSize.width * nextZoom) / 2,
          y: 0,
        },
        nextZoom,
        viewportSize,
        contentSize,
      ),
    );
  }, [contentSize, initialBoardZoom, viewportSize]);

  // Reset user-adjust flag when switching layout so it re-fits.
  useEffect(() => {
    hasUserAdjustedBoardRef.current = false;
    hasInitializedZoomRef.current = false;
  }, [layoutMode]);

  const onBoardWheel = useCallback(
    (event: WheelEvent) => {
      event.preventDefault();
      hasUserAdjustedBoardRef.current = true;
      const viewport = viewportRef.current;
      if (!viewport) {
        return;
      }
      const rect = viewport.getBoundingClientRect();
      const localX = event.clientX - rect.left;
      const localY = event.clientY - rect.top;
      const factor = event.deltaY < 0 ? 1.05 : 0.95;
      setZoom((currentZoom) => {
        const nextZoom = Math.min(
          MAX_ZOOM,
          Math.max(dynamicMinZoom, currentZoom * factor),
        );
        if (nextZoom === currentZoom) {
          return currentZoom;
        }
        setPan((currentPan) => {
          const worldX = (localX - currentPan.x) / currentZoom;
          const worldY = (localY - currentPan.y) / currentZoom;
          return clampPan(
            { x: localX - worldX * nextZoom, y: localY - worldY * nextZoom },
            nextZoom,
            viewportSize,
            contentSize,
          );
        });
        return nextZoom;
      });
    },
    [contentSize, dynamicMinZoom, viewportSize],
  );

  useEffect(() => {
    const viewport = viewportRef.current;
    if (!viewport || layoutMode !== "compact") {
      return;
    }
    viewport.addEventListener("wheel", onBoardWheel, { passive: false });
    return () => viewport.removeEventListener("wheel", onBoardWheel);
  }, [layoutMode, onBoardWheel]);

  const stopPan = useCallback((event: ReactPointerEvent<HTMLElement>) => {
    activeTouchPointsRef.current.delete(event.pointerId);
    if (activeTouchPointsRef.current.size < 2) {
      pinchStartDistanceRef.current = null;
      pinchWorldCenterRef.current = null;
    }
    if (activeTouchPointsRef.current.size === 0) {
      panPointerIdRef.current = null;
      panStartPointRef.current = null;
      setIsPanning(false);
    }
  }, []);

  const onBoardPointerDown = useCallback(
    (event: ReactPointerEvent<HTMLDivElement>) => {
      if (layoutMode !== "compact") {
        return;
      }
      if (event.pointerType === "touch") {
        activeTouchPointsRef.current.set(event.pointerId, {
          x: event.clientX,
          y: event.clientY,
        });
        if (activeTouchPointsRef.current.size === 1) {
          panPointerIdRef.current = event.pointerId;
          panStartPointRef.current = { x: event.clientX, y: event.clientY };
          didPanDuringPointerRef.current = false;
        } else if (activeTouchPointsRef.current.size >= 2) {
          const [first, second] = Array.from(
            activeTouchPointsRef.current.values(),
          );
          const center = midpointBetween(first, second);
          const rect = event.currentTarget.getBoundingClientRect();
          const local = { x: center.x - rect.left, y: center.y - rect.top };
          pinchStartDistanceRef.current = distanceBetween(first, second);
          pinchStartZoomRef.current = zoom;
          pinchWorldCenterRef.current = {
            x: (local.x - pan.x) / zoom,
            y: (local.y - pan.y) / zoom,
          };
          panPointerIdRef.current = null;
          panStartPointRef.current = null;
          setIsPanning(false);
        }
        return;
      }
      if (event.button !== 0) {
        return;
      }
      didPanDuringPointerRef.current = false;
      panPointerIdRef.current = event.pointerId;
      panStartPointRef.current = { x: event.clientX, y: event.clientY };
    },
    [layoutMode, pan.x, pan.y, zoom],
  );

  const onBoardPointerMove = useCallback(
    (event: ReactPointerEvent<HTMLDivElement>) => {
      if (layoutMode !== "compact") {
        return;
      }
      if (
        event.pointerType === "touch" &&
        activeTouchPointsRef.current.has(event.pointerId)
      ) {
        activeTouchPointsRef.current.set(event.pointerId, {
          x: event.clientX,
          y: event.clientY,
        });
      }

      const pinchStartDistance = pinchStartDistanceRef.current;
      const worldCenter = pinchWorldCenterRef.current;
      if (
        activeTouchPointsRef.current.size >= 2 &&
        pinchStartDistance !== null &&
        worldCenter
      ) {
        const [first, second] = Array.from(
          activeTouchPointsRef.current.values(),
        );
        const center = midpointBetween(first, second);
        const currentDistance = distanceBetween(first, second);
        const scale = currentDistance / pinchStartDistance;
        const nextZoom = Math.min(
          MAX_ZOOM,
          Math.max(dynamicMinZoom, pinchStartZoomRef.current * scale),
        );
        const rect = event.currentTarget.getBoundingClientRect();
        const local = {
          x: center.x - rect.left,
          y: center.y - rect.top,
        };
        hasUserAdjustedBoardRef.current = true;
        setZoom(nextZoom);
        setPan(
          clampPan(
            {
              x: local.x - worldCenter.x * nextZoom,
              y: local.y - worldCenter.y * nextZoom,
            },
            nextZoom,
            viewportSize,
            contentSize,
          ),
        );
        return;
      }

      const previous = panStartPointRef.current;
      if (!previous || panPointerIdRef.current !== event.pointerId) {
        return;
      }
      const deltaX = event.clientX - previous.x;
      const deltaY = event.clientY - previous.y;
      if (!didPanDuringPointerRef.current) {
        if (Math.hypot(deltaX, deltaY) < 4) {
          return;
        }
        didPanDuringPointerRef.current = true;
        hasUserAdjustedBoardRef.current = true;
        setIsPanning(true);
      }
      event.preventDefault();
      panStartPointRef.current = { x: event.clientX, y: event.clientY };
      setPan((current) =>
        clampPan(
          { x: current.x + deltaX, y: current.y + deltaY },
          zoom,
          viewportSize,
          contentSize,
        ),
      );
    },
    [contentSize, dynamicMinZoom, layoutMode, viewportSize, zoom],
  );

  // -------------------------------------------------------------------------
  // Reset transient pending state whenever it becomes invalid
  // -------------------------------------------------------------------------

  useEffect(() => {
    if (!canPlay) {
      setSelectedHandCardId(null);
      setColorPicker(null);
      setTargetPicker(null);
      setSetSnatcher(null);
      setPickpocket(null);
      setMultiPick(null);
      setShackPick(null);
      setRearrangeCardId(null);
    }
  }, [canPlay]);

  useEffect(() => {
    if (selfActiveDemands.length === 0) {
      setActivePaymentDemandId(null);
      setSelectedPaymentCardIds(new Set());
      setDistribution(null);
      setDenyState(null);
    }
  }, [selfActiveDemands.length]);

  useEffect(() => {
    if (!hasOutstandingDebt) {
      setSettleDebtId(null);
    }
  }, [hasOutstandingDebt]);


  const raiseError = useCallback(
    (message: string, code: string) => {
      onGameError?.({ message, code });
    },
    [onGameError],
  );

  // -------------------------------------------------------------------------
  // Play flows
  // -------------------------------------------------------------------------

  // Commit the selected hand card as money (or via the Actions pile for
  // action cards / rent / etc.).
  const playSelectedAsMoneyOrAction = useCallback(
    (cardId: string) => {
      const card = getCardById(cardId);
      if (!card) {
        return;
      }

      // Money and property cards go to the money pile as raw value.
      if (isMoneyCard(card) || isPropertyCard(card)) {
        onPlayMoney(cardId);
        return;
      }

      // Action cards: route to the picker their kind needs, else bank as money.
      switch (card.assetKey) {
        case AssetKey.ASSET_KEY_BIG_PAYDAY:
          onPlayBigPayday(cardId);
          return;
        case AssetKey.ASSET_KEY_GO_AGAIN:
          // Engine treats GO AGAIN as the final play of the turn.
          if (movesLeft !== 1) {
            raiseError(
              "GO AGAIN! can only be played as your last move of the turn.",
              "GO_AGAIN_NOT_LAST_PLAY",
            );
            return;
          }
          onPlayGoAgain(cardId);
          return;
        case AssetKey.ASSET_KEY_SHACK: {
          const eligible = selfPropertySets.filter((set) => !set.shack);
          if (eligible.length === 0) {
            raiseError(
              "You have no set without a shack to attach it to.",
              "SHACK_NO_ELIGIBLE_SET",
            );
            return;
          }
          setShackPick({ cardId });
          return;
        }
        case AssetKey.ASSET_KEY_YOINK:
          setTargetPicker({ cardId, action: "yoink" });
          return;
        case AssetKey.ASSET_KEY_BANK_SWAP:
          if (selfMoneyCards.length === 0) {
            raiseError(
              "BANK SWAP needs a non-empty bank of your own.",
              "BANK_SWAP_EMPTY_BANK",
            );
            return;
          }
          setTargetPicker({ cardId, action: "bank_swap" });
          return;
        case AssetKey.ASSET_KEY_DEBT_TRAP:
          setTargetPicker({ cardId, action: "debt_trap" });
          return;
        case AssetKey.ASSET_KEY_REPO_MAN:
          setTargetPicker({ cardId, action: "repo_man" });
          return;
        case AssetKey.ASSET_KEY_TAX_DAY:
          setTargetPicker({ cardId, action: "tax_day" });
          return;
        case AssetKey.ASSET_KEY_SET_SNATCHER:
          setTargetPicker({ cardId, action: "set_snatcher" });
          return;
        case AssetKey.ASSET_KEY_PICKPOCKET:
          setTargetPicker({ cardId, action: "pickpocket" });
          return;
        case AssetKey.ASSET_KEY_PROPERTY_RAID:
          setColorPicker({
            mode: "action_color",
            cardId,
            action: "property_raid",
            colors: ALL_COLORS,
          });
          return;
        case AssetKey.ASSET_KEY_HEIST:
          setMultiPick({ cardId, action: "heist", picks: {} });
          return;
        case AssetKey.ASSET_KEY_MARKET_CRASH:
          setMultiPick({ cardId, action: "market_crash", picks: {} });
          return;
        default:
          break;
      }

      if (isRentAssetKey(card.assetKey)) {
        // Rent charges every opponent regardless of color; the color only picks
        // which of your sets sets the amount. The server auto-selects the
        // highest-rent eligible color when we send COLOR_UNSPECIFIED, so there
        // is no color pick to force here.
        onPlayRent(cardId, Color.COLOR_UNSPECIFIED);
        setSelectedHandCardId(null);
        return;
      }

      // NAH! and any unhandled action card can only be banked here.
      onPlayMoney(cardId);
    },
    [
      getCardById,
      movesLeft,
      onPlayBigPayday,
      onPlayGoAgain,
      onPlayMoney,
      onPlayRent,
      raiseError,
      selfMoneyCards.length,
      selfPropertySets,
    ],
  );

  // Commit the selected hand card as a property onto a set (or a new set).
  const playSelectedAsProperty = useCallback(
    (cardId: string, propertySetId?: string) => {
      const card = getCardById(cardId);
      if (!card || !isPropertyCard(card)) {
        return;
      }
      const selectable = toSelectableColors(card);
      if (!propertySetId && selectable.length > 1) {
        setColorPicker({
          mode: "property_play",
          cardId,
          propertySetId,
          colors: selectable,
        });
        return;
      }
      onPlayProperty(
        cardId,
        propertySetId,
        selectable.length === 1 ? selectable[0] : undefined,
      );
    },
    [getCardById, onPlayProperty],
  );

  // Rearrange a property already on your table to another set / new color.
  const applyRearrange = useCallback(
    (targetSetId?: string) => {
      if (!rearrangeCardId) {
        return;
      }
      const card = getCardById(rearrangeCardId.cardId);
      if (!card || !isPropertyCard(card)) {
        setRearrangeCardId(null);
        return;
      }
      if (targetSetId && targetSetId === rearrangeCardId.sourceSetId) {
        setRearrangeCardId(null);
        return;
      }
      if (!targetSetId) {
        const selectable = toSelectableColors(card);
        if (
          card.category === Category.CATEGORY_WILD_PROPERTY &&
          selectable.length > 1
        ) {
          setColorPicker({
            mode: "property_rearrange",
            cardId: card.cardId,
            colors: selectable,
          });
          setRearrangeCardId(null);
          return;
        }
        onRearrangeCard(card.cardId, undefined, selectable[0]);
        setRearrangeCardId(null);
        return;
      }
      onRearrangeCard(card.cardId, targetSetId);
      setRearrangeCardId(null);
    },
    [getCardById, onRearrangeCard, rearrangeCardId],
  );

  // -------------------------------------------------------------------------
  // Demand handlers
  // -------------------------------------------------------------------------

  const complyDemand = useCallback(
    (demand: Demand) => {
      switch (demand.demandKind) {
        case DemandKind.DEMAND_KIND_PAYMENT: {
          if (activePaymentDemandId !== demand.id) {
            setActivePaymentDemandId(demand.id);
            setSelectedPaymentCardIds(new Set());
            return;
          }
          if (!canConfirmPayment) {
            return;
          }
          onComplyPaymentDemand(demand.id, Array.from(selectedPaymentCardIds));
          setActivePaymentDemandId(null);
          setSelectedPaymentCardIds(new Set());
          return;
        }
        case DemandKind.DEMAND_KIND_PROPERTY:
          onComplyPropertyDemand(demand.id);
          return;
        case DemandKind.DEMAND_KIND_PROPERTY_SET:
          onComplyPropertySetDemand(demand.id);
          return;
        case DemandKind.DEMAND_KIND_COLOR_PROPERTIES:
          onComplyColorPropertiesDemand(demand.id);
          return;
        case DemandKind.DEMAND_KIND_BANK_CARD:
          onComplyBankCardDemand(demand.id);
          return;
        case DemandKind.DEMAND_KIND_BANK_SWAP:
          onComplyBankSwapDemand(demand.id);
          return;
        case DemandKind.DEMAND_KIND_DEBT_TRAP:
          onComplyDebtTrapDemand(demand.id);
          return;
        case DemandKind.DEMAND_KIND_PICKPOCKET:
          onComplyPickpocketDemand(demand.id);
          return;
        case DemandKind.DEMAND_KIND_REPO_MAN:
          setDistribution({
            demandId: demand.id,
            kind: "repo_man",
            assignments: {},
          });
          return;
        case DemandKind.DEMAND_KIND_TAX_DAY:
          setDistribution({
            demandId: demand.id,
            kind: "tax_day",
            assignments: {},
          });
          return;
        default:
          return;
      }
    },
    [
      activePaymentDemandId,
      canConfirmPayment,
      onComplyBankCardDemand,
      onComplyBankSwapDemand,
      onComplyColorPropertiesDemand,
      onComplyDebtTrapDemand,
      onComplyPaymentDemand,
      onComplyPickpocketDemand,
      onComplyPropertyDemand,
      onComplyPropertySetDemand,
      selectedPaymentCardIds,
    ],
  );

  const startDeny = useCallback(
    (demandId: string) => {
      if (nahCardIds.length === 0) {
        return;
      }
      // One NAH! held: use it directly. Otherwise, prompt (rare).
      if (nahCardIds.length === 1) {
        onDenyDemand(demandId, nahCardIds[0]);
        return;
      }
      setDenyState({ demandId, nahCardIds });
    },
    [nahCardIds, onDenyDemand],
  );

  const togglePaymentCard = useCallback((card: Card) => {
    setSelectedPaymentCardIds((current) => {
      const next = new Set(current);
      if (next.has(card.cardId)) {
        next.delete(card.cardId);
      } else {
        next.add(card.cardId);
      }
      return next;
    });
  }, []);

  // -------------------------------------------------------------------------
  // Distribution (Repo Man / Tax Day) helpers
  // -------------------------------------------------------------------------

  // The cards the demand target must keep-one-of / distribute-the-rest. Repo
  // Man works over the target's properties; Tax Day over their bank.
  const distributionCards = useMemo<Card[]>(() => {
    if (!distribution || !selfPlayerId) {
      return [];
    }
    if (distribution.kind === "repo_man") {
      return selfPropertySets.flatMap((set) => set.cards);
    }
    return selfMoneyCards;
  }, [distribution, selfMoneyCards, selfPropertySets, selfPlayerId]);

  const distributionRecipients = useMemo<Player[]>(() => {
    // Any player except the demand target (you) may receive.
    return players.filter((player) => player.playerId !== selfPlayerId);
  }, [players, selfPlayerId]);

  const distributionComplete = useMemo(() => {
    if (!distribution || !distribution.keepCardId) {
      return false;
    }
    // Every non-kept card must be assigned exactly one recipient.
    return distributionCards.every(
      (card) =>
        card.cardId === distribution.keepCardId ||
        !!distribution.assignments[card.cardId],
    );
  }, [distribution, distributionCards]);

  const submitDistribution = useCallback(() => {
    if (!distribution || !distribution.keepCardId || !distributionComplete) {
      return;
    }
    const assignments: DistributionAssignment[] = distributionCards
      .filter((card) => card.cardId !== distribution.keepCardId)
      .map((card) => ({
        cardId: card.cardId,
        recipientId: distribution.assignments[card.cardId],
      }));
    if (distribution.kind === "repo_man") {
      onComplyRepoManDemand(
        distribution.demandId,
        distribution.keepCardId,
        assignments,
      );
    } else {
      onComplyTaxDayDemand(
        distribution.demandId,
        distribution.keepCardId,
        assignments,
      );
    }
    setDistribution(null);
  }, [
    distribution,
    distributionCards,
    distributionComplete,
    onComplyRepoManDemand,
    onComplyTaxDayDemand,
  ]);

  // -------------------------------------------------------------------------
  // Multi-pick (Heist / Market Crash) helpers
  // -------------------------------------------------------------------------

  // Eligible opponents for the active multi-pick, and their selectable cards.
  const multiPickTargets = useMemo<
    { player: Player; cards: Card[] }[]
  >(() => {
    if (!multiPick) {
      return [];
    }
    const opponents = players.filter((p) => p.playerId !== selfPlayerId);
    return opponents
      .map((player) => {
        const cards =
          multiPick.action === "heist"
            ? moneyByPlayer[player.playerId] ?? []
            : (propertySetsByPlayer[player.playerId] ?? []).flatMap(
                (set) => set.cards,
              );
        return { player, cards };
      })
      .filter((entry) => entry.cards.length > 0);
  }, [multiPick, moneyByPlayer, players, propertySetsByPlayer, selfPlayerId]);

  const multiPickComplete = useMemo(() => {
    if (!multiPick) {
      return false;
    }
    if (multiPickTargets.length === 0) {
      return false;
    }
    return multiPickTargets.every(
      (entry) => !!multiPick.picks[entry.player.playerId],
    );
  }, [multiPick, multiPickTargets]);

  const submitMultiPick = useCallback(() => {
    if (!multiPick || !multiPickComplete) {
      return;
    }
    const picks: PropertyPick[] = multiPickTargets.map((entry) => ({
      targetId: entry.player.playerId,
      cardId: multiPick.picks[entry.player.playerId],
    }));
    if (multiPick.action === "heist") {
      onPlayHeist(multiPick.cardId, picks);
    } else {
      onPlayMarketCrash(multiPick.cardId, picks);
    }
    setMultiPick(null);
    setSelectedHandCardId(null);
  }, [multiPick, multiPickComplete, multiPickTargets, onPlayHeist, onPlayMarketCrash]);

  // -------------------------------------------------------------------------
  // Target picker: candidate players + downstream steps
  // -------------------------------------------------------------------------

  const targetCandidates = useMemo<Player[]>(() => {
    if (!targetPicker) {
      return [];
    }
    const opponents = players.filter((p) => p.playerId !== selfPlayerId);
    switch (targetPicker.action) {
      case "bank_swap":
      case "yoink":
      case "debt_trap":
      case "pickpocket":
        return opponents;
      case "repo_man":
      case "tax_day":
        // Target must have at least one property (repo) / bank card (tax).
        return opponents.filter((player) =>
          targetPicker.action === "repo_man"
            ? (propertySetsByPlayer[player.playerId] ?? []).some(
                (set) => set.cards.length > 0,
              )
            : (moneyByPlayer[player.playerId] ?? []).length > 0,
        );
      case "set_snatcher":
        return opponents.filter((player) =>
          (propertySetsByPlayer[player.playerId] ?? []).some((set) =>
            isPropertySetComplete(set),
          ),
        );
      default:
        return opponents;
    }
  }, [
    targetPicker,
    players,
    selfPlayerId,
    moneyByPlayer,
    propertySetsByPlayer,
  ]);

  const onSelectTarget = useCallback(
    (targetId: string) => {
      if (!targetPicker) {
        return;
      }
      const { cardId, action } = targetPicker;
      switch (action) {
        case "yoink":
          onPlayYoink(cardId, targetId);
          break;
        case "bank_swap":
          onPlayBankSwap(cardId, targetId);
          break;
        case "debt_trap":
          onPlayDebtTrap(cardId, targetId);
          break;
        case "repo_man":
          onPlayRepoMan(cardId, targetId);
          break;
        case "tax_day":
          onPlayTaxDay(cardId, targetId);
          break;
        case "set_snatcher":
          setSetSnatcher({ cardId, targetId });
          setTargetPicker(null);
          return;
        case "pickpocket":
          setPickpocket({ cardId, targetId });
          setTargetPicker(null);
          return;
      }
      setTargetPicker(null);
      setSelectedHandCardId(null);
    },
    [
      targetPicker,
      onPlayYoink,
      onPlayBankSwap,
      onPlayDebtTrap,
      onPlayRepoMan,
      onPlayTaxDay,
    ],
  );

  // Complete sets of the set-snatcher target (the pickable options).
  const setSnatcherSets = useMemo<PropertySet[]>(() => {
    if (!setSnatcher) {
      return [];
    }
    return (propertySetsByPlayer[setSnatcher.targetId] ?? []).filter((set) =>
      isPropertySetComplete(set),
    );
  }, [setSnatcher, propertySetsByPlayer]);

  // -------------------------------------------------------------------------
  // Render helpers
  // -------------------------------------------------------------------------

  const nameFor = useCallback(
    (playerId: string): string => {
      if (playerId === selfPlayerId) {
        return "You";
      }
      const player = players.find((p) => p.playerId === playerId);
      return player?.displayName ?? playerId.slice(0, 8);
    },
    [players, selfPlayerId],
  );

  // An inline avatar + name for a player, used inside demand narration so every
  // demand shows the involved players' images next to their names (mirrors
  // classic's md-demand-inline-player).
  const inlinePlayer = useCallback(
    (playerId: string, label?: string) => {
      const player = players.find((p) => p.playerId === playerId);
      const name = label ?? player?.displayName ?? playerId.slice(0, 8);
      return (
        <span className="dnm-demand-inline-player">
          <img
            className="dnm-demand-source__avatar dnm-demand-inline-player__avatar"
            src={player?.avatarUrl ?? ""}
            alt={name}
            loading="lazy"
            referrerPolicy="no-referrer"
          />
          <span className="dnm-demand-source__name">{name}</span>
        </span>
      );
    },
    [players],
  );

  const currentTurnPlayer = players.find(
    (player) => player.playerId === currentPlayerId,
  );

  const shouldDimBoard = hasAnyDemand || hasOutstandingDebt;

  const renderDebtChips = (playerId: string) => {
    const owed = debtOwedByPlayer[playerId] ?? 0;
    const held = debtHeldByPlayer[playerId] ?? 0;
    if (owed === 0 && held === 0) {
      return null;
    }
    return (
      <span className="dnm-debt-chips">
        {owed > 0 ? (
          <span
            className="dnm-debt-chip dnm-debt-chip--owed"
            title={`${owed} debt chip(s) owed`}
          >
            -{owed}
          </span>
        ) : null}
        {held > 0 ? (
          <span
            className="dnm-debt-chip dnm-debt-chip--held"
            title={`${held} debt chip(s) owed to this player`}
          >
            +{held}
          </span>
        ) : null}
      </span>
    );
  };

  const renderPlayerBoard = (player: Player) => {
    const isSelfBoard = !!selfPlayerId && player.playerId === selfPlayerId;
    const isCurrentPlayer = player.playerId === currentPlayerId;
    const playerMoney = moneyByPlayer[player.playerId] ?? [];
    const propertySets = propertySetsByPlayer[player.playerId] ?? [];
    const playerMoneyTotal = playerMoney.reduce(
      (total, card) => total + (card.value ?? 0),
      0,
    );

    const canPlaceOnOwnBoard = isSelfBoard && canPlay && !!selectedHandCardId;
    const canRearrangeOnOwnBoard =
      isSelfBoard && canPlay && !selectedHandCardId;

    // Highlighting when this player is a valid target for the active picker.
    const isMultiPickTarget =
      !!multiPick &&
      multiPickTargets.some((entry) => entry.player.playerId === player.playerId);

    return (
      <article
        key={player.playerId}
        className={[
          "dnm-player-board",
          isCurrentPlayer ? "dnm-player-board--current" : "",
          isMultiPickTarget ? "dnm-player-board--target-selectable" : "",
        ]
          .filter(Boolean)
          .join(" ")}
      >
        <header className="dnm-player-board__header">
          <img
            className="dnm-player-board__avatar"
            src={player.avatarUrl}
            alt={player.displayName}
            loading="lazy"
            referrerPolicy="no-referrer"
          />
          <div>
            <p className="dnm-player-board__name">
              {isSelfBoard ? "You" : player.displayName}
            </p>
            <p className="dnm-player-board__stats">
              ${playerMoneyTotal}M · {player.completedSets} sets ·{" "}
              {player.handCards} in hand
            </p>
          </div>
          {renderDebtChips(player.playerId)}
        </header>

        <div className="dnm-player-board__body">
          {(playerMoney.length > 0 || canPlaceOnOwnBoard || isSelectingPayment) && (
            <div
              className="dnm-dropzone dnm-dropzone--money"
              onClick={() => {
                if (canRearrangeOnOwnBoard && rearrangeCardId) {
                  applyRearrange(undefined);
                  return;
                }
                if (canPlaceOnOwnBoard && selectedHandCardId) {
                  playSelectedAsMoneyOrAction(selectedHandCardId);
                  setSelectedHandCardId(null);
                }
              }}
            >
              <DealNoMercyCardStackBox
                title={`Bank · $${playerMoneyTotal}M`}
                cards={playerMoney}
                assetImageByKey={assetImageByKey}
                layout="stack"
                emptyLabel={
                  isSelfBoard && isSelfTurn
                    ? "Select a card, then tap here to bank it."
                    : ""
                }
                selectableCards={isSelfBoard && isSelectingPayment}
                selectedCardIds={
                  isSelfBoard && isSelectingPayment
                    ? selectedPaymentCardIds
                    : undefined
                }
                onCardClick={
                  isSelfBoard && isSelectingPayment
                    ? (card) => togglePaymentCard(card)
                    : undefined
                }
              />
            </div>
          )}

          {propertySets.map((propertySet, index) => {
            const complete = isPropertySetComplete(propertySet);
            const isShackTarget =
              isSelfBoard && !!shackPick && !propertySet.shack;
            const isShackDisabled =
              isSelfBoard && !!shackPick && !!propertySet.shack;
            const isSnatcherOption =
              !!setSnatcher &&
              setSnatcher.targetId === player.playerId &&
              complete;

            return (
              <div
                key={propertySet.propertySetId}
                className={[
                  "dnm-dropzone",
                  "dnm-dropzone--property-set",
                  isShackTarget || isSnatcherOption ? "is-set-selected" : "",
                  isShackDisabled ? "is-set-disabled" : "",
                ]
                  .filter(Boolean)
                  .join(" ")}
                onClick={(event) => {
                  if ((event.target as HTMLElement | null)?.closest(".dnm-card")) {
                    return;
                  }
                  // Shack attachment.
                  if (isShackTarget && shackPick) {
                    onPlayShack(shackPick.cardId, propertySet.propertySetId);
                    setShackPick(null);
                    setSelectedHandCardId(null);
                    return;
                  }
                  // Set Snatcher set pick.
                  if (isSnatcherOption && setSnatcher) {
                    onPlaySetSnatcher(
                      setSnatcher.cardId,
                      setSnatcher.targetId,
                      propertySet.propertySetId,
                    );
                    setSetSnatcher(null);
                    setSelectedHandCardId(null);
                    return;
                  }
                  // Rearrange destination.
                  if (isSelfBoard && canRearrangeOnOwnBoard && rearrangeCardId) {
                    applyRearrange(propertySet.propertySetId);
                    return;
                  }
                  // Place selected property on this set.
                  if (isSelfBoard && canPlaceOnOwnBoard && selectedHandCardId) {
                    playSelectedAsProperty(
                      selectedHandCardId,
                      propertySet.propertySetId,
                    );
                    setSelectedHandCardId(null);
                  }
                }}
              >
                <div className="dnm-property-set-wrap">
                  {propertySet.shack ? (
                    <span className="dnm-shack-badge" title="Shack: +rent">
                      Shack
                    </span>
                  ) : null}
                  <DealNoMercyCardStackBox
                    title={`Set ${index + 1}${complete ? " · Complete" : ""}`}
                    cards={propertySet.cards}
                    assetImageByKey={assetImageByKey}
                    layout="stack"
                    color={propertySet.color}
                    emptyLabel=""
                    selectableCards={
                      // On your own board, tap a property to start a rearrange.
                      isSelfBoard && canRearrangeOnOwnBoard
                    }
                    onCardClick={
                      isSelfBoard && canRearrangeOnOwnBoard
                        ? (card) => {
                            const c = card;
                            if (!isPropertyCard(c)) {
                              return;
                            }
                            setRearrangeCardId((current) =>
                              current &&
                              current.cardId === c.cardId &&
                              current.sourceSetId === propertySet.propertySetId
                                ? null
                                : {
                                    cardId: c.cardId,
                                    sourceSetId: propertySet.propertySetId,
                                  },
                            );
                          }
                        : undefined
                    }
                  />
                </div>
              </div>
            );
          })}

          {isSelfBoard && isSelfTurn ? (
            <div
              className="dnm-dropzone dnm-dropzone--property-new"
              onClick={() => {
                if (canRearrangeOnOwnBoard && rearrangeCardId) {
                  applyRearrange(undefined);
                  return;
                }
                if (canPlaceOnOwnBoard && selectedHandCardId) {
                  playSelectedAsProperty(selectedHandCardId, undefined);
                  setSelectedHandCardId(null);
                }
              }}
            >
              <div className="dnm-property-empty">
                Select a property, then tap here to start a new set.
              </div>
            </div>
          ) : null}
        </div>
      </article>
    );
  };

  // -------------------------------------------------------------------------
  // Overlays
  // -------------------------------------------------------------------------

  const renderDemandOverlay = () => {
    if (!hasAnyDemand) {
      return null;
    }
    return (
      <div className="dnm-demand-stack">
        {visibleDemands.map((demand) => {
          const isTargetingSelf =
            !!selfPlayerId && demand.playerId === selfPlayerId;
          const timerSeconds = demandTimerSecondsById[demand.id];
          const sourcePlayer = players.find(
            (p) => p.playerId === demand.sourceId,
          );
          const isSelectingThis =
            isSelectingPayment && activePaymentDemandId === demand.id;
          const eyebrow = demandEyebrow(demand.demandKind);
          const sourceName = sourcePlayer?.displayName ?? demand.sourceId;
          const selfMessage = selfDemandMessage(demand);
          const spectatorVerb = demand.isActive
            ? "is targeting"
            : "resolved a demand against";

          return (
            <aside
              key={demand.id}
              className="dnm-demand dnm-demand--payment"
              aria-live="polite"
              onPointerDown={(event) => event.stopPropagation()}
            >
              <p className="dnm-demand__eyebrow">{eyebrow}</p>
              {typeof timerSeconds === "number" ? (
                <p className="dnm-demand__line dnm-demand__timer">
                  Time left: {formatClock(timerSeconds)}
                </p>
              ) : null}
              {isTargetingSelf && demand.isActive ? (
                <div className="dnm-demand-source">
                  <img
                    className="dnm-demand-source__avatar"
                    src={sourcePlayer?.avatarUrl ?? ""}
                    alt={sourceName}
                    loading="lazy"
                    referrerPolicy="no-referrer"
                  />
                  <p className="dnm-demand__line dnm-demand-source__message">
                    <span className="dnm-demand-source__name">{sourceName}</span>{" "}
                    {selfMessage}
                    {isSelectingThis &&
                    demand.demandKind === DemandKind.DEMAND_KIND_PAYMENT
                      ? ` ($${selectedPaymentTotal}M selected)`
                      : ""}
                  </p>
                </div>
              ) : isTargetingSelf ? (
                <p className="dnm-demand__line dnm-demand-source__message dnm-demand__line--muted">
                  {inlinePlayer(demand.sourceId)} {selfMessage}
                </p>
              ) : (
                <p className="dnm-demand__line dnm-demand-source__message dnm-demand__line--muted">
                  {inlinePlayer(demand.sourceId)} {spectatorVerb}{" "}
                  {inlinePlayer(demand.playerId)}.
                </p>
              )}

              {isTargetingSelf ? (
                <div className="dnm-demand__actions">
                  <button
                    type="button"
                    className="dnm-demand__button"
                    onPointerDown={(event) => event.stopPropagation()}
                    onClick={() => complyDemand(demand)}
                    disabled={
                      isSelectingThis &&
                      demand.demandKind === DemandKind.DEMAND_KIND_PAYMENT &&
                      !canConfirmPayment
                    }
                  >
                    {demand.demandKind === DemandKind.DEMAND_KIND_PAYMENT
                      ? isSelectingThis
                        ? "Pay"
                        : "Select payment"
                      : demand.demandKind === DemandKind.DEMAND_KIND_REPO_MAN ||
                          demand.demandKind === DemandKind.DEMAND_KIND_TAX_DAY
                        ? "Choose"
                        : "OK"}
                  </button>
                  <button
                    type="button"
                    className="dnm-demand__button dnm-demand__button--secondary"
                    onPointerDown={(event) => event.stopPropagation()}
                    onClick={() => startDeny(demand.id)}
                    disabled={!hasNah}
                    title={hasNah ? "Play NAH!" : "Requires a NAH! card"}
                  >
                    NAH!
                  </button>
                </div>
              ) : null}
            </aside>
          );
        })}
      </div>
    );
  };

  // Debt-settlement overlay (blocks the board until each chip is settled).
  const renderDebtOverlay = () => {
    if (!hasOutstandingDebt) {
      return null;
    }
    const activeDebt = settleDebtId
      ? selfDebts.find((chip) => chip.id === settleDebtId)
      : selfDebts[0];
    return (
      <aside
        className="dnm-demand dnm-demand--payment"
        aria-live="polite"
        onPointerDown={(event) => event.stopPropagation()}
      >
        <p className="dnm-demand__eyebrow">Settle debt</p>
        <h3 className="dnm-demand__title">
          You owe {selfDebts.length} debt chip
          {selfDebts.length === 1 ? "" : "s"}
        </h3>
        {activeDebt ? (
          <p className="dnm-demand__line dnm-demand__line--muted">
            Pay {nameFor(activeDebt.creditorId)} with any one hand card (this
            consumes one play). Select a card from your hand.
          </p>
        ) : null}
        <p className="dnm-demand__line dnm-demand__line--muted">
          You must settle every chip before you can play.
        </p>
      </aside>
    );
  };

  // -------------------------------------------------------------------------
  // Hand card click routing (single place selection is toggled)
  // -------------------------------------------------------------------------

  const onHandCardClick = useCallback(
    (card: Card) => {
      // Debt settlement takes priority: tapping a hand card pays the next chip.
      if (hasOutstandingDebt) {
        const activeDebt = settleDebtId
          ? selfDebts.find((chip) => chip.id === settleDebtId)
          : selfDebts[0];
        if (activeDebt) {
          onSettleDebt(activeDebt.id, card.cardId);
        }
        return;
      }
      if (isSelectingPayment) {
        return;
      }
      if (isDiscardRequired) {
        onToggleDiscardCard(card.cardId);
        return;
      }
      if (!canSelectHandCards) {
        return;
      }
      setSelectedHandCardId((current) =>
        current === card.cardId ? null : card.cardId,
      );
    },
    [
      canSelectHandCards,
      hasOutstandingDebt,
      isDiscardRequired,
      isSelectingPayment,
      onSettleDebt,
      onToggleDiscardCard,
      selfDebts,
      settleDebtId,
    ],
  );

  const handSelectedIds = useMemo<Set<string>>(() => {
    if (isDiscardRequired) {
      return new Set(selectedDiscardCardIds);
    }
    if (selectedHandCardId) {
      return new Set([selectedHandCardId]);
    }
    return new Set();
  }, [isDiscardRequired, selectedDiscardCardIds, selectedHandCardId]);

  const handSelectable =
    hasOutstandingDebt || isDiscardRequired || canSelectHandCards;

  // -------------------------------------------------------------------------
  // Render
  // -------------------------------------------------------------------------

  if (!gameState) {
    return <p className="dnm-loading">Loading game...</p>;
  }

  const turnStatusLabel = isSelfTurn
    ? null
    : currentTurnPlayer
      ? `${currentTurnPlayer.displayName}'s turn`
      : "Waiting...";

  const passButtonLabel = isDiscardRequired
    ? `Submit Discard (${selectedDiscardCardIds.size}/${requiredDiscardCount})`
    : hasOutstandingDebt
      ? "Settle debt first"
      : "Pass Turn";

  return (
    <div
      className={[
        "dnm-board",
        layoutMode === "compact" ? "dnm-board--compact" : "",
      ]
        .filter(Boolean)
        .join(" ")}
    >
      <div className="dnm-demand-overlay-wrap">
        {hasOutstandingDebt
          ? renderDebtOverlay()
          : hasAnyDemand
            ? renderDemandOverlay()
            : isDiscardRequired
              ? (
                  <aside
                    className="dnm-demand dnm-demand--discard"
                    aria-live="polite"
                    onPointerDown={(event) => event.stopPropagation()}
                  >
                    <p className="dnm-demand__eyebrow">Discard required</p>
                    <h3 className="dnm-demand__title">Choose cards to discard</h3>
                    <p className="dnm-demand__line dnm-demand__line--muted">
                      Selected: {selectedDiscardCardIds.size}/
                      {requiredDiscardCount}. Pick from your hand, then submit.
                    </p>
                  </aside>
                )
              : null}
      </div>

      <section
        className={[
          "dnm-board-viewport",
          isPanning ? "is-panning" : "",
          shouldDimBoard ? "dnm-board-viewport--demand-active" : "",
        ]
          .filter(Boolean)
          .join(" ")}
        ref={viewportRef}
        onPointerDownCapture={onBoardPointerDown}
        onPointerMoveCapture={onBoardPointerMove}
        onPointerUpCapture={stopPan}
        onPointerCancelCapture={stopPan}
      >
        <div
          className={
            shouldDimBoard
              ? "dnm-board-content dnm-board-content--demand-active"
              : "dnm-board-content"
          }
          ref={contentRef}
          style={boardStyle}
        >
          {playerColumns.map((playerColumn, columnIndex) => (
            <div className="dnm-board-column" key={`col-${columnIndex}`}>
              {playerColumn.map((player) => renderPlayerBoard(player))}
            </div>
          ))}
        </div>
      </section>

      <section className="dnm-hand-row">
        <div
          className="dnm-dropzone dnm-dropzone--action"
          onClick={() => {
            if (canPlay && selectedHandCardId) {
              const card = getCardById(selectedHandCardId);
              // Only route non-property cards through the Actions pile; a
              // property tapped here means "bank it as money".
              playSelectedAsMoneyOrAction(selectedHandCardId);
              if (card && !isActionCard(card)) {
                setSelectedHandCardId(null);
              }
            }
          }}
        >
          <DealNoMercyCardStackBox
            title="Play / Bank"
            cards={gameState.lastAction ? [gameState.lastAction] : []}
            assetImageByKey={assetImageByKey}
            layout="spread"
            emptyLabel="Select a card, then tap here to play it."
          />
        </div>

        <div className="dnm-hand-main">
          <DealNoMercyCardStackBox
            title={
              hasOutstandingDebt
                ? "Pay a debt with a hand card"
                : isSelectingPayment
                  ? "Your hand"
                  : "Your hand"
            }
            cards={yourHand}
            assetImageByKey={assetImageByKey}
            layout="spread"
            emptyLabel="Waiting for cards"
            selectableCards={handSelectable}
            selectedCardIds={handSelectedIds}
            onCardClick={(card) => onHandCardClick(card)}
          />
        </div>

        <section className="dnm-hand-turn-controls">
          <button
            type="button"
            className="dnm-demand__button dnm-hand-turn-button"
            onClick={isDiscardRequired ? onSubmitDiscard : onPassTurn}
            disabled={
              !isSelfTurn ||
              hasOutstandingDebt ||
              (isDiscardRequired
                ? selectedDiscardCardIds.size !== requiredDiscardCount
                : false)
            }
          >
            <span className="dnm-hand-turn-button__content">
              {isSelfTurn ? (
                <>
                  <span className="dnm-hand-turn-button__action">
                    {passButtonLabel}
                  </span>
                  {!isDiscardRequired && !hasOutstandingDebt ? (
                    <span className="dnm-hand-turn-button__moves">
                      ({movesLeft} {movesLeft === 1 ? "move" : "moves"} left)
                    </span>
                  ) : null}
                </>
              ) : (
                <span className="dnm-hand-turn-button__status dnm-hand-turn-button__status--opponent">
                  {currentTurnPlayer?.avatarUrl ? (
                    <img
                      className="dnm-hand-turn-button__avatar"
                      src={currentTurnPlayer.avatarUrl}
                      alt={currentTurnPlayer.displayName}
                      loading="lazy"
                      referrerPolicy="no-referrer"
                    />
                  ) : null}
                  {turnStatusLabel}
                </span>
              )}
              {shouldShowTurnTimer ? (
                <span className="dnm-hand-turn-button__timer" aria-live="polite">
                  Time left: {formatClock(remainingTurnSeconds)}
                </span>
              ) : null}
            </span>
          </button>
        </section>
      </section>

      {/* Color picker (property wild / property raid / rent). */}
      {colorPicker ? (
        <div
          className="dnm-picker-backdrop"
          role="presentation"
          onClick={() => setColorPicker(null)}
        >
          <div
            className="dnm-picker"
            role="dialog"
            aria-modal="true"
            aria-label="Choose a color"
            onClick={(event) => event.stopPropagation()}
          >
            <p className="dnm-picker__eyebrow">
              {colorPicker.mode === "action_color"
                ? "Property Raid"
                : "Color selection"}
            </p>
            <p className="dnm-picker__title">
              {colorPicker.mode === "action_color"
                ? "Steal all properties of one color"
                : "Choose a color for this property"}
            </p>
            <div
              className="dnm-color-picker__swatches"
              style={{
                ["--dnm-color-cols" as string]: String(
                  Math.min(5, Math.max(1, colorPicker.colors.length)),
                ),
              }}
            >
              {colorPicker.colors.map((color) => (
                <button
                  type="button"
                  key={`${colorPicker.cardId}-${color}`}
                  className="dnm-color-picker__swatch"
                  style={{ background: colorHex(color) }}
                  aria-label={`Choose ${colorName(color)}`}
                  title={colorName(color)}
                  onClick={() => {
                    if (colorPicker.mode === "property_play") {
                      onPlayProperty(
                        colorPicker.cardId,
                        colorPicker.propertySetId,
                        color,
                      );
                    } else if (colorPicker.mode === "property_rearrange") {
                      onRearrangeCard(colorPicker.cardId, undefined, color);
                    } else {
                      onPlayPropertyRaid(colorPicker.cardId, color);
                    }
                    setColorPicker(null);
                    setSelectedHandCardId(null);
                  }}
                />
              ))}
            </div>
          </div>
        </div>
      ) : null}

      {/* Single-target player picker. */}
      {targetPicker ? (
        <div
          className="dnm-picker-backdrop"
          role="presentation"
          onClick={() => {
            setTargetPicker(null);
          }}
        >
          <aside
            className="dnm-picker"
            role="dialog"
            aria-modal="true"
            aria-label="Choose a player"
            onClick={(event) => event.stopPropagation()}
            onPointerDown={(event) => event.stopPropagation()}
          >
            <p className="dnm-picker__eyebrow">
              {targetPickerLabel(targetPicker.action)}
            </p>
            <p className="dnm-picker__title">Choose a player</p>
            {targetCandidates.length === 0 ? (
              <p className="dnm-picker__empty">No valid target available.</p>
            ) : (
              <div className="dnm-picker__list">
                {targetCandidates.map((player) => (
                  <button
                    key={player.playerId}
                    type="button"
                    className="dnm-picker__item"
                    onClick={() => onSelectTarget(player.playerId)}
                  >
                    <img
                      className="dnm-picker__avatar"
                      src={player.avatarUrl}
                      alt={player.displayName}
                      loading="lazy"
                      referrerPolicy="no-referrer"
                    />
                    <span className="dnm-picker__name">
                      {player.displayName}
                    </span>
                  </button>
                ))}
              </div>
            )}
          </aside>
        </div>
      ) : null}

      {/* Set Snatcher: pick a complete set from the chosen target. */}
      {setSnatcher ? (
        <div
          className="dnm-picker-backdrop"
          role="presentation"
          onClick={() => setSetSnatcher(null)}
        >
          <aside
            className="dnm-picker"
            role="dialog"
            aria-modal="true"
            aria-label="Choose a set to snatch"
            onClick={(event) => event.stopPropagation()}
            onPointerDown={(event) => event.stopPropagation()}
          >
            <p className="dnm-picker__eyebrow">Set Snatcher</p>
            <p className="dnm-picker__title">
              Steal a complete set from {nameFor(setSnatcher.targetId)}
            </p>
            {setSnatcherSets.length === 0 ? (
              <p className="dnm-picker__empty">No complete set available.</p>
            ) : (
              setSnatcherSets.map((set) => (
                <div key={set.propertySetId}>
                  <p className="dnm-picker__section-title">
                    {colorName(set.color)}
                    {set.shack ? " · Shack" : ""}
                  </p>
                  <div className="dnm-picker__card-list">
                    {set.cards.map((card) => (
                      <button
                        key={card.cardId}
                        type="button"
                        className="dnm-picker__card-item"
                        onClick={() => {
                          onPlaySetSnatcher(
                            setSnatcher.cardId,
                            setSnatcher.targetId,
                            set.propertySetId,
                          );
                          setSetSnatcher(null);
                          setSelectedHandCardId(null);
                        }}
                      >
                        <img
                          className="dnm-picker__card-image"
                          src={assetImageByKey[card.assetKey]}
                          alt={String(card.assetKey)}
                          loading="lazy"
                          referrerPolicy="no-referrer"
                        />
                      </button>
                    ))}
                  </div>
                </div>
              ))
            )}
          </aside>
        </div>
      ) : null}

      {/* Pickpocket: choose the category to steal from the target's hand. */}
      {pickpocket ? (
        <div
          className="dnm-picker-backdrop"
          role="presentation"
          onClick={() => setPickpocket(null)}
        >
          <aside
            className="dnm-picker"
            role="dialog"
            aria-modal="true"
            aria-label="Choose a category to steal"
            onClick={(event) => event.stopPropagation()}
            onPointerDown={(event) => event.stopPropagation()}
          >
            <p className="dnm-picker__eyebrow">Pickpocket</p>
            <p className="dnm-picker__title">
              Steal a category from {nameFor(pickpocket.targetId)}&apos;s hand
            </p>
            <div className="dnm-picker__list">
              {(
                [
                  ["Property", StealCategory.STEAL_CATEGORY_PROPERTY],
                  ["Money", StealCategory.STEAL_CATEGORY_MONEY],
                  ["Action", StealCategory.STEAL_CATEGORY_ACTION],
                ] as const
              ).map(([label, category]) => (
                <button
                  key={label}
                  type="button"
                  className="dnm-picker__item"
                  onClick={() => {
                    onPlayPickpocket(pickpocket.cardId, pickpocket.targetId, category);
                    setPickpocket(null);
                    setSelectedHandCardId(null);
                  }}
                >
                  <span className="dnm-picker__name">{label}</span>
                </button>
              ))}
            </div>
          </aside>
        </div>
      ) : null}

      {/* Shack: pick which of your own sets to attach it to. */}
      {shackPick ? (
        <div
          className="dnm-picker-backdrop"
          role="presentation"
          onClick={() => setShackPick(null)}
        >
          <aside
            className="dnm-picker"
            role="dialog"
            aria-modal="true"
            aria-label="Choose a set for the shack"
            onClick={(event) => event.stopPropagation()}
            onPointerDown={(event) => event.stopPropagation()}
          >
            <p className="dnm-picker__eyebrow">Shack</p>
            <p className="dnm-picker__title">
              Attach +rent to one of your sets
            </p>
            <p className="dnm-picker__hint">
              Any set works (including incomplete); one shack per set.
            </p>
            <div className="dnm-picker__list">
              {selfPropertySets.map((set) => {
                const disabled = !!set.shack;
                return (
                  <button
                    key={set.propertySetId}
                    type="button"
                    className={[
                      "dnm-picker__item",
                      disabled ? "dnm-picker__item--disabled" : "",
                    ]
                      .filter(Boolean)
                      .join(" ")}
                    disabled={disabled}
                    onClick={() => {
                      onPlayShack(shackPick.cardId, set.propertySetId);
                      setShackPick(null);
                      setSelectedHandCardId(null);
                    }}
                  >
                    <span className="dnm-picker__name">
                      {colorName(set.color)} · {set.cards.length} card
                      {set.cards.length === 1 ? "" : "s"}
                      {disabled ? " (has shack)" : ""}
                    </span>
                  </button>
                );
              })}
            </div>
          </aside>
        </div>
      ) : null}

      {/* Heist / Market Crash: one pick per opponent, single confirm. */}
      {multiPick ? (
        <div
          className="dnm-picker-backdrop"
          role="presentation"
          onClick={() => {
            setMultiPick(null);
          }}
        >
          <aside
            className="dnm-picker"
            role="dialog"
            aria-modal="true"
            aria-label="Choose one card from each opponent"
            onClick={(event) => event.stopPropagation()}
            onPointerDown={(event) => event.stopPropagation()}
          >
            <p className="dnm-picker__eyebrow">
              {multiPick.action === "heist" ? "Heist" : "Market Crash"}
            </p>
            <p className="dnm-picker__title">
              {multiPick.action === "heist"
                ? "Take one bank card from every opponent"
                : "Take one property from every opponent"}
            </p>
            {multiPickTargets.length === 0 ? (
              <p className="dnm-picker__empty">
                No opponent has a{" "}
                {multiPick.action === "heist" ? "bank card" : "property"} to
                take.
              </p>
            ) : (
              <>
                {multiPickTargets.map((entry) => (
                  <div key={entry.player.playerId}>
                    <p className="dnm-picker__section-title">
                      {entry.player.displayName}
                    </p>
                    <div className="dnm-picker__card-list">
                      {entry.cards.map((card) => {
                        const active =
                          multiPick.picks[entry.player.playerId] === card.cardId;
                        return (
                          <button
                            key={card.cardId}
                            type="button"
                            className={[
                              "dnm-picker__card-item",
                              active ? "dnm-picker__card-item--active" : "",
                            ]
                              .filter(Boolean)
                              .join(" ")}
                            onClick={() =>
                              setMultiPick((current) =>
                                current
                                  ? {
                                      ...current,
                                      picks: {
                                        ...current.picks,
                                        [entry.player.playerId]: card.cardId,
                                      },
                                    }
                                  : current,
                              )
                            }
                          >
                            <img
                              className="dnm-picker__card-image"
                              src={assetImageByKey[card.assetKey]}
                              alt={String(card.assetKey)}
                              loading="lazy"
                              referrerPolicy="no-referrer"
                            />
                          </button>
                        );
                      })}
                    </div>
                  </div>
                ))}
                <p className="dnm-picker__multi-note">
                  Pick exactly one card from each opponent listed.
                </p>
                <div className="dnm-picker__actions">
                  <button
                    type="button"
                    className="dnm-demand__button"
                    disabled={!multiPickComplete}
                    onClick={submitMultiPick}
                  >
                    Confirm
                  </button>
                  <button
                    type="button"
                    className="dnm-demand__button dnm-demand__button--secondary"
                    onClick={() => setMultiPick(null)}
                  >
                    Cancel
                  </button>
                </div>
              </>
            )}
          </aside>
        </div>
      ) : null}

      {/* Repo Man / Tax Day distribution: keep one, hand out the rest. */}
      {distribution ? (
        <div
          className="dnm-picker-backdrop"
          role="presentation"
          onClick={() => setDistribution(null)}
        >
          <aside
            className="dnm-picker"
            role="dialog"
            aria-modal="true"
            aria-label="Distribute your cards"
            onClick={(event) => event.stopPropagation()}
            onPointerDown={(event) => event.stopPropagation()}
          >
            <p className="dnm-picker__eyebrow">
              {distribution.kind === "repo_man" ? "Repo Man" : "Tax Day"}
            </p>
            <p className="dnm-picker__title">
              Keep one {distribution.kind === "repo_man" ? "property" : "bank card"},
              give the rest away
            </p>
            {distributionCards.length === 0 ? (
              <p className="dnm-picker__empty">Nothing to distribute.</p>
            ) : (
              <>
                <div className="dnm-distribution-list">
                  {distributionCards.map((card) => {
                    const isKept = distribution.keepCardId === card.cardId;
                    const assignedTo = distribution.assignments[card.cardId];
                    return (
                      <div key={card.cardId} className="dnm-distribution-row">
                        <div className="dnm-distribution-row__card">
                          <img
                            className="dnm-distribution-row__card-image"
                            src={assetImageByKey[card.assetKey]}
                            alt={String(card.assetKey)}
                            loading="lazy"
                            referrerPolicy="no-referrer"
                          />
                          <button
                            type="button"
                            className={[
                              "dnm-chip-toggle",
                              "dnm-chip-toggle--keep",
                              isKept ? "dnm-chip-toggle--active" : "",
                            ]
                              .filter(Boolean)
                              .join(" ")}
                            onClick={() =>
                              setDistribution((current) =>
                                current
                                  ? {
                                      ...current,
                                      keepCardId: isKept
                                        ? undefined
                                        : card.cardId,
                                      assignments: isKept
                                        ? current.assignments
                                        : (() => {
                                            const next = {
                                              ...current.assignments,
                                            };
                                            delete next[card.cardId];
                                            return next;
                                          })(),
                                    }
                                  : current,
                              )
                            }
                          >
                            Keep
                          </button>
                        </div>
                        {!isKept ? (
                          <div className="dnm-distribution-row__recipients">
                            {distributionRecipients.map((recipient) => (
                              <button
                                key={recipient.playerId}
                                type="button"
                                className={[
                                  "dnm-chip-toggle",
                                  assignedTo === recipient.playerId
                                    ? "dnm-chip-toggle--active"
                                    : "",
                                ]
                                  .filter(Boolean)
                                  .join(" ")}
                                onClick={() =>
                                  setDistribution((current) =>
                                    current
                                      ? {
                                          ...current,
                                          assignments: {
                                            ...current.assignments,
                                            [card.cardId]: recipient.playerId,
                                          },
                                        }
                                      : current,
                                  )
                                }
                              >
                                {nameFor(recipient.playerId)}
                              </button>
                            ))}
                          </div>
                        ) : null}
                      </div>
                    );
                  })}
                </div>
                <div className="dnm-picker__actions">
                  <button
                    type="button"
                    className="dnm-demand__button"
                    disabled={!distributionComplete}
                    onClick={submitDistribution}
                  >
                    Confirm
                  </button>
                  <button
                    type="button"
                    className="dnm-demand__button dnm-demand__button--secondary"
                    onClick={() => setDistribution(null)}
                  >
                    Cancel
                  </button>
                </div>
              </>
            )}
          </aside>
        </div>
      ) : null}

      {/* NAH! deny with multiple NAH! cards held (pick which to use). */}
      {denyState ? (
        <div
          className="dnm-picker-backdrop"
          role="presentation"
          onClick={() => setDenyState(null)}
        >
          <aside
            className="dnm-picker"
            role="dialog"
            aria-modal="true"
            aria-label="Play a NAH! card"
            onClick={(event) => event.stopPropagation()}
            onPointerDown={(event) => event.stopPropagation()}
          >
            <p className="dnm-picker__eyebrow">NAH!</p>
            <p className="dnm-picker__title">Deny this demand</p>
            <div className="dnm-picker__card-list">
              {denyState.nahCardIds.map((cardId) => {
                const card = getCardById(cardId);
                return (
                  <button
                    key={cardId}
                    type="button"
                    className="dnm-picker__card-item"
                    onClick={() => {
                      onDenyDemand(denyState.demandId, cardId);
                      setDenyState(null);
                    }}
                  >
                    <img
                      className="dnm-picker__card-image"
                      src={card ? assetImageByKey[card.assetKey] : ""}
                      alt="NAH!"
                      loading="lazy"
                      referrerPolicy="no-referrer"
                    />
                  </button>
                );
              })}
            </div>
          </aside>
        </div>
      ) : null}
    </div>
  );
};

// ---------------------------------------------------------------------------
// Demand copy helpers
// ---------------------------------------------------------------------------

const demandEyebrow = (kind: DemandKind): string => {
  switch (kind) {
    case DemandKind.DEMAND_KIND_PAYMENT:
      return "Payment";
    case DemandKind.DEMAND_KIND_PROPERTY:
      return "Market Crash";
    case DemandKind.DEMAND_KIND_PROPERTY_SET:
      return "Set Snatcher";
    case DemandKind.DEMAND_KIND_COLOR_PROPERTIES:
      return "Property Raid";
    case DemandKind.DEMAND_KIND_BANK_CARD:
      return "Heist";
    case DemandKind.DEMAND_KIND_BANK_SWAP:
      return "Bank Swap";
    case DemandKind.DEMAND_KIND_DEBT_TRAP:
      return "Debt Trap";
    case DemandKind.DEMAND_KIND_REPO_MAN:
      return "Repo Man";
    case DemandKind.DEMAND_KIND_TAX_DAY:
      return "Tax Day";
    case DemandKind.DEMAND_KIND_PICKPOCKET:
      return "Pickpocket";
    default:
      return "Demand";
  }
};

// The message portion of a demand that targets you. The leading actor (the
// source player) is rendered separately (avatar + name) by the caller, so this
// returns only the trailing clause.
const selfDemandMessage = (demand: Demand): string => {
  const amount = demand.paymentDemand?.amount;
  const amountLabel = typeof amount === "number" ? `$${amount}M` : "";
  switch (demand.demandKind) {
    case DemandKind.DEMAND_KIND_PAYMENT:
      return demand.isActive
        ? `wants you to pay ${amountLabel}.`
        : `— you responded.`;
    case DemandKind.DEMAND_KIND_PROPERTY:
      return "is crashing the market — hand over the chosen property.";
    case DemandKind.DEMAND_KIND_PROPERTY_SET:
      return "is snatching one of your complete sets.";
    case DemandKind.DEMAND_KIND_COLOR_PROPERTIES:
      return `is raiding your ${demand.colorPropertiesDemand ? colorLabel(demand.colorPropertiesDemand.color) : ""} properties.`;
    case DemandKind.DEMAND_KIND_BANK_CARD:
      return "is heisting a card from your bank.";
    case DemandKind.DEMAND_KIND_BANK_SWAP:
      return "wants to swap banks with you.";
    case DemandKind.DEMAND_KIND_DEBT_TRAP:
      return "is setting a debt trap on you.";
    case DemandKind.DEMAND_KIND_REPO_MAN:
      return "repossessed your properties — keep one, give the rest away.";
    case DemandKind.DEMAND_KIND_TAX_DAY:
      return "is taxing your bank — keep one, give the rest away.";
    case DemandKind.DEMAND_KIND_PICKPOCKET:
      return `is pickpocketing your ${stealCategoryLabel(demand.pickpocketDemand?.category)} cards.`;
    default:
      return "wants something from you.";
  }
};

const colorLabel = (color: Color): string => {
  const raw = colorToJSON(color).replace("COLOR_", "").toLowerCase();
  return raw.charAt(0).toUpperCase() + raw.slice(1);
};

const stealCategoryLabel = (category: StealCategory | undefined): string => {
  switch (category) {
    case StealCategory.STEAL_CATEGORY_PROPERTY:
      return "property";
    case StealCategory.STEAL_CATEGORY_MONEY:
      return "money";
    case StealCategory.STEAL_CATEGORY_ACTION:
      return "action";
    default:
      return "";
  }
};

const targetPickerLabel = (action: TargetPickerState["action"]): string => {
  switch (action) {
    case "yoink":
      return "Yoink!";
    case "bank_swap":
      return "Bank Swap";
    case "debt_trap":
      return "Debt Trap";
    case "repo_man":
      return "Repo Man";
    case "tax_day":
      return "Tax Day";
    case "set_snatcher":
      return "Set Snatcher";
    case "pickpocket":
      return "Pickpocket";
    default:
      return "Choose target";
  }
};

export default DealNoMercyBoard;
