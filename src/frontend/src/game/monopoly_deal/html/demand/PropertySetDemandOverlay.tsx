import type { CSSProperties } from "react";
import {
  assetKeyToJSON,
  type Card,
  type Demand,
  type Player,
} from "../../../../generated/monopoly_deal";

type PropertySetDemandOverlayProps = {
  demand: Demand;
  players: Player[];
  targetCardImageById: Record<string, string>;
  requestedCards: Card[];
  canDeny: boolean;
  onComply: (demandId: string) => void;
  onDeny: (demandId: string) => void;
};

const PropertySetDemandOverlay = ({
  demand,
  players,
  targetCardImageById,
  requestedCards,
  canDeny,
  onComply,
  onDeny,
}: PropertySetDemandOverlayProps) => {
  const sourcePlayer = players.find((player) => player.playerId === demand.sourceId);
  const targetPlayer = players.find((player) => player.playerId === demand.playerId);
  const imageStripWidthPx = Math.max(requestedCards.length * 62, 170);
  const demandContentWidthPx = Math.min(imageStripWidthPx, 360);
  const demandStyle = {
    ["--md-demand-content-width" as string]: `${demandContentWidthPx}px`,
  } as CSSProperties;

  return (
    <aside
      className="md-demand md-demand--payment md-demand--property-set"
      style={demandStyle}
      aria-live="polite"
      onPointerDown={(event) => event.stopPropagation()}
    >
      <h3 className="md-demand__title">Deal Breaker</h3>
      <div className="md-demand-source">
        <img
          className="md-demand-source__avatar"
          src={
            demand.isActive
              ? sourcePlayer?.avatarUrl ?? ""
              : targetPlayer?.avatarUrl ?? ""
          }
          alt={
            demand.isActive
              ? sourcePlayer?.displayName ?? demand.sourceId
              : targetPlayer?.displayName ?? demand.playerId
          }
          loading="lazy"
          referrerPolicy="no-referrer"
        />
        <p className="md-demand__line md-demand-source__name">
          {demand.isActive
            ? sourcePlayer?.displayName ?? demand.sourceId
            : targetPlayer?.displayName ?? demand.playerId}
        </p>
      </div>
      <p className="md-demand__line">
        {demand.isActive
          ? "wants your property set:"
          : "said no to your deal breaker for set:"}
      </p>
      {requestedCards.length > 0 ? (
        <div className="md-demand-property-card">
          <div className="md-stack-box__cards" role="list" aria-label="Requested properties">
            {requestedCards.map((card) => {
              const imageUrl = targetCardImageById[card.cardId];
              return (
                <div className="md-stack-box__card-wrap" key={card.cardId} role="listitem">
                  {imageUrl ? (
                    <img
                      className="md-demand-property-card__image"
                      src={imageUrl}
                      alt={assetKeyToJSON(card.assetKey)}
                      loading="lazy"
                      referrerPolicy="no-referrer"
                    />
                  ) : (
                    <p className="md-demand__line">{assetKeyToJSON(card.assetKey)}</p>
                  )}
                </div>
              );
            })}
          </div>
        </div>
      ) : (
        <p className="md-demand__line">No requested cards found.</p>
      )}
      <div className="md-demand__actions">
        <button
          type="button"
          className="md-demand__button"
          onPointerDown={(event) => event.stopPropagation()}
          onClick={() => onComply(demand.id)}
        >
          OK
        </button>
        <button
          type="button"
          className="md-demand__button md-demand__button--secondary"
          onPointerDown={(event) => event.stopPropagation()}
          onClick={() => onDeny(demand.id)}
          disabled={!canDeny}
          title={canDeny ? "Play Just Say No" : "Requires a Just Say No card"}
        >
          Just Say No!
        </button>
      </div>
    </aside>
  );
};

export default PropertySetDemandOverlay;
