import { type PendingRent, type Player } from "../../../../generated/monopoly_deal";

type PendingRentOverlayProps = {
  pendingRent: PendingRent;
  players: Player[];
  canDouble: boolean;
  onDouble: () => void;
  onRent: () => void;
};

const PendingRentOverlay = ({
  pendingRent,
  players,
  canDouble,
  onDouble,
  onRent,
}: PendingRentOverlayProps) => {
  const targets = pendingRent.targetIds.map((targetId) => {
    const player = players.find((candidate) => candidate.playerId === targetId);
    return {
      playerId: targetId,
      displayName: player?.displayName ?? targetId,
      avatarUrl: player?.avatarUrl ?? "",
    };
  });
  const totalAmount = pendingRent.baseAmount * Math.max(1, pendingRent.multiplier);
  const targetLabel = targets.length === 1 ? "Target" : "Targets";

  return (
    <aside
      className="md-demand md-demand--payment"
      aria-live="polite"
      onPointerDown={(event) => event.stopPropagation()}
    >
      <h3 className="md-demand__title">Pending rent</h3>
      <p className="md-demand__line">{targetLabel}:</p>
      <div className="md-rent-targets" role="list" aria-label="Rent targets">
        {targets.length === 0 ? (
          <p className="md-demand__line">-</p>
        ) : (
          targets.map((target) => (
            <div key={target.playerId} className="md-rent-target" role="listitem">
              <img
                className="md-rent-target__avatar"
                src={target.avatarUrl}
                alt={target.displayName}
                loading="lazy"
                referrerPolicy="no-referrer"
              />
              <span className="md-rent-target__name">{target.displayName}</span>
            </div>
          ))
        )}
      </div>
      <p className="md-demand__line">Payment: ${totalAmount}</p>
      <div className="md-demand__actions">
        <button
          type="button"
          className="md-demand__button"
          onPointerDown={(event) => event.stopPropagation()}
          onClick={onDouble}
          disabled={!canDouble}
          title={canDouble ? "Play Double The Rent" : "Requires Double The Rent"}
        >
          Double!
        </button>
        <button
          type="button"
          className="md-demand__button md-demand__button--secondary"
          onPointerDown={(event) => event.stopPropagation()}
          onClick={onRent}
        >
          Rent!
        </button>
      </div>
    </aside>
  );
};

export default PendingRentOverlay;
