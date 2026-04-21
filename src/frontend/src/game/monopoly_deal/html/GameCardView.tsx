import { useEffect, useMemo, useState, type CSSProperties } from "react";
import {
  assetKeyToJSON,
  Color,
  type Card,
} from "../../../generated/monopoly_deal";

type GameCardViewProps = {
  card: Card;
  imageUrl?: string;
  draggable?: boolean;
  isSelected?: boolean;
  className?: string;
  onClick?: (card: Card) => void;
  onDragStart?: (card: Card) => void;
  onDragEnd?: () => void;
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

const GameCardView = ({
  card,
  imageUrl,
  draggable = false,
  isSelected = false,
  className,
  onClick,
  onDragStart,
  onDragEnd,
}: GameCardViewProps) => {
  const [isImageBroken, setIsImageBroken] = useState(false);
  const [cardAspectRatio, setCardAspectRatio] = useState("10 / 16");
  const hasImage = !!imageUrl && !isImageBroken;

  useEffect(() => {
    setIsImageBroken(false);
    setCardAspectRatio("10 / 16");
  }, [imageUrl]);

  const computedClassName = useMemo(() => {
    const classes = ["md-card"];
    if (className) {
      classes.push(className);
    }
    if (onClick) {
      classes.push("md-card--clickable");
    }
    if (isSelected) {
      classes.push("md-card--selected");
    }
    if (hasImage) {
      classes.push("md-card--has-image");
    }
    if (shouldFlipTwoColorWild(card)) {
      classes.push("md-card--flipped");
    }

    return classes.join(" ");
  }, [card, className, hasImage, isSelected, onClick]);

  const cardStyle = useMemo<CSSProperties>(() => {
    return {
      ["--md-card-aspect-ratio" as string]: cardAspectRatio,
    };
  }, [cardAspectRatio]);

  return (
    <article
      className={computedClassName}
      style={cardStyle}
      draggable={draggable}
      onPointerDown={(event) => {
        if (onClick) {
          event.stopPropagation();
        }
      }}
      onClick={(event) => {
        if (!onClick) {
          return;
        }
        event.stopPropagation();
        onClick(card);
      }}
      onDragStart={(event) => {
        event.dataTransfer.setData("text/plain", card.cardId);
        event.dataTransfer.effectAllowed = "move";
        onDragStart?.(card);
      }}
      onDragEnd={() => onDragEnd?.()}
      aria-label={assetKeyToJSON(card.assetKey)}
    >
      {hasImage ? (
        <img
          src={imageUrl}
          alt={assetKeyToJSON(card.assetKey)}
          className="md-card__image"
          draggable={false}
          loading="lazy"
          onLoad={(event) => {
            const { naturalWidth, naturalHeight } = event.currentTarget;
            if (naturalWidth > 0 && naturalHeight > 0) {
              setCardAspectRatio(`${naturalWidth} / ${naturalHeight}`);
            }
          }}
          onError={() => setIsImageBroken(true)}
          referrerPolicy="no-referrer"
        />
      ) : (
        <span className="md-card__fallback">{assetKeyToJSON(card.assetKey)}</span>
      )}
    </article>
  );
};

export default GameCardView;
