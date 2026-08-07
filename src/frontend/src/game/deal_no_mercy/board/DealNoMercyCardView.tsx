// A single No Mercy card face. Mirrors the classic GameCardView but binds to the
// No Mercy generated types (its own asset-key enum). Renders the large asset
// image with a fallback to the asset-key label, and self-corrects its aspect
// ratio from the loaded image so mixed art sizes don't stretch.
import { useEffect, useMemo, useState, type CSSProperties } from "react";
import { assetKeyToJSON, type Card } from "../../../generated/deal_no_mercy";

type DealNoMercyCardViewProps = {
  card: Card;
  imageUrl?: string;
  isSelected?: boolean;
  className?: string;
  onClick?: (card: Card) => void;
};

const DealNoMercyCardView = ({
  card,
  imageUrl,
  isSelected = false,
  className,
  onClick,
}: DealNoMercyCardViewProps) => {
  const [isImageBroken, setIsImageBroken] = useState(false);
  const [cardAspectRatio, setCardAspectRatio] = useState("10 / 16");
  const hasImage = !!imageUrl && !isImageBroken;

  useEffect(() => {
    setIsImageBroken(false);
    setCardAspectRatio("10 / 16");
  }, [imageUrl]);

  const computedClassName = useMemo(() => {
    const classes = ["dnm-card"];
    if (className) {
      classes.push(className);
    }
    if (onClick) {
      classes.push("dnm-card--clickable");
    }
    if (isSelected) {
      classes.push("dnm-card--selected");
    }
    if (hasImage) {
      classes.push("dnm-card--has-image");
    }
    return classes.join(" ");
  }, [className, hasImage, isSelected, onClick]);

  const cardStyle = useMemo<CSSProperties>(
    () => ({ ["--dnm-card-aspect-ratio" as string]: cardAspectRatio }),
    [cardAspectRatio],
  );

  return (
    <article
      className={computedClassName}
      style={cardStyle}
      onClick={(event) => {
        if (!onClick) {
          return;
        }
        event.stopPropagation();
        onClick(card);
      }}
      aria-label={assetKeyToJSON(card.assetKey)}
    >
      {hasImage ? (
        <img
          src={imageUrl}
          alt={assetKeyToJSON(card.assetKey)}
          className="dnm-card__image"
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
        <span className="dnm-card__fallback">{assetKeyToJSON(card.assetKey)}</span>
      )}
    </article>
  );
};

export default DealNoMercyCardView;
