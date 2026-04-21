import { type Demand, type Player } from "../../../../generated/monopoly_deal";

type PropertyDemandOverlayProps = {
  demand: Demand;
  players: Player[];
  targetCardImageUrl?: string;
  canDeny: boolean;
  onComply: (demandId: string) => void;
  onDeny: (demandId: string) => void;
};

const PropertyDemandOverlay = ({
  demand,
  players,
  targetCardImageUrl,
  canDeny,
  onComply,
  onDeny,
}: PropertyDemandOverlayProps) => {
  const sourcePlayer = players.find((player) => player.playerId === demand.sourceId);

  return (
    <aside
      className="md-demand md-demand--payment"
      aria-live="polite"
      onPointerDown={(event) => event.stopPropagation()}
    >
      <p className="md-demand__eyebrow">Active demand</p>
      <h3 className="md-demand__title">Property demand</h3>
      <div className="md-demand-source">
        <img
          className="md-demand-source__avatar"
          src={sourcePlayer?.avatarUrl ?? ""}
          alt={sourcePlayer?.displayName ?? demand.sourceId}
          loading="lazy"
          referrerPolicy="no-referrer"
        />
        <p className="md-demand__line md-demand-source__name">
          From: {sourcePlayer?.displayName ?? demand.sourceId}
        </p>
      </div>
      {targetCardImageUrl ? (
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
          Comply
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
