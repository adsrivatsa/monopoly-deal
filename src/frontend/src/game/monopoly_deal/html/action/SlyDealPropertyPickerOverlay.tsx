import { type Card, type Player, type PropertySet } from "../../../../generated/monopoly_deal";

type SlyDealPropertyPickerOverlayProps = {
  eyebrow: string;
  title: string;
  targetPlayer?: Player;
  propertySets: PropertySet[];
  assetImageByKey: Record<number, string>;
  emptyText?: string;
  onSelectCard: (targetCardId: string) => void;
  onClose: () => void;
};

const SlyDealPropertyPickerOverlay = ({
  eyebrow,
  title,
  targetPlayer,
  propertySets,
  assetImageByKey,
  emptyText = "No selectable property cards.",
  onSelectCard,
  onClose,
}: SlyDealPropertyPickerOverlayProps) => {
  const cards: Card[] = propertySets.flatMap((propertySet) => propertySet.cards);

  return (
    <div className="md-player-picker-backdrop" role="presentation" onClick={onClose}>
      <aside
        className="md-player-picker"
        role="dialog"
        aria-modal="true"
        aria-label="Choose sly deal target card"
        onClick={(event) => event.stopPropagation()}
        onPointerDown={(event) => event.stopPropagation()}
      >
        <p className="md-player-picker__eyebrow">{eyebrow}</p>
        <p className="md-player-picker__title">{title}</p>
        {targetPlayer ? (
          <p className="md-player-picker__empty">Player: {targetPlayer.displayName}</p>
        ) : null}
        {cards.length === 0 ? (
          <p className="md-player-picker__empty">{emptyText}</p>
        ) : (
          <div className="md-sly-card-list" role="list" aria-label="Stealable properties">
            {cards.map((card) => (
              <button
                key={card.cardId}
                type="button"
                className="md-sly-card-item"
                onClick={() => onSelectCard(card.cardId)}
                role="listitem"
              >
                <img
                  className="md-sly-card-item__image"
                  src={assetImageByKey[card.assetKey]}
                  alt={String(card.assetKey)}
                  loading="lazy"
                  referrerPolicy="no-referrer"
                />
              </button>
            ))}
          </div>
        )}
      </aside>
    </div>
  );
};

export default SlyDealPropertyPickerOverlay;
