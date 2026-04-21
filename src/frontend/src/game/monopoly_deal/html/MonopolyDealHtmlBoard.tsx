import {
  useCallback,
  useEffect,
  useMemo,
  useRef,
  useState,
  type CSSProperties,
  type DragEvent,
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
import SlyDealPropertyPickerOverlay from "./action/SlyDealPropertyPickerOverlay";
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
  onDenyDemand: (demandId: string) => void;
  onClientError?: (error: unknown) => void;
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
  mode: "debt_collector" | "wild_rent" | "sly_deal" | "forced_deal";
  cardId: string;
};

type SlyDealPickerState = {
  cardId: string;
  targetPlayerId: string;
  targetCardId?: string;
  mode: "sly_deal" | "forced_deal";
};

type DraggingCardState = {
  cardId: string;
  source: "hand" | "board";
  sourcePropertySetId?: string;
};

const MIN_ZOOM = 0.65;
const MAX_ZOOM = 2.2;

const chooseGrid = (count: number): { rows: number; columns: number } => {
  if (count <= 0) {
    return { rows: 1, columns: 1 };
  }

  let bestRows = 1;
  let bestColumns = count;
  let bestScore = Number.POSITIVE_INFINITY;

  for (let columns = 1; columns <= count; columns += 1) {
    const rows = Math.ceil(count / columns);
    const diff = Math.abs(rows - columns);
    const overflow = rows * columns - count;
    const score = diff * 100 + overflow;

    if (score < bestScore) {
      bestScore = score;
      bestRows = rows;
      bestColumns = columns;
    }
  }

  return { rows: bestRows, columns: bestColumns };
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

const requiredPropertyCountForColor = (color: Color): number => {
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
  const requiredCount = requiredPropertyCountForColor(propertySet.color);
  if (!Number.isFinite(requiredCount)) {
    return false;
  }

  return countPropertyCards(propertySet) >= requiredCount;
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

const MonopolyDealHtmlBoard = ({
  gameState,
  assetImageByKey,
  selfPlayerId,
  onPlayMoneyCard,
  onPlayPassGoCard,
  onPlayDebtCollectorCard,
  onPlayWildRentCard,
  onPlaySlyDealCard,
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
  onClientError,
}: MonopolyDealHtmlBoardProps) => {
  const viewportRef = useRef<HTMLDivElement | null>(null);
  const contentRef = useRef<HTMLDivElement | null>(null);
  const [zoom, setZoom] = useState(1);
  const [pan, setPan] = useState<Pan>({ x: 0, y: 0 });
  const [viewportSize, setViewportSize] = useState<Size>({ width: 1, height: 1 });
  const [contentSize, setContentSize] = useState<Size>({ width: 1, height: 1 });
  const [isPanning, setIsPanning] = useState(false);
  const [draggingCard, setDraggingCard] = useState<DraggingCardState | null>(null);
  const [colorPicker, setColorPicker] = useState<ColorPickerState | null>(null);
  const [debtCollectorPicker, setDebtCollectorPicker] =
    useState<DebtCollectorPickerState | null>(null);
  const [slyDealPicker, setSlyDealPicker] = useState<SlyDealPickerState | null>(null);
  const [isSelectingPaymentCards, setIsSelectingPaymentCards] = useState(false);
  const [selectedPaymentCardIds, setSelectedPaymentCardIds] = useState<Set<string>>(
    () => new Set(),
  );
  const [activePaymentDemandId, setActivePaymentDemandId] = useState<string | null>(null);
  const panPointerIdRef = useRef<number | null>(null);
  const panStartPointRef = useRef<{ x: number; y: number } | null>(null);

  const players = gameState?.players ?? [];
  const visibleDemands = useMemo<Demand[]>(() => {
    const demands = gameState?.demands ?? [];
    const selfDemands = selfPlayerId
      ? demands.filter((demand) => demand.playerId === selfPlayerId)
      : demands;
    const prioritizedDemands = selfDemands.length > 0 ? selfDemands : demands;
    return prioritizedDemands.slice(0, 3);
  }, [gameState?.demands, selfPlayerId]);

  const hasAnyDemand = visibleDemands.length > 0;
  const hasPendingRent = !!gameState?.pendingRent;
  const currentPlayerId = gameState?.currentPlayerId ?? "";
  const isSelfTurn = !!selfPlayerId && selfPlayerId === currentPlayerId;
  const yourHand = gameState?.yourHand?.cards ?? [];
  const hasJustSayNo = yourHand.some(
    (card) => card.assetKey === AssetKey.ASSET_KEY_JUST_SAY_NO,
  );
  const hasDoubleTheRent = yourHand.some(
    (card) => card.assetKey === AssetKey.ASSET_KEY_DOUBLE_THE_RENT,
  );
  const hasMovesLeft = (gameState?.movesLeft ?? 0) > 0;
  const lastActionCards = gameState?.lastAction ? [gameState.lastAction] : [];

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

  const shouldDimBoard =
    isDiscardRequired ||
    hasPendingRent ||
    (hasAnyDemand && !(isSelectedPaymentDemandActive && isSelectingPaymentCards));

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

  const { rows, columns } = useMemo(() => chooseGrid(players.length), [players.length]);
  const boardStyle = useMemo<CSSProperties>(() => {
    return {
      transform: `translate(${pan.x}px, ${pan.y}px) scale(${zoom})`,
      gridTemplateColumns: `repeat(${columns}, minmax(0%, 1fr))`,
      gridTemplateRows: `repeat(${rows}, minmax(0%, auto))`,
    };
  }, [columns, pan.x, pan.y, rows, zoom]);

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
  }, [players.length, rows, columns]);

  useEffect(() => {
    setPan((current) => clampPan(current, zoom, viewportSize, contentSize));
  }, [zoom, viewportSize, contentSize]);

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
            return clampPan(candidatePan, nextZoom, viewportSize, contentSize);
          });

          return nextZoom;
        });
      });
    },
    [contentSize, viewportSize, withErrorHandling],
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

    if (event.button !== 0) {
      return;
    }

    panPointerIdRef.current = event.pointerId;
    panStartPointRef.current = { x: event.clientX, y: event.clientY };
    setIsPanning(true);
    event.currentTarget.setPointerCapture(event.pointerId);
  };

  const onBoardPointerMove = (event: PointerEvent<HTMLDivElement>) => {
    if (!isPanning || panPointerIdRef.current !== event.pointerId) {
      return;
    }

    withErrorHandling(() => {
      const previous = panStartPointRef.current;
      if (!previous) {
        return;
      }

      const deltaX = event.clientX - previous.x;
      const deltaY = event.clientY - previous.y;
      panStartPointRef.current = { x: event.clientX, y: event.clientY };
      setPan((current) =>
        clampPan(
          {
            x: current.x + deltaX,
            y: current.y + deltaY,
          },
          zoom,
          viewportSize,
          contentSize,
        ),
      );
    });
  };

  const stopPan = (event: PointerEvent<HTMLDivElement>) => {
    if (panPointerIdRef.current !== event.pointerId) {
      return;
    }

    setIsPanning(false);
    panPointerIdRef.current = null;
    panStartPointRef.current = null;
    if (event.currentTarget.hasPointerCapture(event.pointerId)) {
      event.currentTarget.releasePointerCapture(event.pointerId);
    }
  };

  const getCardById = useCallback(
    (cardId: string): Card | null => {
      return yourHand.find((candidate) => candidate.cardId === cardId) ?? null;
    },
    [yourHand],
  );

  const onDropMoney = (event: DragEvent<HTMLElement>) => {
    event.preventDefault();
    withErrorHandling(() => {
      if (draggingCard?.source === "board") {
        return;
      }

      const cardId = event.dataTransfer.getData("text/plain");
      if (!cardId) {
        return;
      }
      onPlayMoneyCard(cardId);
    });
  };

  const onDropPassGo = (event: DragEvent<HTMLElement>) => {
    event.preventDefault();
    withErrorHandling(() => {
      if (draggingCard?.source === "board") {
        return;
      }

      const cardId = event.dataTransfer.getData("text/plain");
      if (!cardId) {
        return;
      }

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
        setDebtCollectorPicker({
          mode: "sly_deal",
          cardId,
        });
        return;
      }

      if (card?.assetKey === AssetKey.ASSET_KEY_FORCED_DEAL) {
        setDebtCollectorPicker({
          mode: "forced_deal",
          cardId,
        });
        return;
      }

      onPlayPassGoCard(cardId);
    });
  };

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
        if (card.category === Category.CATEGORY_WILD_PROPERTY && selectableColors.length > 1) {
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

  const onDropProperty = (propertySetId?: string) => (event: DragEvent<HTMLElement>) => {
    event.preventDefault();
    withErrorHandling(() => {
      const cardId = event.dataTransfer.getData("text/plain");
      if (!cardId) {
        return;
      }

      if (draggingCard?.source === "board" && draggingCard.cardId === cardId) {
        if (!draggingCard.sourcePropertySetId) {
          return;
        }

        const sourcePropertySet = selfPropertySets.find((propertySet) => {
          return propertySet.propertySetId === draggingCard.sourcePropertySetId;
        });
        if (!sourcePropertySet || isPropertySetComplete(sourcePropertySet)) {
          console.log("[game-ui] rearrange blocked; source set is complete", {
            cardId,
            sourcePropertySetId: draggingCard.sourcePropertySetId,
          });
          return;
        }

        const sourceCard = sourcePropertySet.cards.find((card) => card.cardId === cardId);
        if (!sourceCard || !isPropertyCard(sourceCard)) {
          console.log("[game-ui] rearrange blocked; card not movable", {
            cardId,
          });
          return;
        }

        submitRearrangeCard(
          sourceCard,
          draggingCard.sourcePropertySetId,
          propertySetId,
        );
        return;
      }

      const card = getCardById(cardId);
      if (!card) {
        return;
      }

      submitPropertyPlay(card, propertySetId);
    });
  };

  const onAllowDrop = (event: DragEvent<HTMLElement>) => {
    event.preventDefault();
  };

  const onDemandComply = useCallback((demandId: string) => {
    const clickedDemand = visibleDemands.find((demand) => demand.id === demandId);
    if (!clickedDemand) {
      console.log("[game-ui] demand action: comply ignored (missing demand)");
      return;
    }

    if (clickedDemand.demandKind === DemandKind.DEMAND_KIND_PROPERTY) {
      onComplyPropertyDemand(demandId);
      return;
    }

    if (clickedDemand.demandKind !== DemandKind.DEMAND_KIND_PAYMENT) {
      console.log("[game-ui] demand action: comply ignored (unsupported demand kind)");
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
      console.log("[game-ui] demand action: comply blocked (selected value too low)", {
        selectedPaymentTotal,
        requiredAmount: paymentAmount,
      });
      return;
    }

    const selectedCardIds = Array.from(selectedPaymentCardIds);
    onComplyPaymentDemand(demandId, selectedCardIds);
    setIsSelectingPaymentCards(false);
    setActivePaymentDemandId(null);
    setSelectedPaymentCardIds(new Set());
  }, [
    activePaymentDemandId,
    canConfirmPaymentSelection,
    isSelectingPaymentCards,
    onComplyPropertyDemand,
    onComplyPaymentDemand,
    paymentAmount,
    selectedPaymentCardIds,
    selectedPaymentTotal,
    visibleDemands,
  ]);

  const onDemandDeny = useCallback((demandId: string) => {
    if (!hasJustSayNo) {
      console.log("[game-ui] demand action: deny blocked (missing JUST_SAY_NO)");
      return;
    }

    console.log("[game-ui] demand action: just-say-no");
    onDenyDemand(demandId);
  }, [hasJustSayNo, onDenyDemand]);

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
        console.log("[game-ui] debt collector target blocked (cannot select self)", {
          cardId: debtCollectorPicker.cardId,
          targetPlayerId,
        });
        return;
      }

      console.log("[game-ui] debt collector target selected", {
        cardId: debtCollectorPicker.cardId,
        mode: debtCollectorPicker.mode,
        targetPlayerId,
      });

      if (debtCollectorPicker.mode === "sly_deal") {
        setSlyDealPicker({
          mode: "sly_deal",
          cardId: debtCollectorPicker.cardId,
          targetPlayerId,
        });
      } else if (debtCollectorPicker.mode === "forced_deal") {
        setSlyDealPicker({
          mode: "forced_deal",
          cardId: debtCollectorPicker.cardId,
          targetPlayerId,
        });
      } else if (debtCollectorPicker.mode === "wild_rent") {
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
      console.log("[game-ui] pending rent action: double blocked (missing DOUBLE_THE_RENT)");
      return;
    }

    console.log("[game-ui] pending rent action: double");
    onPlayDoubleTheRentCard();
  }, [hasDoubleTheRent, onPlayDoubleTheRentCard]);

  const onPendingRentRent = useCallback(() => {
    console.log("[game-ui] pending rent action: rent");
    onResolvePendingRent();
  }, [onResolvePendingRent]);

  const slyDealTargetPlayer = useMemo(() => {
    if (!slyDealPicker) {
      return undefined;
    }

    return players.find((player) => player.playerId === slyDealPicker.targetPlayerId);
  }, [players, slyDealPicker]);

  const slyDealStealableSets = useMemo(() => {
    if (!slyDealPicker) {
      return [] as PropertySet[];
    }

    const candidateSets = propertySetsByPlayer[slyDealPicker.targetPlayerId] ?? [];
    return candidateSets
      .filter((propertySet) => !isPropertySetComplete(propertySet))
      .map((propertySet) => {
        return {
          ...propertySet,
          cards: propertySet.cards.filter((card) => isPropertyCard(card)),
        };
      })
      .filter((propertySet) => propertySet.cards.length > 0);
  }, [propertySetsByPlayer, slyDealPicker]);

  const forcedDealSourceSets = useMemo(() => {
    if (!slyDealPicker || slyDealPicker.mode !== "forced_deal") {
      return [] as PropertySet[];
    }

    return selfPropertySets
      .filter((propertySet) => !isPropertySetComplete(propertySet))
      .map((propertySet) => {
        return {
          ...propertySet,
          cards: propertySet.cards.filter((card) => isPropertyCard(card)),
        };
      })
      .filter((propertySet) => propertySet.cards.length > 0);
  }, [selfPropertySets, slyDealPicker]);

  const onSelectSlyDealCard = useCallback(
    (selectedCardId: string) => {
      if (!slyDealPicker) {
        return;
      }

      if (slyDealPicker.mode === "forced_deal") {
        if (!slyDealPicker.targetCardId) {
          setSlyDealPicker((current) => {
            if (!current || current.mode !== "forced_deal") {
              return current;
            }

            return {
              ...current,
              targetCardId: selectedCardId,
            };
          });
          return;
        }

        onPlayForcedDealCard(
          slyDealPicker.cardId,
          slyDealPicker.targetPlayerId,
          selectedCardId,
          slyDealPicker.targetCardId,
        );
        setSlyDealPicker(null);
        return;
      }

      onPlaySlyDealCard(slyDealPicker.cardId, slyDealPicker.targetPlayerId, selectedCardId);
      setSlyDealPicker(null);
    },
    [onPlayForcedDealCard, onPlaySlyDealCard, slyDealPicker],
  );

  useEffect(() => {
    if (isSelectingPaymentCards && !isSelectedPaymentDemandActive) {
      setIsSelectingPaymentCards(false);
      setActivePaymentDemandId(null);
      setSelectedPaymentCardIds(new Set());
    }
  }, [isSelectedPaymentDemandActive, isSelectingPaymentCards]);

  const renderPlayerBoard = (player: Player) => {
    const playerMoney = moneyByPlayer[player.playerId] ?? [];
    const propertySets = propertySetsByPlayer[player.playerId] ?? [];
    const isCurrentPlayer = player.playerId === currentPlayerId;
    const isSelfBoard = !!selfPlayerId && player.playerId === selfPlayerId;
    const canInteractWithBoard = isSelfBoard && isSelfTurn && !isDiscardRequired;
    const canSelectBoardCards =
      isSelfBoard && isSelectingPaymentCards && isSelectedPaymentDemandActive;
    const dropHandlers = canInteractWithBoard
      ? {
          onDragOver: onAllowDrop,
          onDropMoney,
          onDropPropertyRoot: onDropProperty(),
          onDropPropertySet: (propertySetId: string) => onDropProperty(propertySetId),
        }
      : {
          onDragOver: undefined,
          onDropMoney: undefined,
          onDropPropertyRoot: undefined,
          onDropPropertySet: (_propertySetId: string) => undefined,
        };

    const canRearrangeFromBoard =
      isSelfBoard && isSelfTurn && !isSelectingPaymentCards && !isDiscardRequired;

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
              ${player.money} total · {player.completedSets} sets · {player.handCards} in hand
            </p>
          </div>
        </header>

        <div className="md-player-board__body">
          <div
            className={
              draggingCard
                ? "md-dropzone md-dropzone--money is-drag-active"
                : "md-dropzone md-dropzone--money"
            }
            onDragOver={dropHandlers.onDragOver}
            onDrop={dropHandlers.onDropMoney}
          >
            <GameCardStackBox
              title="Money"
              cards={playerMoney}
              assetImageByKey={assetImageByKey}
              layout="stack"
              emptyLabel="Drop money here"
              selectableCards={canSelectBoardCards}
              selectedCardIds={selectedPaymentCardIds}
              onCardClick={onToggleSelectedPaymentCard}
              isDragActive={!!draggingCard}
            />
          </div>

          {propertySets.map((propertySet, index) => (
            <div
              className={
                draggingCard
                  ? "md-dropzone md-dropzone--property-set is-drag-active"
                  : "md-dropzone md-dropzone--property-set"
              }
              key={propertySet.propertySetId}
              onDragOver={dropHandlers.onDragOver}
              onDrop={dropHandlers.onDropPropertySet(propertySet.propertySetId)}
            >
              <GameCardStackBox
                title={`Set ${index + 1}`}
                cards={propertySet.cards}
                assetImageByKey={assetImageByKey}
                layout="stack"
                color={propertySet.color}
                emptyLabel="Empty"
                selectableCards={canSelectBoardCards}
                selectedCardIds={selectedPaymentCardIds}
                onCardClick={onToggleSelectedPaymentCard}
                draggableCards={canRearrangeFromBoard}
                canDragCard={(card) => {
                  return !isPropertySetComplete(propertySet) && isPropertyCard(card);
                }}
                onCardDragStart={(card) => {
                  setDraggingCard({
                    cardId: card.cardId,
                    source: "board",
                    sourcePropertySetId: propertySet.propertySetId,
                  });
                }}
                onCardDragEnd={() => setDraggingCard(null)}
                isDragActive={!!draggingCard}
              />
            </div>
          ))}

          <div
            className={
              draggingCard
                ? "md-dropzone md-dropzone--property-new is-drag-active"
                : "md-dropzone md-dropzone--property-new"
            }
            onDragOver={dropHandlers.onDragOver}
            onDrop={dropHandlers.onDropPropertyRoot}
          >
            <div className="md-property-empty">Drop property here to start a new set</div>
          </div>
        </div>
      </article>
    );
  };

  return (
    <div className="md-html-board">
      <section
        className={[
          "md-board-viewport",
          isPanning ? "is-panning" : "",
          shouldDimBoard ? "md-board-viewport--demand-active" : "",
        ]
          .filter(Boolean)
          .join(" ")}
        ref={viewportRef}
        onPointerDown={onBoardPointerDown}
        onPointerMove={onBoardPointerMove}
        onPointerUp={stopPan}
        onPointerCancel={stopPan}
      >
        <div className="md-demand-overlay-wrap">
          {hasPendingRent && gameState?.pendingRent ? (
            <PendingRentOverlay
              pendingRent={gameState.pendingRent}
              players={players}
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
              targetCardImageById={propertyCardImageById}
              canDeny={hasJustSayNo}
              isSelectingCards={isSelectingPaymentCards}
              selectingDemandId={activePaymentDemandId ?? undefined}
              canConfirmSelection={canConfirmPaymentSelection}
              onComply={onDemandComply}
              onDeny={onDemandDeny}
            />
          )}
        </div>
        <div
          className={
            shouldDimBoard
              ? "md-board-content md-board-content--demand-active"
              : "md-board-content"
          }
          ref={contentRef}
          style={boardStyle}
        >
          {players.map((player) => renderPlayerBoard(player))}
        </div>
      </section>

      <section className="md-hand-row">
        <div
          className={
            draggingCard
              ? "md-dropzone md-dropzone--action is-drag-active"
              : "md-dropzone md-dropzone--action"
          }
          onDragOver={onAllowDrop}
          onDrop={onDropPassGo}
        >
          <GameCardStackBox
            title="Action pile"
            cards={lastActionCards}
            assetImageByKey={assetImageByKey}
            layout="spread"
            emptyLabel="Drop action cards here"
            isDragActive={!!draggingCard}
          />
        </div>

        <div className="md-hand-main">
          <GameCardStackBox
            title="Your hand"
            cards={yourHand}
            assetImageByKey={assetImageByKey}
            layout="spread"
            emptyLabel="Waiting for cards"
            draggableCards={!isSelectingPaymentCards && !isDiscardRequired}
            selectableCards={isDiscardRequired}
            selectedCardIds={selectedDiscardCardIds}
            onCardClick={(card) => onToggleDiscardCard(card.cardId)}
            onCardDragStart={(card) =>
              setDraggingCard({
                cardId: card.cardId,
                source: "hand",
              })
            }
            onCardDragEnd={() => setDraggingCard(null)}
          />
        </div>
      </section>

      {colorPicker ? (
        <div
          className="md-color-picker-backdrop"
          role="presentation"
          onClick={() => setColorPicker(null)}
        >
          <div
            className="md-color-picker"
            role="dialog"
            aria-modal="true"
            aria-label="Choose property color"
            onClick={(event) => event.stopPropagation()}
          >
            <p className="md-color-picker__title">Choose a color</p>
            <div className="md-color-picker__swatches">
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
                      onPlayPropertyCard(colorPicker.cardId, colorPicker.propertySetId, color);
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
              : debtCollectorPicker.mode === "sly_deal"
                ? "Sly Deal"
                : debtCollectorPicker.mode === "forced_deal"
                  ? "Forced Deal"
                : "Debt Collector"
          }
          ariaLabel={
            debtCollectorPicker.mode === "wild_rent"
              ? "Choose wild rent target"
              : debtCollectorPicker.mode === "sly_deal"
                ? "Choose sly deal target"
                : debtCollectorPicker.mode === "forced_deal"
                  ? "Choose forced deal target"
                : "Choose debt collector target"
          }
          onSelectPlayer={onSelectDebtCollectorTarget}
          onClose={() => setDebtCollectorPicker(null)}
        />
      ) : null}

      {slyDealPicker ? (
        <SlyDealPropertyPickerOverlay
          eyebrow={slyDealPicker.mode === "forced_deal" ? "Forced Deal" : "Sly Deal"}
          title={
            slyDealPicker.mode === "forced_deal"
              ? slyDealPicker.targetCardId
                ? "Choose one of your properties to swap"
                : `Choose a property from ${slyDealTargetPlayer?.displayName ?? "target"}`
              : `Choose a property from ${slyDealTargetPlayer?.displayName ?? "target"}`
          }
          targetPlayer={slyDealTargetPlayer}
          propertySets={
            slyDealPicker.mode === "forced_deal" && slyDealPicker.targetCardId
              ? forcedDealSourceSets
              : slyDealStealableSets
          }
          assetImageByKey={assetImageByKey}
          emptyText={
            slyDealPicker.mode === "forced_deal" && slyDealPicker.targetCardId
              ? "No swappable property cards in your sets."
              : "No stealable property cards."
          }
          onSelectCard={onSelectSlyDealCard}
          onClose={() => setSlyDealPicker(null)}
        />
      ) : null}
    </div>
  );
};

export default MonopolyDealHtmlBoard;
