import {
  DemandSource,
  type Demand,
  type Player,
} from "../../../../generated/monopoly_deal";

type PropertyDemandOverlayProps = {
  demand: Demand;
  players: Player[];
  targetCardImageUrl?: string;
  sourceCardImageUrl?: string;
  canDeny: boolean;
  onComply: (demandId: string) => void;
  onDeny: (demandId: string) => void;
};

const PropertyDemandOverlay = ({
  demand,
  players,
  targetCardImageUrl,
  sourceCardImageUrl,
  canDeny,
  onComply,
  onDeny,
}: PropertyDemandOverlayProps) => {
  const sourcePlayer = players.find((player) => player.playerId === demand.sourceId);
  const sourceName = sourcePlayer?.displayName ?? demand.sourceId;
  const isSlyDealDemand =
    demand.demandSource === DemandSource.DEMAND_SOURCE_SLY_DEAL;
  const isForcedDealDemand =
    demand.demandSource === DemandSource.DEMAND_SOURCE_FORCED_DEAL;
  const hasReturnProperty = !!demand.propertyDemand?.sourceCardId;
  const demandLine = isSlyDealDemand
    ? demand.isActive
      ? "is trying to steal your property:"
      : "blocked your steal of their property:"
    : isForcedDealDemand
      ? demand.isActive
        ? "wants to swap properties:"
        : "doesn't want to swap properties:"
    : demand.isActive
      ? hasReturnProperty
        ? "wants your property, they're giving:"
        : "wants your property:"
      : "said no to your property ask:";
  const firstExchangeImageUrl = demand.isActive
    ? sourceCardImageUrl
    : targetCardImageUrl;
  const secondExchangeImageUrl = demand.isActive
    ? targetCardImageUrl
    : sourceCardImageUrl;

  return (
    <aside
      className="md-demand md-demand--payment"
      aria-live="polite"
      onPointerDown={(event) => event.stopPropagation()}
    >
      <h3 className="md-demand__title">Property demand</h3>
      <div className="md-demand-source">
        <img
          className="md-demand-source__avatar"
          src={sourcePlayer?.avatarUrl ?? ""}
          alt={sourcePlayer?.displayName ?? demand.sourceId}
          loading="lazy"
          referrerPolicy="no-referrer"
        />
        <p className="md-demand__line md-demand-source__message">
          <span className="md-demand-source__name">{sourceName}</span> {demandLine}
        </p>
      </div>
      {hasReturnProperty ? (
        <div className="md-demand-property-card">
          <div className="md-stack-box__cards" role="list" aria-label="Property exchange">
            {firstExchangeImageUrl ? (
              <div className="md-stack-box__card-wrap" role="listitem">
                <img
                  className="md-demand-property-card__image"
                  src={firstExchangeImageUrl}
                  alt="Requested property"
                  loading="lazy"
                  referrerPolicy="no-referrer"
                />
              </div>
            ) : null}
            {firstExchangeImageUrl && secondExchangeImageUrl ? (
              <div className="md-demand-property-card__swap-arrows" aria-hidden="true">
                <span className="md-demand-property-card__swap-arrow md-demand-property-card__swap-arrow--down">
                  ⇩
                </span>
                <span className="md-demand-property-card__swap-arrow md-demand-property-card__swap-arrow--up">
                  ⇧
                </span>
              </div>
            ) : null}
            {secondExchangeImageUrl ? (
              <div className="md-stack-box__card-wrap" role="listitem">
                <img
                  className="md-demand-property-card__image"
                  src={secondExchangeImageUrl}
                  alt="Offered property"
                  loading="lazy"
                  referrerPolicy="no-referrer"
                />
              </div>
            ) : null}
          </div>
        </div>
      ) : targetCardImageUrl ? (
        <div className="md-demand-property-card">
          <img
            className="md-demand-property-card__image"
            src={targetCardImageUrl}
            alt="Requested property"
            loading="lazy"
            referrerPolicy="no-referrer"
          />
        </div>
      ) : null}
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
          Just Say No
        </button>
      </div>
    </aside>
  );
};

export default PropertyDemandOverlay;
