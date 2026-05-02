import {
  assetKeyToJSON,
  type Card,
  type Demand,
  type Player,
} from "../../../../generated/monopoly_deal";

export type PropertySetDemandOverlayProps = {
  demand: Demand;
  players: Player[];
  selfPlayerId?: string;
  targetCardImageById: Record<string, string>;
  requestedCards: Card[];
  demandTimerSeconds?: number;
  canDeny: boolean;
  onComply: (demandId: string) => void;
  onDeny: (demandId: string) => void;
};

const PropertySetDemandOverlay = ({
  demand,
  players,
  selfPlayerId,
  targetCardImageById,
  requestedCards,
  demandTimerSeconds,
  canDeny,
  onComply,
  onDeny,
}: PropertySetDemandOverlayProps) => {
  const sourcePlayer = players.find((player) => player.playerId === demand.sourceId);
  const targetPlayer = players.find((player) => player.playerId === demand.playerId);
  const sourceName = sourcePlayer?.displayName ?? demand.sourceId;
  const targetName = targetPlayer?.displayName ?? demand.playerId;
  const isTargetingSelf = !!selfPlayerId && demand.playerId === selfPlayerId;
  const isSelfSource = !!selfPlayerId && demand.sourceId === selfPlayerId;
  const sourceInlineName = (
    <span className="md-demand-inline-player">
      <img
        className="md-demand-source__avatar md-demand-inline-player__avatar"
        src={sourcePlayer?.avatarUrl ?? ""}
        alt={sourceName}
        loading="lazy"
        referrerPolicy="no-referrer"
      />
      <span className="md-demand-source__name">{sourceName}</span>
    </span>
  );
  const targetInlineName = (
    <span className="md-demand-inline-player">
      <img
        className="md-demand-source__avatar md-demand-inline-player__avatar"
        src={targetPlayer?.avatarUrl ?? ""}
        alt={targetName}
        loading="lazy"
        referrerPolicy="no-referrer"
      />
      <span className="md-demand-source__name">{targetName}</span>
    </span>
  );
  const sourceReference = isSelfSource ? "you" : sourceInlineName;
  const targetReference = isTargetingSelf ? "You" : targetInlineName;
  const sourceActorReference = isSelfSource ? "You" : sourceInlineName;

  return (
    <aside
      className="md-demand md-demand--payment"
      aria-live="polite"
      onPointerDown={(event) => event.stopPropagation()}
    >
      <p className="md-demand__eyebrow">Set Snatcher</p>
      {typeof demandTimerSeconds === "number" ? (
        <p className="md-demand__line md-demand__timer">
          ⏱ {String(Math.floor(demandTimerSeconds / 60)).padStart(2, "0")}:{String(demandTimerSeconds % 60).padStart(2, "0")}
        </p>
      ) : null}
      {isTargetingSelf ? (
        <div className="md-demand-source">
          <img
            className="md-demand-source__avatar"
            src={sourcePlayer?.avatarUrl ?? ""}
            alt={sourceName}
            loading="lazy"
            referrerPolicy="no-referrer"
          />
          <p className="md-demand__line md-demand-source__message">
            <span className="md-demand-source__name">{sourceName}</span>{" "}
            {demand.isActive
              ? "is trying to snatch your set:"
              : "blocked you from snatching their set:"}
          </p>
        </div>
      ) : (
        <p className="md-demand__line md-demand-source__message">
          {demand.isActive
            ? (
                <>
                  {targetReference} is answering a set snatch from {sourceReference}.
                </>
              )
            : (
                <>
                  {sourceActorReference} blocked a set snatcher from {targetInlineName}, they're responding.
                </>
              )}
        </p>
      )}
      {isTargetingSelf && requestedCards.length > 0 ? (
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
      ) : isTargetingSelf ? (
        <p className="md-demand__line">No requested cards found.</p>
      ) : null}
      {isTargetingSelf ? (
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
      ) : null}
    </aside>
  );
};

export default PropertySetDemandOverlay;
