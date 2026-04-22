import { useEffect, useMemo, useRef, type CSSProperties } from "react";
import { type Card, type Color } from "../../../generated/monopoly_deal";
import GameCardView from "./GameCardView";

export type CardStackLayout = "stack" | "spread";

type GameCardStackBoxProps = {
  title: string;
  cards: Card[];
  assetImageByKey: Record<number, string>;
  layout?: CardStackLayout;
  color?: Color;
  className?: string;
  emptyLabel?: string;
  selectableCards?: boolean;
  selectedCardIds?: ReadonlySet<string>;
  onCardClick?: (card: Card) => void;
};

const colorToCss = (color: Color | undefined): string | undefined => {
  switch (color) {
    case 1:
      return "#5a2f2e";
    case 2:
      return "#bdd8ff";
    case 3:
      return "#ce3f8d";
    case 4:
      return "#ff9f47";
    case 5:
      return "#ed453d";
    case 6:
      return "#f4e954";
    case 7:
      return "#2c7241";
    case 8:
      return "#2b56ba";
    case 9:
      return "#bdc9a4";
    case 10:
      return "#222";
    default:
      return undefined;
  }
};

const GameCardStackBox = ({
  title,
  cards,
  assetImageByKey,
  layout = "stack",
  color,
  className,
  emptyLabel = "No cards",
  selectableCards = false,
  selectedCardIds,
  onCardClick,
}: GameCardStackBoxProps) => {
  const cardsContainerRef = useRef<HTMLDivElement | null>(null);

  useEffect(() => {
    const cardsContainer = cardsContainerRef.current;
    if (!cardsContainer) {
      return;
    }

    const onWheel = (event: WheelEvent) => {
      const hasHorizontalOverflow =
        cardsContainer.scrollWidth > cardsContainer.clientWidth + 1;
      const isVerticalWheel = Math.abs(event.deltaY) >= Math.abs(event.deltaX);

      if (!hasHorizontalOverflow || !isVerticalWheel) {
        return;
      }

      cardsContainer.scrollLeft += event.deltaY;
      event.preventDefault();
    };

    cardsContainer.addEventListener("wheel", onWheel, { passive: false });

    return () => {
      cardsContainer.removeEventListener("wheel", onWheel);
    };
  }, []);

  const colorStyle = useMemo<CSSProperties | undefined>(() => {
    if (!color) {
      return undefined;
    }

    return {
      ["--md-stack-color" as string]: colorToCss(color),
    };
  }, [color]);

  const rootClass = useMemo(() => {
    const classes = ["md-stack-box", `md-stack-box--${layout}`];
    if (className) {
      classes.push(className);
    }

    return classes.join(" ");
  }, [className, layout]);

  return (
    <section
      className={rootClass}
      style={colorStyle}
    >
      <p className="md-stack-box__title">{title}</p>
      <div
        className="md-stack-box__cards"
        role="list"
        aria-label={title}
        ref={cardsContainerRef}
      >
        {cards.length === 0 ? (
          <p className="md-stack-box__empty">{emptyLabel}</p>
        ) : (
          cards.map((card) => (
            <div className="md-stack-box__card-wrap" key={card.cardId} role="listitem">
              <GameCardView
                card={card}
                imageUrl={assetImageByKey[card.assetKey]}
                isSelected={selectedCardIds?.has(card.cardId) ?? false}
                className={layout === "stack" ? "md-card--tight" : ""}
                onClick={selectableCards ? onCardClick : undefined}
              />
            </div>
          ))
        )}
      </div>
    </section>
  );
};

export default GameCardStackBox;
