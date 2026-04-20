import { Container, Graphics, Sprite, Text, TextStyle, Texture } from "pixi.js";
import {
  AssetKey,
  assetKeyToJSON,
  Color,
  type Card,
} from "../../../generated/monopoly_deal";

type PaletteLike = {
  surface: { color: number; alpha: number };
  border: { color: number; alpha: number };
  foreground: { color: number; alpha: number };
  mutedForeground: { color: number; alpha: number };
  primary?: { color: number; alpha: number };
};

type HandCardView = {
  card: Card;
  container: Container;
  frame: Graphics;
  mask: Graphics;
  sprite: Sprite;
  fallback: Text;
};

type DragState = {
  view: HandCardView;
  pointerId: number;
  pointerOffsetX: number;
  pointerOffsetY: number;
};

type Bounds = {
  x: number;
  y: number;
  width: number;
  height: number;
};

const DEFAULT_CARD_ASPECT_RATIO = 1.6;

const COLOR_SWATCH_BY_VALUE: Record<number, number> = {
  [Color.COLOR_BROWN]: 0x572e2d,
  [Color.COLOR_SKY]: 0xbed6fd,
  [Color.COLOR_PINK]: 0xca3a85,
  [Color.COLOR_ORANGE]: 0xffa24d,
  [Color.COLOR_RED]: 0xf73c35,
  [Color.COLOR_YELLOW]: 0xf9ff2f,
  [Color.COLOR_GREEN]: 0x24692d,
  [Color.COLOR_BLUE]: 0x2800a0,
  [Color.COLOR_UTILITY]: 0xcbd6aa,
  [Color.COLOR_RAILROAD]: 0x000000,
};

const shouldFlipTwoColorWild = (card: Card): boolean => {
  if (card.colors.length !== 2) {
    return false;
  }

  const secondColor = card.colors[1];
  if (
    secondColor === Color.COLOR_UNSPECIFIED ||
    secondColor === Color.UNRECOGNIZED
  ) {
    return false;
  }

  return card.activeColor === secondColor;
};

const DEFAULT_PALETTE: PaletteLike = {
  surface: { color: 0x0a1320, alpha: 1 },
  border: { color: 0x1d2c3f, alpha: 1 },
  foreground: { color: 0xf8fbff, alpha: 1 },
  mutedForeground: { color: 0x9eb3c9, alpha: 1 },
};

export class HandCardsLayer {
  readonly container = new Container();

  private readonly handAreaBg = new Graphics();
  private readonly handViewport = new Container();
  private readonly handViewportMask = new Graphics();
  private readonly handHeadingStyle = new TextStyle({
    fill: DEFAULT_PALETTE.mutedForeground.color,
    fontFamily: "JetBrains Mono, SF Mono, Menlo, monospace",
    fontSize: 11,
    fontWeight: "700",
    letterSpacing: 0.8,
  });
  private readonly handHeading = new Text({
    text: "MY HAND",
    style: this.handHeadingStyle,
  });
  private readonly lastActionAreaBg = new Graphics();
  private readonly cardsLayer = new Container();
  private readonly lastActionMask = new Graphics();
  private readonly lastActionFrame = new Graphics();
  private readonly lastActionSprite = new Sprite(Texture.EMPTY);
  private readonly lastActionHeadingStyle = new TextStyle({
    fill: DEFAULT_PALETTE.mutedForeground.color,
    fontFamily: "JetBrains Mono, SF Mono, Menlo, monospace",
    fontSize: 11,
    fontWeight: "700",
    letterSpacing: 0.8,
  });
  private readonly lastActionHeading = new Text({
    text: "LAST PLAYED ACTION",
    style: this.lastActionHeadingStyle,
  });
  private readonly fallbackStyle = new TextStyle({
    fill: DEFAULT_PALETTE.foreground.color,
    fontFamily: "JetBrains Mono, SF Mono, Menlo, monospace",
    fontSize: 10,
    align: "center",
    wordWrap: true,
  });
  private readonly textureCache = new Map<string, Texture>();

  private handCardViews: HandCardView[] = [];
  private cards: Card[] = [];
  private lastAction: Card | undefined;
  private assetImageByKey: Record<number, string> = {};
  private palette: PaletteLike = DEFAULT_PALETTE;
  private layout = { width: 0, height: 0 };
  private buildVersion = 0;
  private cardAspectRatio = DEFAULT_CARD_ASPECT_RATIO;
  private handScrollX = 0;
  private handContentWidth = 0;
  private handAreaTopY = 0;
  private handViewportBounds = { x: 0, y: 0, width: 0, height: 0 };
  private dragState: DragState | null = null;
  private onDropMoneyCard: (cardId: string) => void = () => {};
  private onDropActionCard: (cardId: string) => void = () => {};
  private onDropPropertyCard: (
    cardId: string,
    propertySetId?: string,
    activeColor?: Color,
  ) => void = () => {};
  private lastActionDropBounds: Bounds | null = null;
  private getMoneyDropZone: () => Bounds | null = () => null;
  private resolvePropertyDrop:
    | ((
        pointerX: number,
        pointerY: number,
      ) => { propertySetId?: string } | null)
    | null = null;
  private readonly propertyColorPickerLayer = new Container();
  private propertyColorPicker: { container: Container } | null = null;

  constructor() {
    this.handViewport.mask = this.handViewportMask;
    this.handViewport.addChild(this.cardsLayer);
    this.lastActionSprite.mask = this.lastActionMask;
    this.handHeading.resolution = Math.min(window.devicePixelRatio || 1, 2);
    this.lastActionHeading.resolution = Math.min(
      window.devicePixelRatio || 1,
      2,
    );

    this.container.addChild(
      this.handAreaBg,
      this.lastActionAreaBg,
      this.handHeading,
      this.handViewport,
      this.handViewportMask,
      this.lastActionFrame,
      this.lastActionSprite,
      this.lastActionMask,
      this.lastActionHeading,
      this.propertyColorPickerLayer,
    );

    this.container.eventMode = "static";
    this.container.on("globalpointermove", (event) => {
      this.handleGlobalPointerMove(
        event.global.x,
        event.global.y,
        event.pointerId,
      );
    });
  }

  setMoneyDropInteraction(options: {
    onDropMoneyCard: (cardId: string) => void;
    onDropActionCard: (cardId: string) => void;
    onDropPropertyCard: (
      cardId: string,
      propertySetId?: string,
      activeColor?: Color,
    ) => void;
    getMoneyDropZone: () => Bounds | null;
    resolvePropertyDrop: (
      pointerX: number,
      pointerY: number,
    ) => { propertySetId?: string } | null;
  }) {
    this.onDropMoneyCard = options.onDropMoneyCard;
    this.onDropActionCard = options.onDropActionCard;
    this.onDropPropertyCard = options.onDropPropertyCard;
    this.getMoneyDropZone = options.getMoneyDropZone;
    this.resolvePropertyDrop = options.resolvePropertyDrop;
  }

  applyTheme(palette: PaletteLike) {
    this.palette = palette;
    this.fallbackStyle.fill = palette.foreground.color;
    this.handHeadingStyle.fill = palette.mutedForeground.color;
    this.lastActionHeadingStyle.fill = palette.mutedForeground.color;
    this.updateLayout(this.layout.width, this.layout.height);
  }

  async setCards(cards: Card[], assetImageByKey: Record<number, string>) {
    this.cards = cards;
    this.assetImageByKey = assetImageByKey;
    this.cardAspectRatio = DEFAULT_CARD_ASPECT_RATIO;
    this.handScrollX = 0;
    this.cardsLayer.removeChildren();
    this.handCardViews = [];

    const version = ++this.buildVersion;

    for (const card of cards) {
      const cardContainer = new Container();
      cardContainer.eventMode = "static";
      cardContainer.cursor = "grab";
      const frame = new Graphics();
      const mask = new Graphics();
      const sprite = new Sprite(Texture.EMPTY);
      const fallback = new Text({
        text: assetKeyToJSON(card.assetKey),
        style: this.fallbackStyle,
      });

      sprite.mask = mask;
      sprite.anchor.set(0.5);
      fallback.resolution = Math.min(window.devicePixelRatio || 1, 2);
      fallback.anchor.set(0.5);

      cardContainer.addChild(frame, sprite, mask, fallback);
      this.cardsLayer.addChild(cardContainer);
      this.handCardViews.push({
        card,
        container: cardContainer,
        frame,
        mask,
        sprite,
        fallback,
      });

      cardContainer.on("pointerdown", (event) => {
        this.handleCardPointerDown(
          cardContainer,
          event.global.x,
          event.global.y,
          event.pointerId,
        );
      });
      cardContainer.on("pointerup", (event) => {
        this.handleCardPointerUp(
          card,
          event.global.x,
          event.global.y,
          event.pointerId,
        );
      });
      cardContainer.on("pointerupoutside", (event) => {
        this.handleCardPointerUp(
          card,
          event.global.x,
          event.global.y,
          event.pointerId,
        );
      });

      const imageUrl = assetImageByKey[card.assetKey];
      if (!imageUrl) {
        continue;
      }

      try {
        let texture = this.textureCache.get(imageUrl);
        if (!texture) {
          texture = await this.loadTextureFromImageUrl(imageUrl);
          this.textureCache.set(imageUrl, texture);
        }

        if (version !== this.buildVersion) {
          return;
        }

        if (texture.width > 0 && texture.height > 0) {
          this.cardAspectRatio = texture.height / texture.width;
        }

        sprite.texture = texture;
        sprite.visible = true;
        fallback.visible = false;
      } catch (error) {
        console.warn("[game-ui] failed to load hand card image", {
          imageUrl,
          error,
        });
      }
    }

    this.updateLayout(this.layout.width, this.layout.height);
  }

  async setLastAction(lastAction: Card | undefined) {
    this.lastAction = lastAction;
    this.lastActionSprite.visible = false;

    if (!lastAction || lastAction.assetKey === AssetKey.ASSET_KEY_UNSPECIFIED) {
      this.updateLayout(this.layout.width, this.layout.height);
      return;
    }

    const imageUrl = this.assetImageByKey[lastAction.assetKey];
    if (!imageUrl) {
      this.updateLayout(this.layout.width, this.layout.height);
      return;
    }

    try {
      let texture = this.textureCache.get(imageUrl);
      if (!texture) {
        texture = await this.loadTextureFromImageUrl(imageUrl);
        this.textureCache.set(imageUrl, texture);
      }

      this.lastActionSprite.texture = texture;
      this.lastActionSprite.visible = true;
    } catch (error) {
      console.warn("[game-ui] failed to load last action image", {
        imageUrl,
        error,
      });
    }

    this.updateLayout(this.layout.width, this.layout.height);
  }

  updateLayout(boardWidth: number, boardHeight: number) {
    this.layout = { width: boardWidth, height: boardHeight };

    const cardWidth = Math.min(boardWidth * 0.12, 130);
    const cardHeight = cardWidth * this.cardAspectRatio;
    const minGap = 8;
    const maxGap = 16;
    const cardCount = Math.max(this.handCardViews.length, 1);
    const boardPaddingX = boardWidth * 0.04;
    const availableWidth = boardWidth * 0.92;

    const lastActionAreaWidth = Math.max(
      cardWidth + 24,
      Math.min(boardWidth * 0.2, 190),
    );
    const handToActionGap = 12;
    const availableHandWidth = Math.max(
      availableWidth - lastActionAreaWidth - handToActionGap,
      220,
    );

    const rawGap =
      (availableHandWidth - cardWidth * cardCount) / Math.max(cardCount - 1, 1);
    const gap = Math.min(Math.max(rawGap, minGap), maxGap);
    const totalWidth = cardWidth * cardCount + gap * Math.max(cardCount - 1, 0);
    const handStartX = boardPaddingX;
    const handEndX =
      boardWidth - boardPaddingX - lastActionAreaWidth - handToActionGap;
    const startX = Math.max(
      handStartX,
      Math.min((handStartX + handEndX - totalWidth) / 2, handEndX - totalWidth),
    );
    const cardY = boardHeight - cardHeight - boardHeight * 0.04;

    const handAreaPaddingX = 14;
    const handAreaPaddingTop = 8;
    const handAreaPaddingBottom = 12;
    const handHeadingGap = 8;
    const handAreaX = Math.max(startX - handAreaPaddingX, boardPaddingX);
    const handAreaWidth = Math.min(
      totalWidth + handAreaPaddingX * 2,
      handEndX - handAreaX,
    );
    const handAreaY =
      cardY - handAreaPaddingTop - this.handHeading.height - handHeadingGap;
    const handAreaHeight =
      cardHeight +
      handAreaPaddingTop +
      handAreaPaddingBottom +
      this.handHeading.height +
      handHeadingGap;
    this.handAreaTopY = handAreaY;

    this.handHeading.x = handAreaX + 10;
    this.handHeading.y = handAreaY + 8;

    const handViewportX = handAreaX + handAreaPaddingX;
    const handViewportY = cardY;
    const handViewportWidth = Math.max(handAreaWidth - handAreaPaddingX * 2, 0);
    const handViewportHeight = cardHeight;
    this.handViewportBounds = {
      x: handViewportX,
      y: handViewportY,
      width: handViewportWidth,
      height: handViewportHeight,
    };

    this.handViewport.x = handViewportX;
    this.handViewport.y = handViewportY;
    this.handViewportMask
      .clear()
      .roundRect(
        handViewportX,
        handViewportY,
        handViewportWidth,
        handViewportHeight,
        10,
      )
      .fill({ color: 0xffffff, alpha: 1 });

    const handAreaColor =
      this.palette.primary?.color ?? this.palette.surface.color;
    this.handAreaBg
      .clear()
      .roundRect(handAreaX, handAreaY, handAreaWidth, handAreaHeight, 14)
      .fill({ color: handAreaColor, alpha: 0.14 })
      .stroke({
        color: this.palette.border.color,
        alpha: Math.max(this.palette.border.alpha, 0.5),
        width: 1,
      });

    const lastActionX = handAreaX + handAreaWidth + handToActionGap;
    const lastActionY = handAreaY;
    const lastActionHeight = handAreaHeight;

    this.lastActionAreaBg
      .clear()
      .roundRect(
        lastActionX,
        lastActionY,
        lastActionAreaWidth,
        lastActionHeight,
        12,
      )
      .fill({
        color: this.palette.surface.color,
        alpha: Math.max(this.palette.surface.alpha, 0.86),
      })
      .stroke({
        color: this.palette.border.color,
        alpha: Math.max(this.palette.border.alpha, 0.55),
        width: 1,
      });

    this.lastActionHeading.x = lastActionX + 10;
    this.lastActionHeading.y = lastActionY + 8;

    const actionCardPaddingX = 10;
    const actionCardPaddingBottom = 10;
    const actionCardTop =
      this.lastActionHeading.y + this.lastActionHeading.height + 8;
    const maxActionCardWidth = Math.max(
      lastActionAreaWidth - actionCardPaddingX * 2,
      40,
    );
    const maxActionCardHeight = Math.max(
      lastActionHeight -
        (actionCardTop - lastActionY) -
        actionCardPaddingBottom,
      40,
    );
    const actionCardWidth = Math.min(
      maxActionCardWidth,
      maxActionCardHeight / this.cardAspectRatio,
    );
    const actionCardHeight = actionCardWidth * this.cardAspectRatio;
    const actionCardX =
      lastActionX +
      actionCardPaddingX +
      (maxActionCardWidth - actionCardWidth) / 2;

    this.lastActionMask
      .clear()
      .roundRect(
        actionCardX,
        actionCardTop,
        actionCardWidth,
        actionCardHeight,
        10,
      )
      .fill({ color: 0xffffff, alpha: 1 });

    this.lastActionFrame
      .clear()
      .roundRect(
        actionCardX,
        actionCardTop,
        actionCardWidth,
        actionCardHeight,
        10,
      )
      .stroke({
        color: this.palette.border.color,
        alpha: Math.max(this.palette.border.alpha, 0.55),
        width: 1,
      });

    this.lastActionSprite.x = actionCardX;
    this.lastActionSprite.y = actionCardTop;
    this.lastActionSprite.width = actionCardWidth;
    this.lastActionSprite.height = actionCardHeight;
    this.lastActionDropBounds = {
      x: actionCardX,
      y: actionCardTop,
      width: actionCardWidth,
      height: actionCardHeight,
    };
    this.lastActionSprite.visible =
      this.lastActionSprite.texture !== Texture.EMPTY &&
      !!this.lastAction &&
      this.lastAction.assetKey !== AssetKey.ASSET_KEY_UNSPECIFIED;

    this.handContentWidth = totalWidth;
    this.clampHandScroll();
    this.applyHandScroll();

    this.handCardViews.forEach((view, index) => {
      view.container.x = index * (cardWidth + gap);
      view.container.y = 0;

      view.mask
        .clear()
        .roundRect(0, 0, cardWidth, cardHeight, 10)
        .fill({ color: 0xffffff, alpha: 1 });

      if (view.fallback.visible) {
        view.frame
          .clear()
          .roundRect(0, 0, cardWidth, cardHeight, 10)
          .fill({
            color: this.palette.surface.color,
            alpha: this.palette.surface.alpha,
          })
          .stroke({
            color: this.palette.border.color,
            alpha: this.palette.border.alpha,
            width: 1,
          });
      } else {
        view.frame.clear();
      }

      const isFlipped = shouldFlipTwoColorWild(view.card);
      view.sprite.x = cardWidth / 2;
      view.sprite.y = cardHeight / 2;
      view.sprite.width = cardWidth;
      view.sprite.height = cardHeight;
      view.sprite.rotation = isFlipped ? Math.PI : 0;

      view.fallback.x = cardWidth / 2;
      view.fallback.y = cardHeight / 2;
      view.fallback.style.wordWrap = true;
      view.fallback.style.wordWrapWidth = cardWidth * 0.84;
    });
  }

  getHandAreaTopY() {
    return this.handAreaTopY;
  }

  isInsideHandArea(pointerX: number, pointerY: number): boolean {
    return (
      pointerX >= this.handViewportBounds.x &&
      pointerX <= this.handViewportBounds.x + this.handViewportBounds.width &&
      pointerY >= this.handViewportBounds.y &&
      pointerY <= this.handViewportBounds.y + this.handViewportBounds.height
    );
  }

  handleWheel(
    pointerX: number,
    pointerY: number,
    deltaX: number,
    deltaY: number,
  ): boolean {
    if (!this.isInsideHandArea(pointerX, pointerY)) {
      return false;
    }

    if (this.handContentWidth <= this.handViewportBounds.width) {
      return true;
    }

    const scrollDelta = Math.abs(deltaX) > Math.abs(deltaY) ? deltaX : deltaY;
    this.handScrollX += scrollDelta;
    this.clampHandScroll();
    this.applyHandScroll();
    return true;
  }

  private async loadTextureFromImageUrl(url: string): Promise<Texture> {
    const response = await fetch(url, {
      credentials: "include",
    });

    if (!response.ok) {
      throw new Error(`Failed to fetch image: ${url} (${response.status})`);
    }

    const blob = await response.blob();
    const objectUrl = URL.createObjectURL(blob);

    const image = await new Promise<HTMLImageElement>((resolve, reject) => {
      const nextImage = new Image();
      nextImage.onload = () => {
        URL.revokeObjectURL(objectUrl);
        resolve(nextImage);
      };
      nextImage.onerror = () => {
        URL.revokeObjectURL(objectUrl);
        reject(new Error(`Failed to decode image: ${url}`));
      };
      nextImage.src = objectUrl;
    });

    return Texture.from(image);
  }

  private clampHandScroll() {
    const overflow = Math.max(
      this.handContentWidth - this.handViewportBounds.width,
      0,
    );
    this.handScrollX = Math.min(Math.max(this.handScrollX, 0), overflow);
  }

  private applyHandScroll() {
    if (this.handContentWidth <= this.handViewportBounds.width) {
      this.cardsLayer.x =
        (this.handViewportBounds.width - this.handContentWidth) / 2;
      return;
    }

    this.cardsLayer.x = -this.handScrollX;
  }

  private handleCardPointerDown(
    cardContainer: Container,
    pointerX: number,
    pointerY: number,
    pointerId: number,
  ) {
    if (this.dragState) {
      return;
    }

    this.clearPropertyColorPicker();

    const globalPosition = cardContainer.getGlobalPosition();
    this.dragState = {
      view: this.handCardViews.find(
        (view) => view.container === cardContainer,
      )!,
      pointerId,
      pointerOffsetX: pointerX - globalPosition.x,
      pointerOffsetY: pointerY - globalPosition.y,
    };

    this.cardsLayer.removeChild(cardContainer);
    this.container.addChild(cardContainer);
    cardContainer.x = globalPosition.x;
    cardContainer.y = globalPosition.y;
    cardContainer.zIndex = 200;
    cardContainer.alpha = 0.96;
  }

  private handleGlobalPointerMove(
    pointerX: number,
    pointerY: number,
    pointerId: number,
  ) {
    if (!this.dragState || this.dragState.pointerId !== pointerId) {
      return;
    }

    const { view, pointerOffsetX, pointerOffsetY } = this.dragState;
    view.container.x = pointerX - pointerOffsetX;
    view.container.y = pointerY - pointerOffsetY;
  }

  private handleCardPointerUp(
    card: Card,
    pointerX: number,
    pointerY: number,
    pointerId: number,
  ) {
    if (!this.dragState || this.dragState.pointerId !== pointerId) {
      return;
    }

    const dropZone = this.getMoneyDropZone();
    const propertyDrop = this.resolvePropertyDrop
      ? this.resolvePropertyDrop(pointerX, pointerY)
      : null;
    const isDroppedOnMoney =
      !!dropZone &&
      pointerX >= dropZone.x &&
      pointerX <= dropZone.x + dropZone.width &&
      pointerY >= dropZone.y &&
      pointerY <= dropZone.y + dropZone.height;
    const isDroppedOnProperty = propertyDrop !== null;
    const actionDropBounds = this.lastActionDropBounds;
    const isDroppedOnAction =
      !!actionDropBounds &&
      pointerX >= actionDropBounds.x &&
      pointerX <= actionDropBounds.x + actionDropBounds.width &&
      pointerY >= actionDropBounds.y &&
      pointerY <= actionDropBounds.y + actionDropBounds.height;

    const draggedView = this.dragState.view;
    this.dragState = null;
    draggedView.container.alpha = 1;
    draggedView.container.zIndex = 0;

    this.container.removeChild(draggedView.container);
    this.cardsLayer.addChild(draggedView.container);

    if (isDroppedOnMoney) {
      this.onDropMoneyCard(card.cardId);
    } else if (isDroppedOnAction) {
      this.onDropActionCard(card.cardId);
    } else if (isDroppedOnProperty) {
      const uniqueSelectableColors = Array.from(
        new Set(
          card.colors.filter((color) => {
            return (
              color !== Color.COLOR_UNSPECIFIED && color !== Color.UNRECOGNIZED
            );
          }),
        ),
      );
      const shouldChooseColor =
        !propertyDrop?.propertySetId && uniqueSelectableColors.length > 1;

      if (shouldChooseColor) {
        this.showPropertyColorPicker({
          card,
          propertySetId: propertyDrop?.propertySetId,
          selectableColors: uniqueSelectableColors,
          pointerX,
          pointerY,
        });
      } else {
        this.onDropPropertyCard(
          card.cardId,
          propertyDrop?.propertySetId,
          uniqueSelectableColors.length === 1
            ? uniqueSelectableColors[0]
            : undefined,
        );
      }
    }

    this.updateLayout(this.layout.width, this.layout.height);
  }

  private showPropertyColorPicker(options: {
    card: Card;
    propertySetId?: string;
    selectableColors: Color[];
    pointerX: number;
    pointerY: number;
  }) {
    this.clearPropertyColorPicker();

    const pickerContainer = new Container();
    const backdrop = new Graphics();
    backdrop.eventMode = "static";
    backdrop
      .clear()
      .rect(
        0,
        0,
        Math.max(this.layout.width, 1),
        Math.max(this.layout.height, 1),
      )
      .fill({ color: 0x000000, alpha: 0 });
    backdrop.on("pointertap", () => {
      this.clearPropertyColorPicker();
    });
    pickerContainer.addChild(backdrop);

    const swatchStack = new Container();
    const panel = new Container();
    const panelBg = new Graphics();
    const headingStyle = new TextStyle({
      fill: this.palette.mutedForeground.color,
      fontFamily: "JetBrains Mono, SF Mono, Menlo, monospace",
      fontSize: 10,
      fontWeight: "700",
      letterSpacing: 0.7,
    });
    const heading = new Text({ text: "CHOOSE COLOR", style: headingStyle });
    heading.resolution = Math.min(window.devicePixelRatio || 1, 2);

    const swatchBaseSize = 30;
    const swatchGap = 8;
    const panelPadding = 8;
    const headingGap = 8;
    const maxSwatchRowWidth = Math.max(
      this.layout.width - panelPadding * 2 - 16,
      swatchBaseSize,
    );
    const swatchSize = Math.max(
      18,
      Math.min(
        swatchBaseSize,
        Math.floor(
          (maxSwatchRowWidth -
            Math.max(options.selectableColors.length - 1, 0) * swatchGap) /
            Math.max(options.selectableColors.length, 1),
        ),
      ),
    );
    const swatchRowWidth =
      options.selectableColors.length * swatchSize +
      Math.max(options.selectableColors.length - 1, 0) * swatchGap;
    const panelWidth =
      Math.max(heading.width, swatchRowWidth) + panelPadding * 2;
    const panelHeight =
      panelPadding + heading.height + headingGap + swatchSize + panelPadding;

    const maxX = Math.max(this.layout.width - panelWidth - 8, 8);
    const maxY = Math.max(this.layout.height - panelHeight - 8, 8);
    panel.x = Math.min(Math.max(options.pointerX + 14, 8), maxX);
    panel.y = Math.min(Math.max(options.pointerY - panelHeight / 2, 8), maxY);

    panelBg
      .clear()
      .roundRect(0, 0, panelWidth, panelHeight, 10)
      .fill({
        color: this.palette.surface.color,
        alpha: Math.max(this.palette.surface.alpha, 0.92),
      })
      .stroke({
        color: this.palette.border.color,
        alpha: Math.max(this.palette.border.alpha, 0.8),
        width: 1,
      });

    heading.x =
      panelPadding + (panelWidth - panelPadding * 2 - heading.width) / 2;
    heading.y = panelPadding;
    swatchStack.x =
      panelPadding + (panelWidth - panelPadding * 2 - swatchRowWidth) / 2;
    swatchStack.y = panelPadding + heading.height + headingGap;

    options.selectableColors.forEach((color, index) => {
      const swatch = new Graphics();
      swatch.eventMode = "static";
      swatch.cursor = "pointer";
      swatch
        .clear()
        .roundRect(0, 0, swatchSize, swatchSize, 5)
        .fill({ color: COLOR_SWATCH_BY_VALUE[color] ?? 0x4a596f, alpha: 1 })
        .stroke({
          color: this.palette.border.color,
          alpha: Math.max(this.palette.border.alpha, 0.75),
          width: 1.2,
        });
      swatch.x = index * (swatchSize + swatchGap);
      swatch.y = 0;
      swatch.on("pointertap", () => {
        this.onDropPropertyCard(
          options.card.cardId,
          options.propertySetId,
          color,
        );
        this.clearPropertyColorPicker();
      });
      swatchStack.addChild(swatch);
    });

    panel.addChild(panelBg, heading, swatchStack);
    pickerContainer.addChild(panel);
    this.propertyColorPickerLayer.addChild(pickerContainer);
    this.propertyColorPicker = { container: pickerContainer };
  }

  private clearPropertyColorPicker() {
    if (!this.propertyColorPicker) {
      return;
    }

    this.propertyColorPicker.container.removeFromParent();
    this.propertyColorPicker.container.destroy({ children: true });
    this.propertyColorPicker = null;
  }
}
