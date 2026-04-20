// Layout note: use relative percentages and percentage-based offsets for most positioning/scaling,
// and avoid hard-coded pixel values except for fallback minimums.
import { useEffect, useRef } from "react";
import { Application, Container, Graphics } from "pixi.js";
import { type Card, type Color, type GameState } from "../../generated/monopoly_deal";
import { HandCardsLayer } from "./ui/HandCardsLayer";
import { PlayerBoardBoxesLayer } from "./ui/PlayerBoardBoxesLayer";

type PixiColor = {
  color: number;
  alpha: number;
};

type GameThemePalette = {
  background: PixiColor;
  surface: PixiColor;
  card: PixiColor;
  muted: PixiColor;
  border: PixiColor;
  foreground: PixiColor;
  mutedForeground: PixiColor;
  primary: PixiColor;
  primaryForeground: PixiColor;
};

const parseHexColor = (value: string): PixiColor | null => {
  const normalized = value.trim().toLowerCase();
  if (!normalized.startsWith("#")) {
    return null;
  }

  const hex = normalized.slice(1);
  if (hex.length === 3) {
    const r = parseInt(`${hex[0]}${hex[0]}`, 16);
    const g = parseInt(`${hex[1]}${hex[1]}`, 16);
    const b = parseInt(`${hex[2]}${hex[2]}`, 16);
    return { color: (r << 16) | (g << 8) | b, alpha: 1 };
  }

  if (hex.length === 6) {
    return { color: parseInt(hex, 16), alpha: 1 };
  }

  if (hex.length === 8) {
    const color = parseInt(hex.slice(0, 6), 16);
    const alpha = parseInt(hex.slice(6, 8), 16) / 255;
    return { color, alpha };
  }

  return null;
};

const getThemePalette = (): GameThemePalette => {
  const computedStyle = getComputedStyle(document.documentElement);

  const readColor = (name: string, fallback: string): PixiColor => {
    const raw = computedStyle.getPropertyValue(name).trim() || fallback;
    return (
      parseHexColor(raw) ?? parseHexColor(fallback) ?? { color: 0, alpha: 1 }
    );
  };

  return {
    background: readColor("--background", "#06080c"),
    surface: readColor("--surface", "#0a1320"),
    card: readColor("--card", "#0b1018"),
    muted: readColor("--muted", "#121a26"),
    border: readColor("--border", "#1d2c3f"),
    foreground: readColor("--foreground", "#f8fbff"),
    mutedForeground: readColor("--muted-foreground", "#9eb3c9"),
    primary: readColor("--primary", "#36b6ff"),
    primaryForeground: readColor("--primary-foreground", "#06111c"),
  };
};

type SyncSignatures = {
  assetMap: string;
  players: string;
  currentPlayerId: string;
  money: string;
  properties: string;
  hand: string;
  lastAction: string;
};

const EMPTY_SYNC_SIGNATURES: SyncSignatures = {
  assetMap: "",
  players: "",
  currentPlayerId: "",
  money: "",
  properties: "",
  hand: "",
  lastAction: "",
};

const toCardSignature = (card: Card): string => {
  return [
    card.cardId,
    String(card.assetKey),
    String(card.category),
    String(card.activeColor),
    card.colors.join(","),
    String(card.value),
  ].join("|");
};

const toAssetMapSignature = (assetMap: Record<number, string>): string => {
  return Object.entries(assetMap)
    .sort((left, right) => Number(left[0]) - Number(right[0]))
    .map(([assetKey, imageUrl]) => `${assetKey}:${imageUrl}`)
    .join("||");
};

const toSyncSignatures = (
  gameState: GameState | null,
  assetMap: Record<number, string>,
): SyncSignatures => {
  const playersSignature = (gameState?.players ?? [])
    .map((player) => {
      return [
        player.playerId,
        player.displayName,
        player.avatarUrl,
        String(player.money),
        String(player.completedSets),
        String(player.handCards),
      ].join("|");
    })
    .join("||");

  const moneySignature = (gameState?.money ?? [])
    .map((moneyPile) => {
      return [
        moneyPile.playerId,
        moneyPile.cards.map(toCardSignature).join(";;"),
      ].join("|");
    })
    .join("||");

  const propertiesSignature = (gameState?.properties ?? [])
    .map((propertySet) => {
      return [
        propertySet.playerId,
        propertySet.propertySetId,
        String(propertySet.color),
        propertySet.cards.map(toCardSignature).join(";;"),
      ].join("|");
    })
    .join("||");

  const handSignature = (gameState?.yourHand?.cards ?? [])
    .map(toCardSignature)
    .join(";;");
  const lastActionSignature = gameState?.lastAction
    ? toCardSignature(gameState.lastAction)
    : "";

  return {
    assetMap: toAssetMapSignature(assetMap),
    players: playersSignature,
    currentPlayerId: gameState?.currentPlayerId ?? "",
    money: moneySignature,
    properties: propertiesSignature,
    hand: handSignature,
    lastAction: lastActionSignature,
  };
};

type MonopolyDealGameMountProps = {
  initialGameState: GameState | null;
  assetImageByKey: Record<number, string>;
  onPlayMoneyCard: (cardId: string) => void;
  onPlayPassGoCard: (cardId: string) => void;
  onPlayPropertyCard: (
    cardId: string,
    propertySetId?: string,
    activeColor?: Color,
  ) => void;
};

const MonopolyDealGameMount = ({
  initialGameState,
  assetImageByKey,
  onPlayMoneyCard,
  onPlayPassGoCard,
  onPlayPropertyCard,
}: MonopolyDealGameMountProps) => {
  const mountRef = useRef<HTMLDivElement | null>(null);
  const appRef = useRef<Application | null>(null);
  const boardBgRef = useRef<Graphics | null>(null);
  const handCardsLayerRef = useRef<HandCardsLayer | null>(null);
  const playerBoxesLayerRef = useRef<PlayerBoardBoxesLayer | null>(null);
  const paletteRef = useRef<GameThemePalette>(getThemePalette());
  const onPlayMoneyCardRef = useRef(onPlayMoneyCard);
  const onPlayPassGoCardRef = useRef(onPlayPassGoCard);
  const onPlayPropertyCardRef = useRef(onPlayPropertyCard);
  const recenterRef = useRef<() => void>(() => {});
  const latestGameStateRef = useRef<GameState | null>(initialGameState);
  const latestAssetMapRef = useRef<Record<number, string>>(assetImageByKey);
  const syncLayersRef = useRef<() => Promise<void>>(async () => {});
  const lastSyncedSignaturesRef = useRef<SyncSignatures>(EMPTY_SYNC_SIGNATURES);

  useEffect(() => {
    onPlayMoneyCardRef.current = onPlayMoneyCard;
  }, [onPlayMoneyCard]);

  useEffect(() => {
    onPlayPassGoCardRef.current = onPlayPassGoCard;
  }, [onPlayPassGoCard]);

  useEffect(() => {
    onPlayPropertyCardRef.current = onPlayPropertyCard;
  }, [onPlayPropertyCard]);

  useEffect(() => {
    latestGameStateRef.current = initialGameState;
    latestAssetMapRef.current = assetImageByKey;

    void syncLayersRef.current();
  }, [initialGameState, assetImageByKey]);

  useEffect(() => {
    const mount = mountRef.current;
    if (!mount) {
      return;
    }

    const app = new Application();
    let initialized = false;
    let disposed = false;
    let cleanupResizeObserver = () => {};
    let cleanupThemeObserver = () => {};
    let cleanupPlayerBoardInteractions = () => {};

    void (async () => {
      try {
        await app.init({
          width: mount.clientWidth || 1024,
          height: mount.clientHeight || 640,
          background: paletteRef.current.background.color,
          backgroundAlpha: paletteRef.current.background.alpha,
          antialias: true,
          autoDensity: true,
          resolution: Math.min(window.devicePixelRatio || 1, 2),
          roundPixels: true,
        });
      } catch (error) {
        console.error("[game-ui] failed to initialize pixi app", error);
        return;
      }

      initialized = true;

      if (disposed) {
        app.destroy();
        return;
      }

      appRef.current = app;
      mount.appendChild(app.canvas);

      const root = new Container();
      app.stage.addChild(root);

      const boardBg = new Graphics();
      const handCardsLayer = new HandCardsLayer();
      const playerBoxesLayer = new PlayerBoardBoxesLayer();

      boardBgRef.current = boardBg;
      handCardsLayerRef.current = handCardsLayer;
      playerBoxesLayerRef.current = playerBoxesLayer;

      handCardsLayer.setMoneyDropInteraction({
        onDropMoneyCard: (cardId) => {
          onPlayMoneyCardRef.current(cardId);
        },
        onDropActionCard: (cardId) => {
          onPlayPassGoCardRef.current(cardId);
        },
        onDropPropertyCard: (cardId, propertySetId, activeColor) => {
          onPlayPropertyCardRef.current(cardId, propertySetId, activeColor);
        },
        getMoneyDropZone: () => playerBoxesLayer.getCurrentPlayerMoneyDropZone(),
        resolvePropertyDrop: (pointerX, pointerY) =>
          playerBoxesLayer.resolveCurrentPlayerPropertyDrop(pointerX, pointerY),
      });

      root.addChild(boardBg);
      root.addChild(playerBoxesLayer.container);
      root.addChild(handCardsLayer.container);

      const recenter = () => {
        const activeApp = appRef.current;
        const activeBoardBg = boardBgRef.current;
        const activeHandLayer = handCardsLayerRef.current;
        const activePlayerLayer = playerBoxesLayerRef.current;

        if (
          disposed ||
          !activeApp ||
          !activeBoardBg ||
          !activeHandLayer ||
          !activePlayerLayer ||
          !(activeApp as { renderer?: unknown }).renderer
        ) {
          return;
        }

        const width = activeApp.renderer.width;
        const height = activeApp.renderer.height;

        activeBoardBg.clear().rect(0, 0, width, height).fill({
          color: paletteRef.current.background.color,
          alpha: paletteRef.current.background.alpha,
        });

        activeHandLayer.updateLayout(width, height);

        const playerAreaTop = 0;
        const playerAreaBottom = activeHandLayer.getHandAreaTopY() - 8;
        const playerAreaHeight = Math.max(playerAreaBottom - playerAreaTop, 0);

        activePlayerLayer.updateLayout({
          x: 0,
          y: playerAreaTop,
          width,
          height: playerAreaHeight,
        });
      };

      recenterRef.current = recenter;

      syncLayersRef.current = async () => {
        const handCardsLayer = handCardsLayerRef.current;
        const playerBoxesLayer = playerBoxesLayerRef.current;
        const gameState = latestGameStateRef.current;
        const currentAssetMap = latestAssetMapRef.current;

        if (!handCardsLayer || !playerBoxesLayer) {
          return;
        }

        const nextSignatures = toSyncSignatures(gameState, currentAssetMap);
        const previousSignatures = lastSyncedSignaturesRef.current;
        const didAssetMapChange =
          nextSignatures.assetMap !== previousSignatures.assetMap;
        const shouldSyncBoard =
          didAssetMapChange ||
          nextSignatures.players !== previousSignatures.players ||
          nextSignatures.currentPlayerId !== previousSignatures.currentPlayerId ||
          nextSignatures.money !== previousSignatures.money ||
          nextSignatures.properties !== previousSignatures.properties;
        const shouldSyncHand =
          didAssetMapChange || nextSignatures.hand !== previousSignatures.hand;
        const shouldSyncLastAction =
          didAssetMapChange ||
          nextSignatures.lastAction !== previousSignatures.lastAction;

        if (shouldSyncBoard) {
          await playerBoxesLayer.setPlayers(
            gameState?.players ?? [],
            gameState?.currentPlayerId,
            gameState?.money ?? [],
            gameState?.properties ?? [],
            currentAssetMap,
          );
        }

        if (shouldSyncHand) {
          await handCardsLayer.setCards(
            gameState?.yourHand?.cards ?? [],
            currentAssetMap,
          );
        }

        if (shouldSyncLastAction) {
          await handCardsLayer.setLastAction(gameState?.lastAction);
        }

        lastSyncedSignaturesRef.current = nextSignatures;
        recenterRef.current();
      };

      const applyTheme = () => {
        const activeApp = appRef.current;
        const activeHandLayer = handCardsLayerRef.current;
        const activePlayerLayer = playerBoxesLayerRef.current;

        if (
          disposed ||
          !activeApp ||
          !activeHandLayer ||
          !activePlayerLayer ||
          !(activeApp as { renderer?: unknown }).renderer
        ) {
          return;
        }

        paletteRef.current = getThemePalette();
        activePlayerLayer.applyTheme(paletteRef.current);
        activeHandLayer.applyTheme(paletteRef.current);

        const backgroundSystem = (
          activeApp.renderer as {
            background?: {
              color: number;
              alpha: number;
            };
          }
        ).background;
        if (backgroundSystem) {
          backgroundSystem.color = paletteRef.current.background.color;
          backgroundSystem.alpha = paletteRef.current.background.alpha;
        }

        recenterRef.current();
      };

      const resize = () => {
        const activeApp = appRef.current;
        if (
          disposed ||
          !activeApp ||
          !(activeApp as { renderer?: unknown }).renderer
        ) {
          return;
        }

        const width = mount.clientWidth || 1;
        const height = mount.clientHeight || 1;
        activeApp.renderer.resize(width, height);
        recenterRef.current();
      };

      const resizeObserver = new ResizeObserver(resize);
      resizeObserver.observe(mount);

      let isPanningPlayerBoard = false;
      let lastPanX = 0;
      let lastPanY = 0;

      const getCanvasPoint = (event: PointerEvent | WheelEvent) => {
        const rect = app.canvas.getBoundingClientRect();
        return {
          x: event.clientX - rect.left,
          y: event.clientY - rect.top,
        };
      };

      const handleWheel = (event: WheelEvent) => {
        const handLayer = handCardsLayerRef.current;
        const playerLayer = playerBoxesLayerRef.current;
        if (!handLayer || !playerLayer) {
          return;
        }

        const point = getCanvasPoint(event);
        const handledByPlayers = playerLayer.handleWheel(
          point.x,
          point.y,
          event.deltaY,
        );
        const handledByHand = handLayer.handleWheel(
          point.x,
          point.y,
          event.deltaX,
          event.deltaY,
        );
        if (handledByPlayers || handledByHand) {
          event.preventDefault();
        }
      };

      const handlePointerDown = (event: PointerEvent) => {
        if (event.button !== 0) {
          return;
        }

        const playerLayer = playerBoxesLayerRef.current;
        if (!playerLayer) {
          return;
        }

        const point = getCanvasPoint(event);
        if (!playerLayer.isInsideBounds(point.x, point.y)) {
          return;
        }

        isPanningPlayerBoard = true;
        lastPanX = point.x;
        lastPanY = point.y;
        app.canvas.style.cursor = "grabbing";
      };

      const handlePointerMove = (event: PointerEvent) => {
        const playerLayer = playerBoxesLayerRef.current;
        if (!playerLayer) {
          return;
        }

        const point = getCanvasPoint(event);

        if (!isPanningPlayerBoard) {
          app.canvas.style.cursor = playerLayer.isInsideBounds(point.x, point.y)
            ? "grab"
            : "default";
          return;
        }

        const deltaX = point.x - lastPanX;
        const deltaY = point.y - lastPanY;
        lastPanX = point.x;
        lastPanY = point.y;
        playerLayer.panBy(deltaX, deltaY);
      };

      const stopPan = () => {
        isPanningPlayerBoard = false;
        app.canvas.style.cursor = "default";
      };

      app.canvas.addEventListener("wheel", handleWheel, { passive: false });
      app.canvas.addEventListener("pointerdown", handlePointerDown);
      app.canvas.addEventListener("pointermove", handlePointerMove);
      window.addEventListener("pointerup", stopPan);
      window.addEventListener("pointercancel", stopPan);

      cleanupPlayerBoardInteractions = () => {
        app.canvas.removeEventListener("wheel", handleWheel);
        app.canvas.removeEventListener("pointerdown", handlePointerDown);
        app.canvas.removeEventListener("pointermove", handlePointerMove);
        window.removeEventListener("pointerup", stopPan);
        window.removeEventListener("pointercancel", stopPan);
      };

      const themeObserver = new MutationObserver(() => {
        applyTheme();
      });
      themeObserver.observe(document.documentElement, {
        attributes: true,
        attributeFilter: ["data-theme", "style", "class"],
      });

      cleanupResizeObserver = () => {
        resizeObserver.disconnect();
      };
      cleanupThemeObserver = () => {
        themeObserver.disconnect();
      };

      applyTheme();
      resize();
      await syncLayersRef.current();
    })();

    return () => {
      disposed = true;
      cleanupResizeObserver();
      cleanupThemeObserver();
      cleanupPlayerBoardInteractions();

      const currentApp = appRef.current;
      if (initialized && currentApp) {
        currentApp.destroy();
      }

      appRef.current = null;
      boardBgRef.current = null;
      handCardsLayerRef.current = null;
      playerBoxesLayerRef.current = null;
      recenterRef.current = () => {};
      syncLayersRef.current = async () => {};
      lastSyncedSignaturesRef.current = EMPTY_SYNC_SIGNATURES;

      if (mount.firstChild) {
        mount.replaceChildren();
      }
    };
  }, []);

  return <div className="game-mount" ref={mountRef} />;
};

export default MonopolyDealGameMount;
