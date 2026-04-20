import {
  Container,
  Graphics,
  Sprite,
  Text,
  TextStyle,
  Texture,
} from "pixi.js";
import type {
  Money,
  Player,
  PropertySet,
} from "../../../generated/monopoly_deal";
import { MoneyPileBox } from "./MoneyPileBox";
import { PropertySetBox } from "./PropertySetBox";

type PaletteLike = {
  card: { color: number; alpha: number };
  muted: { color: number; alpha: number };
  border: { color: number; alpha: number };
  foreground: { color: number; alpha: number };
  mutedForeground: { color: number; alpha: number };
  primary: { color: number; alpha: number };
};

type Bounds = {
  x: number;
  y: number;
  width: number;
  height: number;
};

type PlayerView = {
  playerId: string;
  container: Container;
  box: Graphics;
  avatarBg: Graphics;
  avatarMask: Graphics;
  avatar: Sprite;
  name: Text;
  moneyBox: MoneyPileBox;
  propertySetsLayer: Container;
  propertySetBoxes: Array<{
    propertySetId: string;
    box: PropertySetBox;
    bounds: Bounds;
  }>;
  isCurrentPlayer: boolean;
};

const DEFAULT_PALETTE: PaletteLike = {
  card: { color: 0x0b1018, alpha: 1 },
  muted: { color: 0x121a26, alpha: 1 },
  border: { color: 0x1d2c3f, alpha: 1 },
  foreground: { color: 0xf8fbff, alpha: 1 },
  mutedForeground: { color: 0x9eb3c9, alpha: 1 },
  primary: { color: 0x36b6ff, alpha: 1 },
};

export class PlayerBoardBoxesLayer {
  readonly container = new Container();

  private readonly viewportMask = new Graphics();
  private readonly viewportBorder = new Graphics();
  private readonly viewport = new Container();
  private readonly content = new Container();
  private readonly textureCache = new Map<string, Texture>();
  private readonly nameStyle = new TextStyle({
    fill: DEFAULT_PALETTE.foreground.color,
    fontFamily: "JetBrains Mono, SF Mono, Menlo, monospace",
    fontSize: 13,
    fontWeight: "700",
  });

  private palette: PaletteLike = DEFAULT_PALETTE;
  private bounds: Bounds = { x: 0, y: 0, width: 0, height: 0 };
  private views: PlayerView[] = [];
  private assetImageByKey: Record<number, string> = {};
  private buildVersion = 0;
  private zoom = 1;
  private panX = 0;
  private panY = 0;
  private contentWidth = 0;
  private contentHeight = 0;
  private currentPlayerMoneyLocalBounds: Bounds | null = null;
  private currentPlayerBoxLocalBounds: Bounds | null = null;
  private currentPlayerView: PlayerView | null = null;
  private readonly minZoom = 0.65;
  private readonly maxZoom = 2.2;

  applyTheme(palette: PaletteLike) {
    this.palette = palette;
    this.nameStyle.fill = palette.foreground.color;
    this.views.forEach((view) => {
      view.moneyBox.applyTheme(palette);
      view.propertySetBoxes.forEach((entry) => {
        entry.box.applyTheme({
          card: palette.card,
          muted: palette.muted,
          border: palette.border,
          mutedForeground: palette.mutedForeground,
        });
      });
    });
    this.updateLayout(this.bounds);
  }

  constructor() {
    this.viewport.mask = this.viewportMask;
    this.viewport.addChild(this.content);
    this.container.addChild(this.viewport, this.viewportMask, this.viewportBorder);
  }

  async setPlayers(
    players: Player[],
    currentPlayerId?: string,
    money: Money[] = [],
    properties: PropertySet[] = [],
    assetImageByKey: Record<number, string> = {},
  ) {
    this.views = [];
    this.content.removeChildren();
    this.assetImageByKey = assetImageByKey;

    const version = ++this.buildVersion;

    for (const player of players) {
      const view: PlayerView = {
        playerId: player.playerId,
        container: new Container(),
        box: new Graphics(),
        avatarBg: new Graphics(),
        avatarMask: new Graphics(),
        avatar: new Sprite(Texture.EMPTY),
        name: new Text({ text: player.displayName, style: this.nameStyle }),
        moneyBox: new MoneyPileBox(),
        propertySetsLayer: new Container(),
        propertySetBoxes: [],
        isCurrentPlayer: currentPlayerId === player.playerId,
      };

      view.name.resolution = Math.min(window.devicePixelRatio || 1, 2);
      view.avatar.mask = view.avatarMask;
      view.moneyBox.applyTheme(this.palette);

      view.container.addChild(
        view.box,
        view.avatarBg,
        view.avatar,
        view.avatarMask,
        view.name,
        view.moneyBox.container,
        view.propertySetsLayer,
      );
      this.content.addChild(view.container);
      this.views.push(view);

      const moneyCards = money.find((pile) => pile.playerId === player.playerId)?.cards ?? [];
      await view.moneyBox.setCards(moneyCards, this.assetImageByKey);
      view.propertySetBoxes = properties
        .filter((propertySet) => propertySet.playerId === player.playerId)
        .map((propertySet) => {
          const propertySetBox = new PropertySetBox();
          void propertySetBox.setPropertySet(propertySet, this.assetImageByKey);
          propertySetBox.applyTheme({
            card: this.palette.card,
            muted: this.palette.muted,
            border: this.palette.border,
            mutedForeground: this.palette.mutedForeground,
          });
          view.propertySetsLayer.addChild(propertySetBox.container);
          return {
            propertySetId: propertySet.propertySetId,
            box: propertySetBox,
            bounds: { x: 0, y: 0, width: 0, height: 0 },
          };
        });
      if (version !== this.buildVersion) {
        return;
      }

      if (!player.avatarUrl) {
        continue;
      }

      try {
        let texture = this.textureCache.get(player.avatarUrl);
        if (!texture) {
          texture = await this.loadTextureFromImageUrl(player.avatarUrl);
          this.textureCache.set(player.avatarUrl, texture);
        }

        if (version !== this.buildVersion) {
          return;
        }

        view.avatar.texture = texture;
      } catch (error) {
        console.warn("[game-ui] failed to load player avatar", {
          avatarUrl: player.avatarUrl,
          error,
        });
      }
    }

    this.updateLayout(this.bounds);
  }

  updateLayout(bounds: Bounds) {
    this.bounds = bounds;
    this.currentPlayerMoneyLocalBounds = null;
    this.currentPlayerBoxLocalBounds = null;
    this.currentPlayerView = null;

    this.viewport.x = bounds.x;
    this.viewport.y = bounds.y;

    this.viewportMask
      .clear()
      .roundRect(bounds.x, bounds.y, bounds.width, bounds.height, 14)
      .fill({ color: 0xffffff, alpha: 1 });

    this.viewportBorder
      .clear()
      .roundRect(bounds.x, bounds.y, bounds.width, bounds.height, 14);

    if (this.views.length === 0 || bounds.width <= 0 || bounds.height <= 0) {
      return;
    }

    const gap = 12;
    const { rows, columns } = this.chooseGrid(this.views.length);

    const boxWidth =
      (bounds.width - gap * Math.max(columns - 1, 0)) / Math.max(columns, 1);
    const baseBoxHeight =
      (bounds.height - gap * Math.max(rows - 1, 0)) / Math.max(rows, 1);

    const topPad = 10;
    const leftPad = 10;
    const sectionGap = 10;
    const setGapY = 6;
    const innerColumns = 3;
    const avatarSize = Math.min(36, baseBoxHeight * 0.28);
    const sectionsTop = topPad + avatarSize + 12;
    const baseSectionHeight = Math.max(baseBoxHeight - sectionsTop - leftPad, 36);
    const sectionWidth = Math.max(boxWidth - leftPad * 2, 20);
    const boxSectionHeight = Math.max(baseSectionHeight * 0.5, 20);
    const boxSectionWidth = Math.max(
      (sectionWidth - sectionGap * Math.max(innerColumns - 1, 0)) / innerColumns,
      20,
    );
    const slotX = (slotIndex: number) => {
      const slotCol = slotIndex % innerColumns;
      return leftPad + slotCol * (boxSectionWidth + sectionGap);
    };
    const slotY = (slotIndex: number) => {
      const slotRow = Math.floor(slotIndex / innerColumns);
      return sectionsTop + slotRow * (boxSectionHeight + setGapY);
    };

    const boxHeight = this.views.reduce((maxHeight, view) => {
      const totalInnerBoxes = 1 + view.propertySetBoxes.length;
      const rowsNeeded = Math.max(1, Math.ceil(totalInnerBoxes / innerColumns));
      const propertyBottom =
        sectionsTop +
        rowsNeeded * boxSectionHeight +
        Math.max(rowsNeeded - 1, 0) * setGapY;
      const minimumHeight = propertyBottom + leftPad;
      return Math.max(maxHeight, minimumHeight);
    }, baseBoxHeight);

    this.views.forEach((view, index) => {
      const col = index % columns;
      const row = Math.floor(index / columns);
      const x = col * (boxWidth + gap);
      const y = row * (boxHeight + gap);

      view.container.x = x;
      view.container.y = y;

      const radius = 12;
      const borderColor = view.isCurrentPlayer
        ? this.palette.primary.color
        : this.palette.border.color;

      view.box
        .clear()
        .roundRect(0, 0, boxWidth, boxHeight, radius)
        .fill({ color: this.palette.card.color, alpha: this.palette.card.alpha })
        .stroke({ color: borderColor, alpha: this.palette.border.alpha, width: 1.2 });

      view.avatarBg
        .clear()
        .roundRect(leftPad, topPad, avatarSize, avatarSize, 9)
        .fill({ color: this.palette.muted.color, alpha: this.palette.muted.alpha });

      view.avatarMask
        .clear()
        .roundRect(leftPad, topPad, avatarSize, avatarSize, 9)
        .fill({ color: 0xffffff, alpha: 1 });

      view.avatar.x = leftPad;
      view.avatar.y = topPad;
      view.avatar.width = avatarSize;
      view.avatar.height = avatarSize;

      view.name.x = leftPad + avatarSize + 10;
      view.name.y = topPad + 8;
      view.name.style.wordWrap = true;
      view.name.style.wordWrapWidth = Math.max(boxWidth - avatarSize - 30, 30);

      const moneySlotIndex = 0;
      const moneySectionX = slotX(moneySlotIndex);
      const moneySectionY = slotY(moneySlotIndex);
      view.moneyBox.updateLayout({
        x: moneySectionX,
        y: moneySectionY,
        width: boxSectionWidth,
        height: boxSectionHeight,
      });

      if (view.isCurrentPlayer) {
        this.currentPlayerView = view;
        this.currentPlayerBoxLocalBounds = {
          x,
          y,
          width: boxWidth,
          height: boxHeight,
        };

        this.currentPlayerMoneyLocalBounds = {
          x: x + moneySectionX,
          y: y + moneySectionY,
          width: boxSectionWidth,
          height: boxSectionHeight,
        };
      }

      view.propertySetBoxes.forEach((entry, propertyIndex) => {
        const slotIndex = propertyIndex + 1;
        entry.bounds = {
          x: slotX(slotIndex),
          y: slotY(slotIndex),
          width: boxSectionWidth,
          height: boxSectionHeight,
        };
        entry.box.updateLayout(entry.bounds);
      });
    });

    this.contentWidth = boxWidth * columns + gap * Math.max(columns - 1, 0);
    this.contentHeight = boxHeight * rows + gap * Math.max(rows - 1, 0);
    this.clampPan();
    this.applyTransform();
  }

  getCurrentPlayerMoneyDropZone(): Bounds | null {
    if (!this.currentPlayerMoneyLocalBounds) {
      return null;
    }

    return {
      x:
        this.bounds.x +
        this.panX +
        this.currentPlayerMoneyLocalBounds.x * this.zoom,
      y:
        this.bounds.y +
        this.panY +
        this.currentPlayerMoneyLocalBounds.y * this.zoom,
      width: this.currentPlayerMoneyLocalBounds.width * this.zoom,
      height: this.currentPlayerMoneyLocalBounds.height * this.zoom,
    };
  }

  getCurrentPlayerPropertyDropZone(): Bounds | null {
    if (!this.currentPlayerBoxLocalBounds) {
      return null;
    }

    return {
      x:
        this.bounds.x +
        this.panX +
        this.currentPlayerBoxLocalBounds.x * this.zoom,
      y:
        this.bounds.y +
        this.panY +
        this.currentPlayerBoxLocalBounds.y * this.zoom,
      width: this.currentPlayerBoxLocalBounds.width * this.zoom,
      height: this.currentPlayerBoxLocalBounds.height * this.zoom,
    };
  }

  resolveCurrentPlayerPropertyDrop(
    pointerX: number,
    pointerY: number,
  ): { propertySetId?: string } | null {
    if (!this.currentPlayerView || !this.currentPlayerBoxLocalBounds) {
      return null;
    }

    const localX = (pointerX - this.bounds.x - this.panX) / this.zoom;
    const localY = (pointerY - this.bounds.y - this.panY) / this.zoom;

    const insideCurrentPlayerBox =
      localX >= this.currentPlayerBoxLocalBounds.x &&
      localX <= this.currentPlayerBoxLocalBounds.x + this.currentPlayerBoxLocalBounds.width &&
      localY >= this.currentPlayerBoxLocalBounds.y &&
      localY <= this.currentPlayerBoxLocalBounds.y + this.currentPlayerBoxLocalBounds.height;

    if (!insideCurrentPlayerBox) {
      return null;
    }

    const viewLocalX = localX - this.currentPlayerView.container.x;
    const viewLocalY = localY - this.currentPlayerView.container.y;
    const propertySetHit = this.currentPlayerView.propertySetBoxes.find((entry) => {
      return (
        viewLocalX >= entry.bounds.x &&
        viewLocalX <= entry.bounds.x + entry.bounds.width &&
        viewLocalY >= entry.bounds.y &&
        viewLocalY <= entry.bounds.y + entry.bounds.height
      );
    });

    if (propertySetHit) {
      return { propertySetId: propertySetHit.propertySetId };
    }

    const moneyBounds = this.currentPlayerMoneyLocalBounds;
    const isOnMoney =
      !!moneyBounds &&
      localX >= moneyBounds.x &&
      localX <= moneyBounds.x + moneyBounds.width &&
      localY >= moneyBounds.y &&
      localY <= moneyBounds.y + moneyBounds.height;

    if (isOnMoney) {
      return null;
    }

    return {};
  }

  handleWheel(pointerX: number, pointerY: number, deltaY: number): boolean {
    if (!this.isInsideBounds(pointerX, pointerY)) {
      return false;
    }

    const oldZoom = this.zoom;
    const factor = deltaY < 0 ? 1.015 : 0.985;
    this.zoom = Math.min(this.maxZoom, Math.max(this.minZoom, this.zoom * factor));

    if (oldZoom === this.zoom) {
      return true;
    }

    const localX = pointerX - this.bounds.x;
    const localY = pointerY - this.bounds.y;
    const worldX = (localX - this.panX) / oldZoom;
    const worldY = (localY - this.panY) / oldZoom;

    this.panX = localX - worldX * this.zoom;
    this.panY = localY - worldY * this.zoom;

    this.clampPan();
    this.applyTransform();
    return true;
  }

  panBy(deltaX: number, deltaY: number) {
    this.panX += deltaX;
    this.panY += deltaY;
    this.clampPan();
    this.applyTransform();
  }

  isInsideBounds(x: number, y: number): boolean {
    return (
      x >= this.bounds.x &&
      x <= this.bounds.x + this.bounds.width &&
      y >= this.bounds.y &&
      y <= this.bounds.y + this.bounds.height
    );
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

  private chooseGrid(count: number): { rows: number; columns: number } {
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
  }

  private clampPan() {
    const scaledWidth = this.contentWidth * this.zoom;
    const scaledHeight = this.contentHeight * this.zoom;

    if (scaledWidth <= this.bounds.width) {
      this.panX = (this.bounds.width - scaledWidth) / 2;
    } else {
      const minX = this.bounds.width - scaledWidth;
      this.panX = Math.min(0, Math.max(minX, this.panX));
    }

    if (scaledHeight <= this.bounds.height) {
      this.panY = (this.bounds.height - scaledHeight) / 2;
    } else {
      const minY = this.bounds.height - scaledHeight;
      this.panY = Math.min(0, Math.max(minY, this.panY));
    }
  }

  private applyTransform() {
    this.content.x = this.panX;
    this.content.y = this.panY;
    this.content.scale.set(this.zoom);
  }
}
