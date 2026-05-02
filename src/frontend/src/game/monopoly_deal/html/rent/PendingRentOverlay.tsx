import { type PendingRent, type Player } from "../../../../generated/monopoly_deal";

type PendingRentOverlayProps = {
  pendingRent: PendingRent;
  players: Player[];
  selfPlayerId?: string;
  timerSeconds?: number;
  canDouble: boolean;
  onDouble: () => void;
  onRent: () => void;
};

const PendingRentOverlay = ({
  pendingRent,
  players,
  selfPlayerId,
  timerSeconds,
  canDouble,
  onDouble,
  onRent,
}: PendingRentOverlayProps) => {
  const renter = players.find((player) => player.playerId === pendingRent.playerId);
  const renterName = renter?.displayName ?? pendingRent.playerId;
  const selfPlayer = selfPlayerId
    ? players.find((player) => player.playerId === selfPlayerId)
    : undefined;
  const isSelfRenter = !!selfPlayerId && pendingRent.playerId === selfPlayerId;
  const isSelfTarget = !!selfPlayerId && pendingRent.targetIds.includes(selfPlayerId);
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
  const primaryTarget = targets[0];

  if (!isSelfRenter) {
    return (
      <aside
        className="md-demand md-demand--payment"
        aria-live="polite"
        onPointerDown={(event) => event.stopPropagation()}
      >
        <p className="md-demand__eyebrow">Rent</p>
        {typeof timerSeconds === "number" ? (
          <p className="md-demand__line">Time left: {String(Math.floor(Math.max(0, timerSeconds) / 60)).padStart(2, "0")}:{String(Math.max(0, timerSeconds) % 60).padStart(2, "0")}</p>
        ) : null}
        {isSelfTarget ? (
          <p className="md-demand__line md-demand__line--rent-warning">
            <img
              className="md-rent-target__avatar"
              src={renter?.avatarUrl ?? ""}
              alt={renterName}
              loading="lazy"
              referrerPolicy="no-referrer"
            />{" "}
            <strong>{renterName}</strong> is about to rent{" "}
            <img
              className="md-rent-target__avatar"
              src={selfPlayer?.avatarUrl ?? ""}
              alt={selfPlayer?.displayName ?? "You"}
              loading="lazy"
              referrerPolicy="no-referrer"
            />{" "}
            <strong>you</strong>!
          </p>
        ) : primaryTarget ? (
          <p className="md-demand__line md-demand__line--rent-warning">
            <img
              className="md-rent-target__avatar"
              src={renter?.avatarUrl ?? ""}
              alt={renterName}
              loading="lazy"
              referrerPolicy="no-referrer"
            />{" "}
            <strong>{renterName}</strong> is about to rent{" "}
            <img
              className="md-rent-target__avatar"
              src={primaryTarget.avatarUrl}
              alt={primaryTarget.displayName}
              loading="lazy"
              referrerPolicy="no-referrer"
            />{" "}
            <strong>{primaryTarget.displayName}</strong>!
          </p>
        ) : (
          <p className="md-demand__line">
            <strong>{renterName}</strong> is about to rent!
          </p>
        )}
      </aside>
    );
  }

  return (
    <aside
      className="md-demand md-demand--payment"
      aria-live="polite"
      onPointerDown={(event) => event.stopPropagation()}
    >
      <p className="md-demand__eyebrow">Rent</p>
      {typeof timerSeconds === "number" ? (
        <p className="md-demand__line">Time left: {String(Math.floor(Math.max(0, timerSeconds) / 60)).padStart(2, "0")}:{String(Math.max(0, timerSeconds) % 60).padStart(2, "0")}</p>
      ) : null}
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
