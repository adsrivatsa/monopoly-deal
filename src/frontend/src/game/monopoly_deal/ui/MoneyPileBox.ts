import { Container, Graphics, Sprite, Text, TextStyle, Texture } from "pixi.js";
import { Color, type Card } from "../../../generated/monopoly_deal";

type PaletteLike = {
  card: { color: number; alpha: number };
  muted: { color: number; alpha: number };
  border: { color: number; alpha: number };
  mutedForeground: { color: number; alpha: number };
};

type Bounds = {
  x: number;
  y: number;
  width: number;
  height: number;
};

type MoneyCardItemView = {
  card: Card;
  container: Container;
  frame: Graphics;
  mask: Graphics;
  sprite: Sprite;
  fallback: Text;
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
  card: { color: 0x0b1018, alpha: 1 },
  muted: { color: 0x121a26, alpha: 1 },
  border: { color: 0x1d2c3f, alpha: 1 },
  mutedForeground: { color: 0x9eb3c9, alpha: 1 },
};

export class MoneyPileBox {
  readonly container = new Container();

  private readonly bg = new Graphics();
  private readonly titleStyle = new TextStyle({
    fill: DEFAULT_PALETTE.mutedForeground.color,
    fontFamily: "JetBrains Mono, SF Mono, Menlo, monospace",
    fontSize: 10,
    fontWeight: "700",
    letterSpacing: 0.7,
  });
  private readonly title = new Text({ text: "MONEY", style: this.titleStyle });
  private readonly stack = new Container();
  private readonly fallbackStyle = new TextStyle({
    fill: DEFAULT_PALETTE.mutedForeground.color,
    fontFamily: "JetBrains Mono, SF Mono, Menlo, monospace",
    fontSize: 11,
  });

  private readonly textureCache = new Map<string, Texture>();
  private palette: PaletteLike = DEFAULT_PALETTE;
  private cards: Card[] = [];
  private assetImageByKey: Record<number, string> = {};
  private items: MoneyCardItemView[] = [];
  private bounds: Bounds = { x: 0, y: 0, width: 0, height: 0 };
  private buildVersion = 0;
  private hoverIndex: number | null = null;
  private cardXPositions: number[] = [];
  private cardWidth = 0;
  private cardHeight = 0;
  private cardAreaY = 0;
  private cardAreaHeight = 0;

  constructor() {
    this.title.resolution = Math.min(window.devicePixelRatio || 1, 2);
    this.stack.sortableChildren = true;

    this.bg.eventMode = "static";
    this.bg.on("pointermove", (event) => {
      const local = this.container.toLocal(event.global);
      this.updateHoverFromPointer(local.x, local.y);
    });
    this.bg.on("pointerout", () => {
      if (this.hoverIndex !== null) {
        this.hoverIndex = null;
        this.updateLayout(this.bounds);
      }
    });

    this.container.addChild(this.bg, this.title, this.stack);
  }

  applyTheme(palette: PaletteLike) {
    this.palette = palette;
    this.titleStyle.fill = palette.mutedForeground.color;
    this.updateLayout(this.bounds);
  }

  async setCards(cards: Card[], assetImageByKey: Record<number, string>) {
    this.cards = cards;
    this.assetImageByKey = assetImageByKey;
    this.stack.removeChildren();
    this.items = [];

    const version = ++this.buildVersion;
    const stackCards = cards;
    for (const card of stackCards) {
      const cardContainer = new Container();
      const frame = new Graphics();
      const mask = new Graphics();
      const sprite = new Sprite(Texture.EMPTY);
      const fallback = new Text({ text: "?", style: this.fallbackStyle });
      fallback.anchor.set(0.5);
      fallback.resolution = Math.min(window.devicePixelRatio || 1, 2);
      sprite.anchor.set(0.5);
      sprite.mask = mask;

      cardContainer.addChild(frame, sprite, mask, fallback);
      this.stack.addChild(cardContainer);
      const item: MoneyCardItemView = {
        card,
        container: cardContainer,
        frame,
        mask,
        sprite,
        fallback,
      };
      this.items.push(item);

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

        sprite.texture = texture;
        sprite.visible = true;
        fallback.visible = false;
      } catch {
        // Keep fallback visible.
      }
    }

    this.updateLayout(this.bounds);
  }

  updateLayout(bounds: Bounds) {
    this.bounds = bounds;

    this.bg
      .clear()
      .roundRect(bounds.x, bounds.y, bounds.width, bounds.height, 9)
      .fill({
        color: this.palette.muted.color,
        alpha: Math.max(this.palette.muted.alpha, 0.88),
      })
      .stroke({
        color: this.palette.border.color,
        alpha: this.palette.border.alpha,
        width: 1,
      });

    this.title.x = bounds.x + 8;
    this.title.y = bounds.y + 6;

    const moneyCardAreaX = bounds.x + 8;
    const moneyCardAreaY = bounds.y + 22;
    const moneyCardAreaWidth = Math.max(bounds.width - 16, 20);
    const moneyCardAreaHeight = Math.max(bounds.height - 28, 16);

    const stackCount = this.items.length;
    if (stackCount === 0) {
      return;
    }

    const cardAspectRatio = 1.6;
    const widthByHeight = moneyCardAreaHeight / cardAspectRatio;
    const moneyCardWidth = Math.max(
      Math.min(widthByHeight, moneyCardAreaWidth),
      20,
    );
    const moneyCardHeight = moneyCardWidth * cardAspectRatio;
    const totalExpandedWidth = moneyCardWidth * stackCount;
    const cardY = moneyCardAreaY + (moneyCardAreaHeight - moneyCardHeight) / 2;

    this.cardXPositions = [];
    this.cardWidth = moneyCardWidth;
    this.cardHeight = moneyCardHeight;
    this.cardAreaY = cardY;
    this.cardAreaHeight = moneyCardHeight;

    let xPositions: number[] = [];

    if (totalExpandedWidth <= moneyCardAreaWidth) {
      const extraGap =
        stackCount > 1
          ? Math.min(
              (moneyCardAreaWidth - totalExpandedWidth) / (stackCount - 1),
              8,
            )
          : 0;
      const startX = moneyCardAreaX;
      xPositions = Array.from({ length: stackCount }, (_, index) => {
        return startX + index * (moneyCardWidth + extraGap);
      });
    } else {
      const focusIndex = Math.min(
        Math.max(this.hoverIndex ?? stackCount - 1, 0),
        stackCount - 1,
      );
      const usableFocusWidth = Math.max(moneyCardAreaWidth - moneyCardWidth, 0);
      const focusX =
        stackCount > 1
          ? moneyCardAreaX + (focusIndex / (stackCount - 1)) * usableFocusWidth
          : moneyCardAreaX;

      xPositions = Array.from({ length: stackCount }, () => focusX);

      const minReveal = Math.max(moneyCardWidth * 0.035, 2);
      const maxReveal = Math.max(moneyCardWidth * 0.24, 8);
      const revealForDistance = (distance: number) => {
        const decay = Math.exp(-(distance - 1) * 0.65);
        return minReveal + (maxReveal - minReveal) * decay;
      };

      for (let index = focusIndex - 1; index >= 0; index -= 1) {
        const distance = focusIndex - index;
        xPositions[index] = xPositions[index + 1] - revealForDistance(distance);
      }

      for (let index = focusIndex + 1; index < stackCount; index += 1) {
        const distance = index - focusIndex;
        xPositions[index] = xPositions[index - 1] + revealForDistance(distance);
      }

      const applyBoundaryShift = () => {
        const minX = Math.min(...xPositions);
        const maxX = Math.max(...xPositions.map((x) => x + this.cardWidth));
        let shift = 0;
        if (minX < moneyCardAreaX) {
          shift = moneyCardAreaX - minX;
        } else if (maxX > moneyCardAreaX + moneyCardAreaWidth) {
          shift = moneyCardAreaX + moneyCardAreaWidth - maxX;
        }

        if (shift !== 0) {
          for (let i = 0; i < xPositions.length; i += 1) {
            xPositions[i] += shift;
          }
        }
      };

      const scaleAroundFocus = (factor: number) => {
        const pivot = xPositions[focusIndex];
        for (let i = 0; i < xPositions.length; i += 1) {
          xPositions[i] = pivot + (xPositions[i] - pivot) * factor;
        }
      };

      applyBoundaryShift();

      const currentSpan =
        Math.max(...xPositions.map((x) => x + this.cardWidth)) -
        Math.min(...xPositions);
      const minSpan = this.cardWidth;
      const maxSpan = moneyCardAreaWidth;

      if (currentSpan > maxSpan + 0.01 && currentSpan > minSpan + 0.01) {
        const factor = (maxSpan - minSpan) / (currentSpan - minSpan);
        scaleAroundFocus(Math.max(Math.min(factor, 1), 0.1));
        applyBoundaryShift();
      } else if (currentSpan < maxSpan * 0.92 && currentSpan > minSpan + 0.01) {
        const targetSpan = maxSpan * 0.95;
        const factor = (targetSpan - minSpan) / (currentSpan - minSpan);
        scaleAroundFocus(Math.max(factor, 1));
        applyBoundaryShift();
      }
    }

    this.stack.x = 0;
    this.stack.y = 0;
    this.cardXPositions = xPositions;
    this.items.forEach((item, itemIndex) => {
      const cardX = xPositions[itemIndex] ?? moneyCardAreaX;
      const cardYForItem = this.cardAreaY;

      item.container.zIndex =
        this.hoverIndex !== null && itemIndex === this.hoverIndex
          ? stackCount + 10
          : itemIndex;

      item.mask
        .clear()
        .roundRect(cardX, cardYForItem, this.cardWidth, this.cardHeight, 7)
        .fill({ color: 0xffffff, alpha: 1 });

      if (item.fallback.visible) {
        item.frame
          .clear()
          .roundRect(cardX, cardYForItem, this.cardWidth, this.cardHeight, 7)
          .fill({
            color: this.palette.card.color,
            alpha: this.palette.card.alpha,
          })
          .stroke({
            color: this.palette.border.color,
            alpha: this.palette.border.alpha,
            width: 1,
          });
      } else {
        item.frame.clear();
      }

      item.sprite.x = cardX + this.cardWidth / 2;
      item.sprite.y = cardYForItem + this.cardHeight / 2;
      item.sprite.width = this.cardWidth;
      item.sprite.height = this.cardHeight;
      const isFlipped = shouldFlipTwoColorWild(item.card);
      item.sprite.rotation = isFlipped ? Math.PI : 0;

      item.fallback.x = cardX + this.cardWidth / 2;
      item.fallback.y = cardYForItem + this.cardHeight / 2;
    });
  }

  private updateHoverFromPointer(pointerX: number, pointerY: number) {
    if (this.items.length === 0) {
      return;
    }

    const insideY =
      pointerY >= this.cardAreaY &&
      pointerY <= this.cardAreaY + this.cardAreaHeight;
    if (!insideY) {
      if (this.hoverIndex !== null) {
        this.hoverIndex = null;
        this.updateLayout(this.bounds);
      }
      return;
    }

    let bestIndex = 0;
    let bestDistance = Number.POSITIVE_INFINITY;
    for (let i = 0; i < this.cardXPositions.length; i += 1) {
      const centerX = this.cardXPositions[i] + this.cardWidth / 2;
      const distance = Math.abs(pointerX - centerX);
      if (distance < bestDistance) {
        bestDistance = distance;
        bestIndex = i;
      }
    }

    if (this.hoverIndex !== bestIndex) {
      this.hoverIndex = bestIndex;
      this.updateLayout(this.bounds);
    }
  }

  private async loadTextureFromImageUrl(url: string): Promise<Texture> {
    const blob = await this.fetchImageBlob(url);
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

  private async fetchImageBlob(url: string): Promise<Blob> {
    const useCredentialsFirst = this.shouldUseCredentialsFirst(url);

    if (!useCredentialsFirst) {
      const response = await fetch(url);
      if (response.ok) {
        return response.blob();
      }

      throw new Error(`Failed to fetch image: ${url} (${response.status})`);
    }

    let credentialsStatus = "request_failed";

    try {
      const withCredentials = await fetch(url, {
        credentials: "include",
      });

      credentialsStatus = String(withCredentials.status);
      if (withCredentials.ok) {
        return withCredentials.blob();
      }
    } catch {
      credentialsStatus = "request_failed";
    }

    const withoutCredentials = await fetch(url);
    if (withoutCredentials.ok) {
      return withoutCredentials.blob();
    }

    throw new Error(
      `Failed to fetch image: ${url} (${credentialsStatus}/${withoutCredentials.status})`,
    );
  }

  private shouldUseCredentialsFirst(url: string): boolean {
    try {
      const parsed = new URL(url, window.location.href);

      if (parsed.hostname.endsWith("googleusercontent.com")) {
        return false;
      }

      return true;
    } catch {
      return true;
    }
  }
}
