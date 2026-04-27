import {
  useCallback,
  useEffect,
  useMemo,
  useRef,
  useState,
  type CSSProperties,
  type MouseEvent,
  type PointerEvent,
} from "react";
import {
  AssetKey,
  Category,
  Color,
  DemandKind,
  type Demand,
  type Card,
  type GameState,
  type PropertySet,
  type Player,
} from "../../../generated/monopoly_deal";
import DebtCollectorTargetPickerOverlay from "./action/DebtCollectorTargetPickerOverlay";
import DiscardPromptOverlay from "./discard/DiscardPromptOverlay";
import GameCardStackBox from "./GameCardStackBox";
import DemandOverlay from "./demand/DemandOverlay";
import PendingRentOverlay from "./rent/PendingRentOverlay";
import "./monopoly-deal-html.css";

type MonopolyDealHtmlBoardProps = {
  gameState: GameState | null;
  assetImageByKey: Record<number, string>;
  selfPlayerId?: string;
  onPlayMoneyCard: (cardId: string) => void;
  onPlayPassGoCard: (cardId: string) => void;
  onPlayDebtCollectorCard: (cardId: string, targetPlayerId: string) => void;
  onPlayWildRentCard: (cardId: string, targetPlayerId: string) => void;
  onPlaySlyDealCard: (
    cardId: string,
    targetPlayerId: string,
    targetCardId: string,
  ) => void;
  onPlayDealBreakerCard: (
    cardId: string,
    targetPlayerId: string,
    propertySetId: string,
  ) => void;
  onPlayForcedDealCard: (
    cardId: string,
    targetPlayerId: string,
    sourceCardId: string,
    targetCardId: string,
  ) => void;
  onPlayDoubleTheRentCard: () => void;
  onResolvePendingRent: () => void;
  onRearrangeCard: (
    cardId: string,
    propertySetId?: string,
    color?: Color,
  ) => void;
  isDiscardRequired: boolean;
  requiredDiscardCount: number;
  selectedDiscardCardIds: ReadonlySet<string>;
  onToggleDiscardCard: (cardId: string) => void;
  onPlayPropertyCard: (
    cardId: string,
    propertySetId?: string,
    activeColor?: Color,
  ) => void;
  onComplyPaymentDemand: (demandId: string, cardIds: string[]) => void;
  onComplyPropertyDemand: (demandId: string) => void;
  onComplyPropertySetDemand: (demandId: string) => void;
  onDenyDemand: (demandId: string) => void;
  onPassTurn: () => void;
  onSubmitDiscard: () => void;
  onClientError?: (error: unknown) => void;
  onGameError?: (error: { message: string; code: string }) => void;
};

type Size = {
  width: number;
  height: number;
};

type Pan = {
  x: number;
  y: number;
};

type ColorPickerState = {
  mode: "play" | "rearrange";
  cardId: string;
  propertySetId?: string;
  sourcePropertySetId?: string;
  colors: Color[];
};

type DebtCollectorPickerState = {
  mode: "debt_collector" | "wild_rent";
  cardId: string;
};

type DealPickerState = {
  mode: "sly_deal" | "forced_deal" | "deal_breaker";
  cardId: string;
  selectedOpponentCardId?: string;
  selectedOpponentPlayerId?: string;
  selectedTargetPropertySetId?: string;
  selectedSourceCardId?: string;
  isOpponentSelectionConfirmed?: boolean;
};

type RearrangeSelectionState = {
  cardId: string;
  sourcePropertySetId: string;
};

const MIN_ZOOM = 0.65;
const MAX_ZOOM = 2.2;
const BOARD_COLUMN_GAP_REM = 0.7;

const computeBoardGrid = (playerCount: number): { columns: number } => {
  if (playerCount <= 1) {
    return { columns: 1 };
  }

  if (playerCount === 2) {
    return { columns: 1 };
  }

  const columns = Math.ceil(Math.sqrt(playerCount));
  return { columns };
};

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

const isPropertyCard = (card: Card): boolean => {
  return (
    card.category === Category.CATEGORY_PURE_PROPERTY ||
    card.category === Category.CATEGORY_WILD_PROPERTY
  );
};

const countPropertyCards = (propertySet: PropertySet): number => {
  return propertySet.cards.filter((card) => isPropertyCard(card)).length;
};

const isPropertySetComplete = (propertySet: PropertySet): boolean => {
  const minRequiredCount = minPropertyCountForCompleteSet(propertySet.color);
  if (!Number.isFinite(minRequiredCount)) {
    return false;
  }

  return countPropertyCards(propertySet) >= minRequiredCount;
};

const isPropertySetLocked = (propertySet: PropertySet): boolean => {
  return propertySet.cards.some((card) => {
    return (
      card.assetKey === AssetKey.ASSET_KEY_HOUSE ||
      card.assetKey === AssetKey.ASSET_KEY_HOTEL
    );
  });
};

const BASE_RENT_BY_COLOR: Partial<Record<Color, Record<number, number>>> = {
  [Color.COLOR_UNSPECIFIED]: { 0: 0 },
  [Color.COLOR_BROWN]: { 1: 1, 2: 2 },
  [Color.COLOR_SKY]: { 1: 1, 2: 2, 3: 3 },
  [Color.COLOR_PINK]: { 1: 1, 2: 2, 3: 4 },
  [Color.COLOR_ORANGE]: { 1: 1, 2: 3, 3: 5 },
  [Color.COLOR_RED]: { 1: 2, 2: 3, 3: 6 },
  [Color.COLOR_YELLOW]: { 1: 2, 2: 4, 3: 6 },
  [Color.COLOR_GREEN]: { 1: 2, 2: 4, 3: 7 },
  [Color.COLOR_BLUE]: { 1: 3, 2: 8 },
  [Color.COLOR_UTILITY]: { 1: 1, 2: 2 },
  [Color.COLOR_RAILROAD]: { 1: 1, 2: 2, 3: 3, 4: 4 },
};

const getBaseRentForPropertySet = (propertySet: PropertySet): number => {
  const rentByCount = BASE_RENT_BY_COLOR[propertySet.color];
  if (!rentByCount) {
    return 0;
  }

  const propertyCount = countPropertyCards(propertySet);
  if (propertyCount <= 0) {
    return 0;
  }

  const direct = rentByCount[propertyCount];
  if (typeof direct === "number") {
    return direct;
  }

  const maxCount = Math.max(...Object.keys(rentByCount).map(Number));
  return rentByCount[maxCount] ?? 0;
};

const getRentBonusLabels = (propertySet: PropertySet): string[] => {
  const hasHouse = propertySet.cards.some((card) => {
    return card.assetKey === AssetKey.ASSET_KEY_HOUSE;
  });
  const hasHotel = propertySet.cards.some((card) => {
    return card.assetKey === AssetKey.ASSET_KEY_HOTEL;
  });

  const labels: string[] = [];
  if (hasHouse) {
    labels.push("+3M");
  }
  if (hasHotel) {
    labels.push("+4M");
  }

  return labels;
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

  if (scaledWidth <= viewportSize.width) {
    nextX = (viewportSize.width - scaledWidth) / 2;
  } else {
    const minX = viewportSize.width - scaledWidth;
    nextX = Math.min(0, Math.max(minX, nextX));
  }

  if (scaledHeight <= viewportSize.height) {
    nextY = (viewportSize.height - scaledHeight) / 2;
  } else {
    const minY = viewportSize.height - scaledHeight;
    nextY = Math.min(0, Math.max(minY, nextY));
  }

  return { x: nextX, y: nextY };
};

const distanceBetweenPoints = (
  a: { x: number; y: number },
  b: { x: number; y: number },
): number => {
  const dx = b.x - a.x;
  const dy = b.y - a.y;
  return Math.hypot(dx, dy);
};

const midpointBetweenPoints = (
  a: { x: number; y: number },
  b: { x: number; y: number },
): { x: number; y: number } => {
  return {
    x: (a.x + b.x) / 2,
    y: (a.y + b.y) / 2,
  };
};

const MonopolyDealHtmlBoard = ({
  gameState,
  assetImageByKey,
  selfPlayerId,
  onPlayMoneyCard,
  onPlayPassGoCard,
  onPlayDebtCollectorCard,
  onPlayWildRentCard,
  onPlaySlyDealCard,
  onPlayDealBreakerCard,
  onPlayForcedDealCard,
  onPlayDoubleTheRentCard,
  onResolvePendingRent,
  onRearrangeCard,
  onDenyDemand,
  isDiscardRequired,
  requiredDiscardCount,
  selectedDiscardCardIds,
  onToggleDiscardCard,
  onPlayPropertyCard,
  onComplyPaymentDemand,
  onComplyPropertyDemand,
  onComplyPropertySetDemand,
  onClientError,
  onGameError,
  onPassTurn,
  onSubmitDiscard,
}: MonopolyDealHtmlBoardProps) => {
  const viewportRef = useRef<HTMLDivElement | null>(null);
  const contentRef = useRef<HTMLDivElement | null>(null);
  const [zoom, setZoom] = useState(1);
  const [pan, setPan] = useState<Pan>({ x: 0, y: 0 });
  const [viewportSize, setViewportSize] = useState<Size>({
    width: 1,
    height: 1,
  });
  const [contentSize, setContentSize] = useState<Size>({ width: 1, height: 1 });
  const [isPanning, setIsPanning] = useState(false);
  const [colorPicker, setColorPicker] = useState<ColorPickerState | null>(null);
  const [debtCollectorPicker, setDebtCollectorPicker] =
    useState<DebtCollectorPickerState | null>(null);
  const [dealPicker, setDealPicker] = useState<DealPickerState | null>(null);
  const [isSelectingPaymentCards, setIsSelectingPaymentCards] = useState(false);
  const [selectedPaymentCardIds, setSelectedPaymentCardIds] = useState<
    Set<string>
  >(() => new Set());
  const [activePaymentDemandId, setActivePaymentDemandId] = useState<
    string | null
  >(null);
  const [selectedHandPlacementCardId, setSelectedHandPlacementCardId] =
    useState<string | null>(null);
  const [selectedRearrangeCard, setSelectedRearrangeCard] =
    useState<RearrangeSelectionState | null>(null);
  const panPointerIdRef = useRef<number | null>(null);
  const panStartPointRef = useRef<{ x: number; y: number } | null>(null);
  const didPanDuringPointerRef = useRef(false);
  const suppressClickUntilRef = useRef(0);
  const activeTouchPointsRef = useRef<Map<number, { x: number; y: number }>>(
    new Map(),
  );
  const pinchStartDistanceRef = useRef<number | null>(null);
  const pinchStartZoomRef = useRef(1);
  const pinchWorldCenterRef = useRef<{ x: number; y: number } | null>(null);
  const autoResolvedPendingRentKeyRef = useRef<string | null>(null);

  const players = gameState?.players ?? [];

  const orderedPlayers = useMemo(() => {
    if (!selfPlayerId) {
      return players;
    }

    const selfIndex = players.findIndex(
      (player) => player.playerId === selfPlayerId,
    );
    if (selfIndex <= 0) {
      return players;
    }

    const nextPlayers = [...players];
    const [selfPlayer] = nextPlayers.splice(selfIndex, 1);
    if (selfPlayer) {
      nextPlayers.unshift(selfPlayer);
    }

    return nextPlayers;
  }, [players, selfPlayerId]);
  const visibleDemands = useMemo<Demand[]>(() => {
    const demands = gameState?.demands ?? [];
    const selfDemands = selfPlayerId
      ? demands.filter((demand) => demand.playerId === selfPlayerId)
      : demands;
    if (selfDemands.length > 0) {
      return selfDemands.slice(0, 3);
    }

    return demands.slice(0, 2);
  }, [gameState?.demands, selfPlayerId]);

  const hasAnyDemand = visibleDemands.length > 0;
  const pendingRent = gameState?.pendingRent;
  const hasPendingRent = !!pendingRent;
  const currentPlayerId = gameState?.currentPlayerId ?? "";
  const isSelfTurn = !!selfPlayerId && selfPlayerId === currentPlayerId;
  const yourHand = gameState?.yourHand?.cards ?? [];
  const hasJustSayNo = yourHand.some(
    (card) => card.assetKey === AssetKey.ASSET_KEY_JUST_SAY_NO,
  );
  const hasDoubleTheRent = yourHand.some(
    (card) => card.assetKey === AssetKey.ASSET_KEY_DOUBLE_THE_RENT,
  );
  const shouldAutoResolvePendingRent =
    !!pendingRent &&
    (!selfPlayerId || pendingRent.playerId === selfPlayerId) &&
    !hasDoubleTheRent;
  const shouldShowPendingRentOverlay =
    hasPendingRent && !shouldAutoResolvePendingRent;
  const hasMovesLeft = (gameState?.movesLeft ?? 0) > 0;
  const movesLeft = gameState?.movesLeft ?? 0;
  const lastActionCards =
    gameState?.lastAction &&
    gameState.lastAction.assetKey !== AssetKey.ASSET_KEY_UNSPECIFIED
      ? [gameState.lastAction]
      : [];

  const moneyByPlayer = useMemo(() => {
    const lookup: Record<string, Card[]> = {};
    for (const pile of gameState?.money ?? []) {
      lookup[pile.playerId] = pile.cards;
    }
    return lookup;
  }, [gameState?.money]);

  const propertySetsByPlayer = useMemo(() => {
    const lookup: Record<string, GameState["properties"]> = {};
    for (const propertySet of gameState?.properties ?? []) {
      if (!lookup[propertySet.playerId]) {
        lookup[propertySet.playerId] = [];
      }
      lookup[propertySet.playerId].push(propertySet);
    }
    return lookup;
  }, [gameState?.properties]);

  const propertyCardImageById = useMemo(() => {
    const lookup: Record<string, string> = {};
    for (const propertySet of gameState?.properties ?? []) {
      for (const card of propertySet.cards) {
        const imageUrl = assetImageByKey[card.assetKey];
        if (imageUrl) {
          lookup[card.cardId] = imageUrl;
        }
      }
    }

    return lookup;
  }, [assetImageByKey, gameState?.properties]);

  const propertySetCardsById = useMemo(() => {
    const lookup: Record<string, Card[]> = {};
    for (const propertySet of gameState?.properties ?? []) {
      lookup[propertySet.propertySetId] = propertySet.cards;
    }
    return lookup;
  }, [gameState?.properties]);

  const selfPropertySets = useMemo(() => {
    if (!selfPlayerId) {
      return [] as PropertySet[];
    }

    return propertySetsByPlayer[selfPlayerId] ?? [];
  }, [propertySetsByPlayer, selfPlayerId]);

  const selectableBoardCardsById = useMemo(() => {
    const lookup = new Map<string, Card>();
    if (!selfPlayerId) {
      return lookup;
    }

    const moneyCards = moneyByPlayer[selfPlayerId] ?? [];
    for (const card of moneyCards) {
      lookup.set(card.cardId, card);
    }

    const propertySets = propertySetsByPlayer[selfPlayerId] ?? [];
    for (const propertySet of propertySets) {
      for (const card of propertySet.cards) {
        lookup.set(card.cardId, card);
      }
    }

    return lookup;
  }, [moneyByPlayer, propertySetsByPlayer, selfPlayerId]);

  const selectedPaymentTotal = useMemo(() => {
    let total = 0;
    for (const cardId of selectedPaymentCardIds) {
      total += selectableBoardCardsById.get(cardId)?.value ?? 0;
    }
    return total;
  }, [selectableBoardCardsById, selectedPaymentCardIds]);

  const selectedPaymentDemand = useMemo(() => {
    if (!activePaymentDemandId) {
      return undefined;
    }

    return visibleDemands.find((demand) => demand.id === activePaymentDemandId);
  }, [activePaymentDemandId, visibleDemands]);

  const isSelectedPaymentDemandActive =
    selectedPaymentDemand?.demandKind === DemandKind.DEMAND_KIND_PAYMENT &&
    selectedPaymentDemand?.isActive === true;
  const shouldShowPaymentMoneyPicker =
    isSelectingPaymentCards && isSelectedPaymentDemandActive;
  const selfMoneyCards = useMemo(() => {
    if (!selfPlayerId) {
      return [] as Card[];
    }

    return moneyByPlayer[selfPlayerId] ?? [];
  }, [moneyByPlayer, selfPlayerId]);

  const isSelectingDealCards = !!dealPicker;
  const canSelectHandCards =
    !isDiscardRequired &&
    !isSelectingPaymentCards &&
    !isSelectingDealCards &&
    !colorPicker &&
    !debtCollectorPicker;
  const canTapPlaceFromHand =
    isSelfTurn &&
    canSelectHandCards;

  const shouldDimBoard =
    isDiscardRequired ||
    shouldShowPendingRentOverlay ||
    (hasAnyDemand &&
      !(isSelectedPaymentDemandActive && isSelectingPaymentCards));

  const paymentAmount = selectedPaymentDemand?.paymentDemand?.amount ?? 0;
  const hasSelectedAllBoardCards = useMemo(() => {
    for (const cardId of selectableBoardCardsById.keys()) {
      if (!selectedPaymentCardIds.has(cardId)) {
        return false;
      }
    }

    return true;
  }, [selectableBoardCardsById, selectedPaymentCardIds]);

  const canConfirmPaymentSelection =
    selectedPaymentTotal >= paymentAmount || hasSelectedAllBoardCards;

  const stealableCardOwnerById = useMemo(() => {
    const lookup = new Map<string, string>();

    for (const [playerId, propertySets] of Object.entries(
      propertySetsByPlayer,
    )) {
      if (selfPlayerId && playerId === selfPlayerId) {
        continue;
      }

      for (const propertySet of propertySets) {
        if (isPropertySetComplete(propertySet)) {
          continue;
        }

        for (const card of propertySet.cards) {
          if (isPropertyCard(card)) {
            lookup.set(card.cardId, playerId);
          }
        }
      }
    }

    return lookup;
  }, [propertySetsByPlayer, selfPlayerId]);

  const forcedDealSourceCardIds = useMemo(() => {
    const selectableIds = new Set<string>();

    for (const propertySet of selfPropertySets) {
      if (isPropertySetComplete(propertySet)) {
        continue;
      }

      for (const card of propertySet.cards) {
        if (isPropertyCard(card)) {
          selectableIds.add(card.cardId);
        }
      }
    }

    return selectableIds;
  }, [selfPropertySets]);

  const stealableCompleteSetOwnerById = useMemo(() => {
    const lookup = new Map<string, string>();

    for (const [playerId, propertySets] of Object.entries(
      propertySetsByPlayer,
    )) {
      if (selfPlayerId && playerId === selfPlayerId) {
        continue;
      }

      for (const propertySet of propertySets) {
        if (isPropertySetComplete(propertySet)) {
          lookup.set(propertySet.propertySetId, playerId);
        }
      }
    }

    return lookup;
  }, [propertySetsByPlayer, selfPlayerId]);

  const selectedDealCardIds = useMemo(() => {
    const selectedIds = new Set<string>();

    if (dealPicker?.selectedOpponentCardId) {
      selectedIds.add(dealPicker.selectedOpponentCardId);
    }

    if (dealPicker?.selectedSourceCardId) {
      selectedIds.add(dealPicker.selectedSourceCardId);
    }

    if (dealPicker?.selectedTargetPropertySetId) {
      for (const propertySet of gameState?.properties ?? []) {
        if (
          propertySet.propertySetId !== dealPicker.selectedTargetPropertySetId
        ) {
          continue;
        }

        for (const card of propertySet.cards) {
          selectedIds.add(card.cardId);
        }
        break;
      }
    }

    return selectedIds;
  }, [
    dealPicker?.selectedOpponentCardId,
    dealPicker?.selectedSourceCardId,
    dealPicker?.selectedTargetPropertySetId,
    gameState?.properties,
  ]);

  const selectedHandPlacementCardIds = useMemo(() => {
    if (!selectedHandPlacementCardId) {
      return new Set<string>();
    }

    return new Set([selectedHandPlacementCardId]);
  }, [selectedHandPlacementCardId]);

  const selectedRearrangeCardIds = useMemo(() => {
    if (!selectedRearrangeCard) {
      return new Set<string>();
    }

    return new Set([selectedRearrangeCard.cardId]);
  }, [selectedRearrangeCard]);

  const selectedDealPropertySetIds = useMemo(() => {
    const selectedIds = new Set<string>();

    if (dealPicker?.selectedTargetPropertySetId) {
      selectedIds.add(dealPicker.selectedTargetPropertySetId);
    }

    return selectedIds;
  }, [dealPicker?.selectedTargetPropertySetId]);

  const { columns } = useMemo(
    () => computeBoardGrid(orderedPlayers.length),
    [orderedPlayers.length],
  );

  const boardColumnWidth = useMemo(() => {
    const viewportWidth = Math.max(viewportSize.width, 1);
    return Math.min(1000, viewportWidth);
  }, [viewportSize.width]);

  const boardColumnGapPx = useMemo(() => {
    const rootFontSize = Number.parseFloat(
      window.getComputedStyle(document.documentElement).fontSize,
    );
    const safeRootFontSize = Number.isFinite(rootFontSize) ? rootFontSize : 16;
    return BOARD_COLUMN_GAP_REM * safeRootFontSize;
  }, []);

  const boardMinWidth = useMemo(() => {
    return (
      columns * boardColumnWidth + Math.max(0, columns - 1) * boardColumnGapPx
    );
  }, [boardColumnGapPx, boardColumnWidth, columns]);

  const playerColumns = useMemo(() => {
    const nextColumns: Player[][] = Array.from({ length: columns }, () => []);

    orderedPlayers.forEach((player, index) => {
      nextColumns[index % columns].push(player);
    });

    return nextColumns;
  }, [columns, orderedPlayers]);

  const boardStyle = useMemo<CSSProperties>(() => {
    return {
      transform: `translate(${pan.x}px, ${pan.y}px) scale(${zoom})`,
      display: "flex",
      alignItems: "flex-start",
      gap: `${BOARD_COLUMN_GAP_REM}rem`,
      minWidth: `${boardMinWidth}px`,
    };
  }, [boardMinWidth, pan.x, pan.y, zoom]);

  const boardContentSize = useMemo<Size>(() => {
    return {
      width: Math.max(contentSize.width, boardMinWidth),
      height: contentSize.height,
    };
  }, [boardMinWidth, contentSize.height, contentSize.width]);

  const withErrorHandling = useCallback(
    (fn: () => void) => {
      try {
        fn();
      } catch (error) {
        onClientError?.(error);
      }
    },
    [onClientError],
  );

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
  }, [orderedPlayers.length]);

  useEffect(() => {
    setPan((current) =>
      clampPan(current, zoom, viewportSize, boardContentSize),
    );
  }, [zoom, viewportSize, boardContentSize]);

  const onBoardWheel = useCallback(
    (event: globalThis.WheelEvent) => {
      withErrorHandling(() => {
        event.preventDefault();
        const viewport = viewportRef.current;
        if (!viewport) {
          return;
        }

        const rect = viewport.getBoundingClientRect();
        const localX = event.clientX - rect.left;
        const localY = event.clientY - rect.top;
        const factor = event.deltaY < 0 ? 1.03 : 0.97;

        setZoom((currentZoom) => {
          const nextZoom = Math.min(
            MAX_ZOOM,
            Math.max(MIN_ZOOM, currentZoom * factor),
          );
          if (nextZoom === currentZoom) {
            return currentZoom;
          }

          setPan((currentPan) => {
            const worldX = (localX - currentPan.x) / currentZoom;
            const worldY = (localY - currentPan.y) / currentZoom;
            const candidatePan: Pan = {
              x: localX - worldX * nextZoom,
              y: localY - worldY * nextZoom,
            };
            return clampPan(
              candidatePan,
              nextZoom,
              viewportSize,
              boardContentSize,
            );
          });

          return nextZoom;
        });
      });
    },
    [boardContentSize, viewportSize, withErrorHandling],
  );

  useEffect(() => {
    const viewport = viewportRef.current;
    if (!viewport) {
      return;
    }

    viewport.addEventListener("wheel", onBoardWheel, { passive: false });

    return () => {
      viewport.removeEventListener("wheel", onBoardWheel);
    };
  }, [onBoardWheel]);

  const onBoardPointerDown = (event: PointerEvent<HTMLDivElement>) => {
    if (isDiscardRequired) {
      return;
    }

    const target = event.target as HTMLElement | null;
    if (
      target?.closest(
        "button, input, textarea, select, .md-demand, .md-player-picker, .md-color-picker",
      )
    ) {
      return;
    }

    if (event.pointerType === "touch") {
      didPanDuringPointerRef.current = false;
      activeTouchPointsRef.current.set(event.pointerId, {
        x: event.clientX,
        y: event.clientY,
      });
      event.currentTarget.setPointerCapture(event.pointerId);

      if (activeTouchPointsRef.current.size === 1) {
        panPointerIdRef.current = event.pointerId;
        panStartPointRef.current = { x: event.clientX, y: event.clientY };
        setIsPanning(true);
      }

      if (activeTouchPointsRef.current.size >= 2) {
        const [first, second] = Array.from(
          activeTouchPointsRef.current.values(),
        );
        if (!first || !second) {
          return;
        }

        const center = midpointBetweenPoints(first, second);
        const rect = event.currentTarget.getBoundingClientRect();
        const localCenter = {
          x: center.x - rect.left,
          y: center.y - rect.top,
        };
        const distance = distanceBetweenPoints(first, second);

        pinchStartDistanceRef.current = distance;
        pinchStartZoomRef.current = zoom;
        pinchWorldCenterRef.current = {
          x: (localCenter.x - pan.x) / zoom,
          y: (localCenter.y - pan.y) / zoom,
        };

        setIsPanning(false);
        panPointerIdRef.current = null;
        panStartPointRef.current = null;
      }

      event.preventDefault();
      return;
    }

    if (event.button !== 0) {
      return;
    }

    didPanDuringPointerRef.current = false;
    panPointerIdRef.current = event.pointerId;
    panStartPointRef.current = { x: event.clientX, y: event.clientY };
    setIsPanning(false);
  };

  const resetPanTracking = () => {
    didPanDuringPointerRef.current = false;
    setIsPanning(false);
    panPointerIdRef.current = null;
    panStartPointRef.current = null;
  };

  const onBoardPointerMove = (event: PointerEvent<HTMLDivElement>) => {
    if (event.pointerType === "touch") {
      const point = activeTouchPointsRef.current.get(event.pointerId);
      if (point) {
        point.x = event.clientX;
        point.y = event.clientY;
      }

      const pinchStartDistance = pinchStartDistanceRef.current;
      const worldCenter = pinchWorldCenterRef.current;
      if (
        activeTouchPointsRef.current.size >= 2 &&
        pinchStartDistance !== null &&
        worldCenter !== null
      ) {
        withErrorHandling(() => {
          event.preventDefault();

          const [first, second] = Array.from(
            activeTouchPointsRef.current.values(),
          );
          if (!first || !second) {
            return;
          }

          const center = midpointBetweenPoints(first, second);
          const currentDistance = distanceBetweenPoints(first, second);
          const scale = currentDistance / pinchStartDistance;
          const candidateZoom = pinchStartZoomRef.current * scale;
          const nextZoom = Math.min(
            MAX_ZOOM,
            Math.max(MIN_ZOOM, candidateZoom),
          );
          const rect = event.currentTarget.getBoundingClientRect();
          const localCenter = {
            x: center.x - rect.left,
            y: center.y - rect.top,
          };
          const candidatePan = {
            x: localCenter.x - worldCenter.x * nextZoom,
            y: localCenter.y - worldCenter.y * nextZoom,
          };

          setZoom(nextZoom);
          setPan(
            clampPan(candidatePan, nextZoom, viewportSize, boardContentSize),
          );
        });
        return;
      }
    }

    if (!isPanning || panPointerIdRef.current !== event.pointerId) {
      if (panPointerIdRef.current !== event.pointerId) {
        return;
      }
    }

    if (event.pointerType === "mouse" && (event.buttons & 1) === 0) {
      resetPanTracking();
      return;
    }

    withErrorHandling(() => {
      const previous = panStartPointRef.current;
      if (!previous) {
        return;
      }

      const deltaX = event.clientX - previous.x;
      const deltaY = event.clientY - previous.y;

      if (!didPanDuringPointerRef.current) {
        const movement = Math.hypot(deltaX, deltaY);
        if (movement < 4) {
          return;
        }
        didPanDuringPointerRef.current = true;
        setIsPanning(true);
      }

      event.preventDefault();
      window.getSelection()?.removeAllRanges();
      panStartPointRef.current = { x: event.clientX, y: event.clientY };
      setPan((current) =>
        clampPan(
          {
            x: current.x + deltaX,
            y: current.y + deltaY,
          },
          zoom,
          viewportSize,
          boardContentSize,
        ),
      );
    });
  };

  const stopPan = (event: PointerEvent<HTMLDivElement>) => {
    if (event.pointerType === "touch") {
      activeTouchPointsRef.current.delete(event.pointerId);
      if (activeTouchPointsRef.current.size < 2) {
        pinchStartDistanceRef.current = null;
        pinchWorldCenterRef.current = null;
      }
    }

    if (panPointerIdRef.current !== event.pointerId) {
      if (event.currentTarget.hasPointerCapture(event.pointerId)) {
        event.currentTarget.releasePointerCapture(event.pointerId);
      }
      return;
    }

    if (didPanDuringPointerRef.current) {
      suppressClickUntilRef.current = performance.now() + 220;
    }
    resetPanTracking();
    if (event.currentTarget.hasPointerCapture(event.pointerId)) {
      event.currentTarget.releasePointerCapture(event.pointerId);
    }
  };

  const onBoardClickCapture = (event: MouseEvent<HTMLDivElement>) => {
    if (performance.now() >= suppressClickUntilRef.current) {
      return;
    }

    event.preventDefault();
    event.stopPropagation();
  };

  const getCardById = useCallback(
    (cardId: string): Card | null => {
      return yourHand.find((candidate) => candidate.cardId === cardId) ?? null;
    },
    [yourHand],
  );

  const placeCardOnMoney = useCallback(
    (cardId: string) => {
      withErrorHandling(() => {
        onPlayMoneyCard(cardId);
      });
    },
    [onPlayMoneyCard, withErrorHandling],
  );

  const placeCardOnActionPile = useCallback(
    (cardId: string) => {
      withErrorHandling(() => {
        const card = getCardById(cardId);
        if (card?.assetKey === AssetKey.ASSET_KEY_DEBT_COLLECTOR) {
          setDebtCollectorPicker({
            mode: "debt_collector",
            cardId,
          });
          return;
        }

        if (card?.assetKey === AssetKey.ASSET_KEY_RENT_WILD) {
          setDebtCollectorPicker({
            mode: "wild_rent",
            cardId,
          });
          return;
        }

        if (card?.assetKey === AssetKey.ASSET_KEY_SLY_DEAL) {
          if (!hasMovesLeft) {
            console.log("[game-ui] sly deal blocked; no moves left");
            onGameError?.({
              message: "You have no moves left this turn.",
              code: "NO_MOVES_LEFT",
            });
            return;
          }

          if (stealableCardOwnerById.size === 0) {
            console.log(
              "[game-ui] sly deal blocked; no opponent property available",
            );
            onGameError?.({
              message:
                "No opponent has any property cards down. You cannot play Sly Deal.",
              code: "SLY_DEAL_NO_OPPONENT_PROPERTY",
            });
            return;
          }

          setDealPicker({
            mode: "sly_deal",
            cardId,
          });
          return;
        }

        if (card?.assetKey === AssetKey.ASSET_KEY_DEAL_BREAKER) {
          if (!hasMovesLeft) {
            console.log("[game-ui] deal breaker blocked; no moves left");
            onGameError?.({
              message: "You have no moves left this turn.",
              code: "NO_MOVES_LEFT",
            });
            return;
          }

          if (stealableCompleteSetOwnerById.size === 0) {
            console.log(
              "[game-ui] deal breaker blocked; no opponent complete set available",
            );
            onGameError?.({
              message:
                "No opponent has a complete property set. You cannot play Deal Breaker.",
              code: "DEAL_BREAKER_NO_COMPLETE_SET",
            });
            return;
          }

          setDealPicker({
            mode: "deal_breaker",
            cardId,
          });
          return;
        }

        if (card?.assetKey === AssetKey.ASSET_KEY_FORCED_DEAL) {
          if (!hasMovesLeft) {
            console.log("[game-ui] forced deal blocked; no moves left");
            onGameError?.({
              message: "You have no moves left this turn.",
              code: "NO_MOVES_LEFT",
            });
            return;
          }

          if (stealableCardOwnerById.size === 0) {
            console.log(
              "[game-ui] forced deal blocked; no opponent property available",
            );
            onGameError?.({
              message:
                "No opponent has any property cards down. You cannot play Forced Deal.",
              code: "FORCED_DEAL_NO_OPPONENT_PROPERTY",
            });
            return;
          }

          if (forcedDealSourceCardIds.size === 0) {
            console.log(
              "[game-ui] forced deal blocked; no own property available",
            );
            onGameError?.({
              message:
                "You have no property cards down. You cannot play Forced Deal.",
              code: "FORCED_DEAL_NO_OWN_PROPERTY",
            });
            return;
          }

          setDealPicker({
            mode: "forced_deal",
            cardId,
          });
          return;
        }

        onPlayPassGoCard(cardId);
      });
    },
    [
      forcedDealSourceCardIds.size,
      getCardById,
      hasMovesLeft,
      onGameError,
      onPlayPassGoCard,
      stealableCardOwnerById.size,
      stealableCompleteSetOwnerById.size,
      withErrorHandling,
    ],
  );

  const submitPropertyPlay = useCallback(
    (card: Card, propertySetId?: string) => {
      const selectableColors = toSelectableColors(card);
      if (!propertySetId && selectableColors.length > 1) {
        setColorPicker({
          mode: "play",
          cardId: card.cardId,
          propertySetId,
          colors: selectableColors,
        });
        return;
      }

      onPlayPropertyCard(
        card.cardId,
        propertySetId,
        selectableColors.length === 1 ? selectableColors[0] : undefined,
      );
    },
    [onPlayPropertyCard],
  );

  const submitRearrangeCard = useCallback(
    (card: Card, sourcePropertySetId: string, propertySetId?: string) => {
      if (!isSelfTurn) {
        console.log("[game-ui] rearrange blocked; not your turn", {
          cardId: card.cardId,
        });
        return;
      }

      if (!isPropertyCard(card)) {
        console.log("[game-ui] rearrange blocked; card not movable property", {
          cardId: card.cardId,
          category: card.category,
        });
        return;
      }

      if (propertySetId && propertySetId === sourcePropertySetId) {
        return;
      }

      if (!propertySetId) {
        const selectableColors = toSelectableColors(card);
        if (
          card.category === Category.CATEGORY_WILD_PROPERTY &&
          selectableColors.length > 1
        ) {
          setColorPicker({
            mode: "rearrange",
            cardId: card.cardId,
            propertySetId: undefined,
            sourcePropertySetId,
            colors: selectableColors,
          });
          return;
        }

        onRearrangeCard(card.cardId, undefined, selectableColors[0]);
        return;
      }

      onRearrangeCard(card.cardId, propertySetId);
    },
    [isSelfTurn, onRearrangeCard],
  );

  const placeCardOnProperty = useCallback(
    (cardId: string, propertySetId?: string) => {
      withErrorHandling(() => {
        const card = getCardById(cardId);
        if (!card) {
          return;
        }

        submitPropertyPlay(card, propertySetId);
      });
    },
    [getCardById, submitPropertyPlay, withErrorHandling],
  );

  const onDemandComply = useCallback(
    (demandId: string) => {
      const clickedDemand = visibleDemands.find(
        (demand) => demand.id === demandId,
      );
      if (!clickedDemand) {
        console.log("[game-ui] demand action: comply ignored (missing demand)");
        return;
      }

      if (clickedDemand.demandKind === DemandKind.DEMAND_KIND_PROPERTY) {
        onComplyPropertyDemand(demandId);
        return;
      }

      if (clickedDemand.demandKind === DemandKind.DEMAND_KIND_PROPERTY_SET) {
        onComplyPropertySetDemand(demandId);
        return;
      }

      if (clickedDemand.demandKind !== DemandKind.DEMAND_KIND_PAYMENT) {
        console.log(
          "[game-ui] demand action: comply ignored (unsupported demand kind)",
        );
        return;
      }

      if (!clickedDemand.isActive) {
        onComplyPaymentDemand(demandId, []);
        setIsSelectingPaymentCards(false);
        setActivePaymentDemandId(null);
        setSelectedPaymentCardIds(new Set());
        return;
      }

      if (!isSelectingPaymentCards || activePaymentDemandId !== demandId) {
        setIsSelectingPaymentCards(true);
        setActivePaymentDemandId(demandId);
        setSelectedPaymentCardIds(new Set());
        console.log("[game-ui] demand action: comply (selection started)");
        return;
      }

      if (!canConfirmPaymentSelection) {
        console.log(
          "[game-ui] demand action: comply blocked (selected value too low)",
          {
            selectedPaymentTotal,
            requiredAmount: paymentAmount,
          },
        );
        return;
      }

      const selectedCardIds = Array.from(selectedPaymentCardIds);
      onComplyPaymentDemand(demandId, selectedCardIds);
      setIsSelectingPaymentCards(false);
      setActivePaymentDemandId(null);
      setSelectedPaymentCardIds(new Set());
    },
    [
      activePaymentDemandId,
      canConfirmPaymentSelection,
      isSelectingPaymentCards,
      onComplyPropertyDemand,
      onComplyPropertySetDemand,
      onComplyPaymentDemand,
      paymentAmount,
      selectedPaymentCardIds,
      selectedPaymentTotal,
      visibleDemands,
    ],
  );

  const onDemandDeny = useCallback(
    (demandId: string) => {
      if (!hasJustSayNo) {
        console.log(
          "[game-ui] demand action: deny blocked (missing JUST_SAY_NO)",
        );
        return;
      }

      console.log("[game-ui] demand action: just-say-no");
      onDenyDemand(demandId);
    },
    [hasJustSayNo, onDenyDemand],
  );

  const onToggleSelectedPaymentCard = useCallback((card: Card) => {
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

  const onSelectDebtCollectorTarget = useCallback(
    (targetPlayerId: string) => {
      if (!debtCollectorPicker) {
        return;
      }

      if (selfPlayerId && targetPlayerId === selfPlayerId) {
        console.log(
          "[game-ui] debt collector target blocked (cannot select self)",
          {
            cardId: debtCollectorPicker.cardId,
            targetPlayerId,
          },
        );
        return;
      }

      console.log("[game-ui] debt collector target selected", {
        cardId: debtCollectorPicker.cardId,
        mode: debtCollectorPicker.mode,
        targetPlayerId,
      });

      if (debtCollectorPicker.mode === "wild_rent") {
        onPlayWildRentCard(debtCollectorPicker.cardId, targetPlayerId);
      } else {
        onPlayDebtCollectorCard(debtCollectorPicker.cardId, targetPlayerId);
      }
      setDebtCollectorPicker(null);
    },
    [
      debtCollectorPicker,
      onPlayDebtCollectorCard,
      onPlayWildRentCard,
      selfPlayerId,
    ],
  );

  const onPendingRentDouble = useCallback(() => {
    if (!hasDoubleTheRent) {
      console.log(
        "[game-ui] pending rent action: double blocked (missing DOUBLE_THE_RENT)",
      );
      return;
    }

    console.log("[game-ui] pending rent action: double");
    onPlayDoubleTheRentCard();
  }, [hasDoubleTheRent, onPlayDoubleTheRentCard]);

  const onPendingRentRent = useCallback(() => {
    console.log("[game-ui] pending rent action: rent");
    onResolvePendingRent();
  }, [onResolvePendingRent]);

  useEffect(() => {
    if (!pendingRent) {
      autoResolvedPendingRentKeyRef.current = null;
      return;
    }

    if (!shouldAutoResolvePendingRent) {
      autoResolvedPendingRentKeyRef.current = null;
      return;
    }

    const pendingRentKey = [
      pendingRent.playerId,
      pendingRent.baseAmount,
      pendingRent.multiplier,
      pendingRent.targetIds.join("|"),
    ].join(":");

    if (autoResolvedPendingRentKeyRef.current === pendingRentKey) {
      return;
    }

    autoResolvedPendingRentKeyRef.current = pendingRentKey;
    console.log(
      "[game-ui] pending rent action: auto-rent (missing DOUBLE_THE_RENT)",
    );
    onResolvePendingRent();
  }, [onResolvePendingRent, pendingRent, shouldAutoResolvePendingRent]);

  const onSelectDealCard = useCallback(
    (card: Card) => {
      if (!dealPicker) {
        return;
      }

      if (dealPicker.mode === "sly_deal") {
        const targetPlayerId = stealableCardOwnerById.get(card.cardId);
        if (!targetPlayerId) {
          return;
        }

        setDealPicker((current) => {
          if (!current || current.mode !== "sly_deal") {
            return current;
          }

          return {
            ...current,
            selectedOpponentCardId: card.cardId,
            selectedOpponentPlayerId: targetPlayerId,
          };
        });
        return;
      }

      if (dealPicker.mode === "deal_breaker") {
        return;
      }

      const targetPlayerId = stealableCardOwnerById.get(card.cardId);
      if (targetPlayerId) {
        setDealPicker((current) => {
          if (!current || current.mode !== "forced_deal") {
            return current;
          }

          return {
            ...current,
            selectedOpponentCardId: card.cardId,
            selectedOpponentPlayerId: targetPlayerId,
          };
        });
        return;
      }

      if (!forcedDealSourceCardIds.has(card.cardId)) {
        return;
      }

      setDealPicker((current) => {
        if (!current || current.mode !== "forced_deal") {
          return current;
        }

        return {
          ...current,
          selectedSourceCardId: card.cardId,
        };
      });
    },
    [dealPicker, forcedDealSourceCardIds, stealableCardOwnerById],
  );

  const onSelectCardFromCompletedSet = useCallback(() => {
    onGameError?.({
      message:
        "Cards in completed property sets cannot be targeted by Property Swap or Snatch.",
      code: "DEAL_CARD_IN_COMPLETED_SET",
    });
  }, [onGameError]);

  const onSelectIncompleteSetForDealBreaker = useCallback(() => {
    onGameError?.({
      message: "Set Snatcher can only snatch complete sets.",
      code: "DEAL_BREAKER_REQUIRES_COMPLETE_SET",
    });
  }, [onGameError]);

  const onSelectDealPropertySet = useCallback(
    (propertySet: PropertySet, playerId: string) => {
      if (!dealPicker || dealPicker.mode !== "deal_breaker") {
        return;
      }

      if (!isPropertySetComplete(propertySet)) {
        onSelectIncompleteSetForDealBreaker();
        return;
      }

      if (selfPlayerId && playerId === selfPlayerId) {
        return;
      }

      setDealPicker((current) => {
        if (!current || current.mode !== "deal_breaker") {
          return current;
        }

        return {
          ...current,
          selectedTargetPropertySetId: propertySet.propertySetId,
          selectedOpponentPlayerId: playerId,
        };
      });
    },
    [dealPicker, onSelectIncompleteSetForDealBreaker, selfPlayerId],
  );

  const onToggleSelectedRearrangeCard = useCallback(
    (card: Card, sourcePropertySetId: string) => {
      if (!isPropertyCard(card)) {
        return;
      }

      setSelectedRearrangeCard((current) => {
        if (
          current &&
          current.cardId === card.cardId &&
          current.sourcePropertySetId === sourcePropertySetId
        ) {
          return null;
        }

        return {
          cardId: card.cardId,
          sourcePropertySetId,
        };
      });
    },
    [],
  );

  const onApplySelectedRearrangeCard = useCallback(
    (targetPropertySetId?: string) => {
      if (!selectedRearrangeCard) {
        return;
      }

      const sourcePropertySet = selfPropertySets.find((propertySet) => {
        return (
          propertySet.propertySetId ===
          selectedRearrangeCard.sourcePropertySetId
        );
      });
      if (!sourcePropertySet || isPropertySetLocked(sourcePropertySet)) {
        setSelectedRearrangeCard(null);
        return;
      }

      const sourceCard = sourcePropertySet.cards.find(
        (card) => card.cardId === selectedRearrangeCard.cardId,
      );
      if (!sourceCard || !isPropertyCard(sourceCard)) {
        setSelectedRearrangeCard(null);
        return;
      }

      submitRearrangeCard(
        sourceCard,
        selectedRearrangeCard.sourcePropertySetId,
        targetPropertySetId,
      );
      setSelectedRearrangeCard(null);
    },
    [selectedRearrangeCard, selfPropertySets, submitRearrangeCard],
  );

  const onConfirmDealSelection = useCallback(() => {
    if (!dealPicker || !dealPicker.selectedOpponentPlayerId) {
      return;
    }

    if (dealPicker.mode === "deal_breaker") {
      if (!dealPicker.selectedTargetPropertySetId) {
        return;
      }

      if (
        stealableCompleteSetOwnerById.get(
          dealPicker.selectedTargetPropertySetId,
        ) !== dealPicker.selectedOpponentPlayerId
      ) {
        return;
      }

      onPlayDealBreakerCard(
        dealPicker.cardId,
        dealPicker.selectedOpponentPlayerId,
        dealPicker.selectedTargetPropertySetId,
      );
      setDealPicker(null);
      return;
    }

    if (!dealPicker.selectedOpponentCardId) {
      return;
    }

    if (
      stealableCardOwnerById.get(dealPicker.selectedOpponentCardId) !==
      dealPicker.selectedOpponentPlayerId
    ) {
      return;
    }

    if (dealPicker.mode === "sly_deal") {
      onPlaySlyDealCard(
        dealPicker.cardId,
        dealPicker.selectedOpponentPlayerId,
        dealPicker.selectedOpponentCardId,
      );
      setDealPicker(null);
      return;
    }

    if (
      !dealPicker.selectedSourceCardId ||
      !forcedDealSourceCardIds.has(dealPicker.selectedSourceCardId)
    ) {
      return;
    }

    onPlayForcedDealCard(
      dealPicker.cardId,
      dealPicker.selectedOpponentPlayerId,
      dealPicker.selectedSourceCardId,
      dealPicker.selectedOpponentCardId,
    );
    setDealPicker(null);
  }, [
    dealPicker,
    forcedDealSourceCardIds,
    onPlayDealBreakerCard,
    onPlayForcedDealCard,
    onPlaySlyDealCard,
    stealableCompleteSetOwnerById,
    stealableCardOwnerById,
  ]);

  const canConfirmDealSelection =
    !!dealPicker &&
    (dealPicker.mode === "deal_breaker"
      ? !!dealPicker.selectedTargetPropertySetId &&
        !!dealPicker.selectedOpponentPlayerId &&
        stealableCompleteSetOwnerById.get(
          dealPicker.selectedTargetPropertySetId,
        ) === dealPicker.selectedOpponentPlayerId
      : !!dealPicker.selectedOpponentCardId &&
        !!dealPicker.selectedOpponentPlayerId &&
        stealableCardOwnerById.get(dealPicker.selectedOpponentCardId) ===
          dealPicker.selectedOpponentPlayerId &&
        (dealPicker.mode === "sly_deal" ||
          (!!dealPicker.selectedSourceCardId &&
            forcedDealSourceCardIds.has(dealPicker.selectedSourceCardId))));

  const onCancelDealPicker = useCallback(() => {
    setDealPicker(null);
  }, []);

  useEffect(() => {
    if (isSelectingPaymentCards && !isSelectedPaymentDemandActive) {
      setIsSelectingPaymentCards(false);
      setActivePaymentDemandId(null);
      setSelectedPaymentCardIds(new Set());
    }
  }, [isSelectedPaymentDemandActive, isSelectingPaymentCards]);

  useEffect(() => {
    if (!dealPicker) {
      return;
    }

    const stillInHand = yourHand.some(
      (card) => card.cardId === dealPicker.cardId,
    );
    if (!stillInHand || !isSelfTurn || isDiscardRequired) {
      setDealPicker(null);
    }
  }, [dealPicker, isDiscardRequired, isSelfTurn, yourHand]);

  useEffect(() => {
    if (!selectedHandPlacementCardId) {
      return;
    }

    const stillInHand = yourHand.some(
      (card) => card.cardId === selectedHandPlacementCardId,
    );
    if (!stillInHand || !canSelectHandCards) {
      setSelectedHandPlacementCardId(null);
    }
  }, [canSelectHandCards, selectedHandPlacementCardId, yourHand]);

  useEffect(() => {
    if (!selectedRearrangeCard) {
      return;
    }

    const sourcePropertySet = selfPropertySets.find((propertySet) => {
      return (
        propertySet.propertySetId === selectedRearrangeCard.sourcePropertySetId
      );
    });
    const stillExists = sourcePropertySet?.cards.some((card) => {
      return card.cardId === selectedRearrangeCard.cardId;
    });

    if (
      !isSelfTurn ||
      isDiscardRequired ||
      isSelectingPaymentCards ||
      isSelectingDealCards ||
      !stillExists ||
      (sourcePropertySet ? isPropertySetLocked(sourcePropertySet) : true)
    ) {
      setSelectedRearrangeCard(null);
    }
  }, [
    isDiscardRequired,
    isSelectingDealCards,
    isSelectingPaymentCards,
    isSelfTurn,
    selectedRearrangeCard,
    selfPropertySets,
  ]);

  const renderPlayerBoard = (player: Player) => {
    const playerMoney = moneyByPlayer[player.playerId] ?? [];
    const playerMoneyTotal = playerMoney.reduce((total, card) => {
      return total + (card.value ?? 0);
    }, 0);
    const propertySets = propertySetsByPlayer[player.playerId] ?? [];
    const isCurrentPlayer = player.playerId === currentPlayerId;
    const isSelfBoard = !!selfPlayerId && player.playerId === selfPlayerId;
    const canInteractWithBoard =
      isSelfBoard && isSelfTurn && !isDiscardRequired && !isSelectingDealCards;
    const canSelectBoardCards =
      isSelfBoard && isSelectingPaymentCards && isSelectedPaymentDemandActive;
    const canSelectDealCards =
      !!dealPicker &&
      (dealPicker.mode === "deal_breaker"
        ? !isSelfBoard
        : dealPicker.mode === "forced_deal"
          ? true
          : !isSelfBoard);

    const canRearrangeFromBoard =
      isSelfBoard &&
      isSelfTurn &&
      !isSelectingPaymentCards &&
      !isSelectingDealCards &&
      !isDiscardRequired;
    const canClickRearrangeDestination =
      canRearrangeFromBoard && !!selectedRearrangeCard;
    const moneyEmptyLabel =
      isSelfBoard && isSelfTurn
        ? "Select a card, then tap here to play as money."
        : "";
    const propertyNewEmptyLabel =
      isSelfBoard && isSelfTurn
        ? "Select a card, then tap here to start a new set."
        : "";

    return (
      <article
        key={player.playerId}
        className={
          isCurrentPlayer
            ? "md-player-board md-player-board--current"
            : "md-player-board"
        }
      >
        <header className="md-player-board__header">
          <img
            className="md-player-board__avatar"
            src={player.avatarUrl}
            alt={player.displayName}
            loading="lazy"
            referrerPolicy="no-referrer"
          />
          <div>
            <p className="md-player-board__name">{player.displayName}</p>
            <p className="md-player-board__stats">
              ${playerMoneyTotal}M total · {player.completedSets} sets ·{" "}
              {player.handCards} in hand
            </p>
          </div>
        </header>

        <div className="md-player-board__body">
          <div
            className="md-dropzone md-dropzone--money"
            onClick={() => {
              if (canClickRearrangeDestination) {
                onApplySelectedRearrangeCard(undefined);
                return;
              }

              if (!canInteractWithBoard || !selectedHandPlacementCardId) {
                return;
              }

              placeCardOnMoney(selectedHandPlacementCardId);
              setSelectedHandPlacementCardId(null);
            }}
          >
            <GameCardStackBox
              title={`Money · $${playerMoneyTotal}M`}
              cards={playerMoney}
              assetImageByKey={assetImageByKey}
              layout="stack"
              emptyLabel={moneyEmptyLabel}
              selectableCards={
                canSelectBoardCards ||
                (canSelectDealCards && dealPicker?.mode !== "deal_breaker")
              }
              selectedCardIds={
                canSelectBoardCards
                  ? selectedPaymentCardIds
                  : canClickRearrangeDestination
                    ? selectedRearrangeCardIds
                    : selectedDealCardIds
              }
              onCardClick={(card) => {
                if (canClickRearrangeDestination) {
                  onApplySelectedRearrangeCard(undefined);
                  return;
                }

                if (canInteractWithBoard && selectedHandPlacementCardId) {
                  placeCardOnMoney(selectedHandPlacementCardId);
                  setSelectedHandPlacementCardId(null);
                  return;
                }

                if (canSelectBoardCards) {
                  onToggleSelectedPaymentCard(card);
                  return;
                }

                if (canSelectDealCards) {
                  onSelectDealCard(card);
                }
              }}
            />
          </div>

          {propertySets.map((propertySet, index) => (
            <div
              className={[
                "md-dropzone",
                "md-dropzone--property-set",
                selectedDealPropertySetIds.has(propertySet.propertySetId)
                  ? "is-set-selected"
                  : "",
              ]
                .filter(Boolean)
                .join(" ")}
              key={propertySet.propertySetId}
              onClick={(event) => {
                if ((event.target as HTMLElement | null)?.closest(".md-card")) {
                  return;
                }

                if (canSelectDealCards && dealPicker?.mode === "deal_breaker") {
                  onSelectDealPropertySet(propertySet, player.playerId);
                  return;
                }

                if (canClickRearrangeDestination) {
                  onApplySelectedRearrangeCard(propertySet.propertySetId);
                  return;
                }

                if (!canInteractWithBoard || !selectedHandPlacementCardId) {
                  return;
                }

                placeCardOnProperty(
                  selectedHandPlacementCardId,
                  propertySet.propertySetId,
                );
                setSelectedHandPlacementCardId(null);
              }}
            >
              <GameCardStackBox
                title={`Set ${index + 1} · Rent $${getBaseRentForPropertySet(propertySet)}M${(() => {
                  const rentBonuses = getRentBonusLabels(propertySet);
                  return rentBonuses.length > 0 ? ` ${rentBonuses.join(" ")}` : "";
                })()}`}
                cards={propertySet.cards}
                assetImageByKey={assetImageByKey}
                layout="stack"
                color={propertySet.color}
                emptyLabel=""
                selectableCards={
                  canSelectBoardCards ||
                  canSelectDealCards ||
                  canRearrangeFromBoard
                }
                selectedCardIds={
                  canSelectBoardCards
                    ? selectedPaymentCardIds
                    : selectedRearrangeCard
                      ? selectedRearrangeCardIds
                      : selectedDealCardIds
                }
                onCardClick={(card) => {
                  if (canClickRearrangeDestination) {
                    onApplySelectedRearrangeCard(propertySet.propertySetId);
                    return;
                  }

                  if (canInteractWithBoard && selectedHandPlacementCardId) {
                    placeCardOnProperty(
                      selectedHandPlacementCardId,
                      propertySet.propertySetId,
                    );
                    setSelectedHandPlacementCardId(null);
                    return;
                  }

                  if (canSelectBoardCards) {
                    onToggleSelectedPaymentCard(card);
                    return;
                  }

                  if (canSelectDealCards) {
                    if (dealPicker?.mode === "deal_breaker") {
                      onSelectDealPropertySet(propertySet, player.playerId);
                      return;
                    }

                    if (isPropertySetComplete(propertySet)) {
                      onSelectCardFromCompletedSet();
                      return;
                    }

                    onSelectDealCard(card);
                    return;
                  }

                  if (
                    canRearrangeFromBoard &&
                    !isPropertySetLocked(propertySet) &&
                    isPropertyCard(card)
                  ) {
                    onToggleSelectedRearrangeCard(
                      card,
                      propertySet.propertySetId,
                    );
                  }
                }}
              />
            </div>
          ))}

          {isSelfBoard ? (
            <div
              className="md-dropzone md-dropzone--property-new"
              onClick={() => {
                if (canClickRearrangeDestination) {
                  onApplySelectedRearrangeCard(undefined);
                  return;
                }

                if (!canInteractWithBoard || !selectedHandPlacementCardId) {
                  return;
                }

                placeCardOnProperty(selectedHandPlacementCardId, undefined);
                setSelectedHandPlacementCardId(null);
              }}
            >
              <div className="md-property-empty">
                {propertyNewEmptyLabel}
              </div>
            </div>
          ) : null}
        </div>
      </article>
    );
  };

  return (
    <div className="md-html-board">
      <div className="md-demand-overlay-wrap">
        {dealPicker ? (
          <aside
            className="md-demand md-demand--payment"
            aria-live="polite"
            onPointerDown={(event) => event.stopPropagation()}
          >
            <p className="md-demand__eyebrow">
              {dealPicker.mode === "forced_deal"
                ? "Property Swap"
                : dealPicker.mode === "deal_breaker"
                  ? "Set Snatcher"
                  : "Property Steal"}
            </p>
            <h3 className="md-demand__title">
              {dealPicker.mode === "deal_breaker"
                ? "Select an opponenets complete set to steal."
                : dealPicker.mode === "forced_deal"
                  ? "Choose one of your opponents properties and one of yours to swap."
                  : "Choose one of your opponents properties to steal."}
            </h3>
            <div className="md-demand__actions">
              <button
                type="button"
                className="md-demand__button"
                onPointerDown={(event) => event.stopPropagation()}
                onClick={onConfirmDealSelection}
                disabled={!canConfirmDealSelection}
              >
                Confirm
              </button>
              <button
                type="button"
                className="md-demand__button md-demand__button--secondary"
                onPointerDown={(event) => event.stopPropagation()}
                onClick={onCancelDealPicker}
              >
                Cancel
              </button>
            </div>
          </aside>
        ) : shouldShowPendingRentOverlay && pendingRent ? (
          <PendingRentOverlay
            pendingRent={pendingRent}
            players={players}
            selfPlayerId={selfPlayerId ?? undefined}
            canDouble={hasDoubleTheRent && hasMovesLeft}
            onDouble={onPendingRentDouble}
            onRent={onPendingRentRent}
          />
        ) : isDiscardRequired ? (
          <DiscardPromptOverlay
            requiredDiscardCount={requiredDiscardCount}
            selectedDiscardCount={selectedDiscardCardIds.size}
          />
        ) : (
          <DemandOverlay
            demands={visibleDemands}
            players={players}
            selfPlayerId={selfPlayerId}
            targetCardImageById={propertyCardImageById}
            propertySetCardsById={propertySetCardsById}
            canDeny={hasJustSayNo}
            isSelectingCards={isSelectingPaymentCards}
            selectingDemandId={activePaymentDemandId ?? undefined}
            canConfirmSelection={canConfirmPaymentSelection}
            selectedPaymentTotal={selectedPaymentTotal}
            onComply={onDemandComply}
            onDeny={onDemandDeny}
          />
        )}
      </div>

      <section
        className={[
          "md-board-viewport",
          isPanning ? "is-panning" : "",
          shouldDimBoard ? "md-board-viewport--demand-active" : "",
        ]
          .filter(Boolean)
          .join(" ")}
        ref={viewportRef}
        onPointerDownCapture={onBoardPointerDown}
        onPointerMoveCapture={onBoardPointerMove}
        onPointerUpCapture={stopPan}
        onPointerCancelCapture={stopPan}
        onClickCapture={onBoardClickCapture}
      >
        <div
          className={
            shouldDimBoard
              ? "md-board-content md-board-content--demand-active"
              : "md-board-content"
          }
          ref={contentRef}
          style={boardStyle}
        >
          {playerColumns.map((playerColumn, columnIndex) => (
            <div
              className="md-board-column"
              key={`col-${columnIndex}`}
              style={{
                width: `${boardColumnWidth}px`,
              }}
            >
              {playerColumn.map((player) => renderPlayerBoard(player))}
            </div>
          ))}
        </div>
      </section>

      <section className="md-hand-row">
        <div
          className="md-dropzone md-dropzone--action"
          onClick={() => {
            if (!canTapPlaceFromHand || !selectedHandPlacementCardId) {
              return;
            }

            placeCardOnActionPile(selectedHandPlacementCardId);
            setSelectedHandPlacementCardId(null);
          }}
        >
          <GameCardStackBox
            title="Actions"
            cards={lastActionCards}
            assetImageByKey={assetImageByKey}
            layout="spread"
            emptyLabel="Select a card, then tap here"
          />
        </div>

        <div className="md-hand-main">
          <GameCardStackBox
            title={shouldShowPaymentMoneyPicker ? "Your money" : "Your hand"}
            cards={shouldShowPaymentMoneyPicker ? selfMoneyCards : yourHand}
            assetImageByKey={assetImageByKey}
            layout="spread"
            emptyLabel={
              shouldShowPaymentMoneyPicker
                ? "No money cards available"
                : "Waiting for cards"
            }
            selectableCards={
              shouldShowPaymentMoneyPicker || isDiscardRequired || canSelectHandCards
            }
            selectedCardIds={
              shouldShowPaymentMoneyPicker
                ? selectedPaymentCardIds
                : isDiscardRequired
                ? selectedDiscardCardIds
                : selectedHandPlacementCardIds
            }
            onCardClick={(card) => {
              if (shouldShowPaymentMoneyPicker) {
                onToggleSelectedPaymentCard(card);
                return;
              }

              if (isDiscardRequired) {
                onToggleDiscardCard(card.cardId);
                return;
              }

              if (!canSelectHandCards) {
                return;
              }

              setSelectedHandPlacementCardId((current) => {
                return current === card.cardId ? null : card.cardId;
              });
            }}
          />
        </div>

        <section className="md-hand-turn-controls">
          <button
            type="button"
            className="md-demand__button md-hand-turn-button"
            onClick={isDiscardRequired ? onSubmitDiscard : onPassTurn}
            disabled={
              !isSelfTurn ||
              (isDiscardRequired
                ? selectedDiscardCardIds.size !== requiredDiscardCount
                : false)
            }
          >
            {isDiscardRequired
              ? `Submit Discard (${selectedDiscardCardIds.size}/${requiredDiscardCount})`
              : isSelfTurn
                ? (
                    <>
                      Pass Turn{" "}
                      <span className="md-hand-turn-button__moves">({movesLeft} left)</span>
                    </>
                  )
                : "Pass Turn"}
          </button>
        </section>
      </section>

      {colorPicker ? (
        <div
          className="md-color-picker-backdrop"
          role="presentation"
          onClick={() => setColorPicker(null)}
        >
          <div
            className="md-color-picker md-demand md-demand--payment"
            role="dialog"
            aria-modal="true"
            aria-label="Choose property color"
            onClick={(event) => event.stopPropagation()}
          >
            <p className="md-demand__eyebrow">Color selection</p>
            <h3 className="md-demand__title">Choose a color</h3>
            <p className="md-demand__line">Pick one option to continue.</p>
            <div
              className="md-color-picker__swatches"
              style={{
                ["--md-color-cols" as string]: String(
                  Math.min(5, Math.max(1, colorPicker.colors.length)),
                ),
              }}
            >
              {colorPicker.colors.map((color) => (
                <button
                  type="button"
                  key={`${colorPicker.cardId}-${color}`}
                  className="md-color-picker__swatch"
                  style={{ background: colorHex(color) }}
                  onClick={() => {
                    if (colorPicker.mode === "rearrange") {
                      onRearrangeCard(colorPicker.cardId, undefined, color);
                    } else {
                      onPlayPropertyCard(
                        colorPicker.cardId,
                        colorPicker.propertySetId,
                        color,
                      );
                    }
                    setColorPicker(null);
                  }}
                  aria-label={`Choose ${colorName(color)}`}
                  title={colorName(color)}
                />
              ))}
            </div>
          </div>
        </div>
      ) : null}

      {debtCollectorPicker ? (
        <DebtCollectorTargetPickerOverlay
          players={players}
          selfPlayerId={selfPlayerId}
          eyebrow={
            debtCollectorPicker.mode === "wild_rent"
              ? "Wild Rent"
              : "SETTLE UP"
          }
          ariaLabel={
            debtCollectorPicker.mode === "wild_rent"
              ? "Choose wild rent target"
              : "Choose debt collector target"
          }
          onSelectPlayer={onSelectDebtCollectorTarget}
          onClose={() => setDebtCollectorPicker(null)}
        />
      ) : null}
    </div>
  );
};

export default MonopolyDealHtmlBoard;
