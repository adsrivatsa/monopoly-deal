import {
  DemandSource,
  type Demand,
  type Player,
} from "../../../../generated/monopoly_deal";

type PropertyDemandOverlayProps = {
  demand: Demand;
  players: Player[];
  selfPlayerId?: string;
  targetCardImageUrl?: string;
  sourceCardImageUrl?: string;
  canDeny: boolean;
  demandTimerSeconds?: number;
  isChoosingForcedDealPlacement: boolean;
  selectedForcedDealPlacementSetId?: string;
  onStartForcedDealPlacementSelection: (demandId: string) => void;
  onConfirmForcedDealPlacementSelection: (demandId: string) => void;
  onCancelForcedDealPlacementSelection: () => void;
  onComply: (demandId: string, propertySetId?: string) => void;
  onDeny: (demandId: string) => void;
};

const PropertyDemandOverlay = ({
  demand,
  players,
  selfPlayerId,
  targetCardImageUrl,
  sourceCardImageUrl,
  canDeny,
  demandTimerSeconds,
  isChoosingForcedDealPlacement,
  selectedForcedDealPlacementSetId,
  onStartForcedDealPlacementSelection,
  onConfirmForcedDealPlacementSelection,
  onCancelForcedDealPlacementSelection,
  onComply,
  onDeny,
}: PropertyDemandOverlayProps) => {
  const sourcePlayer = players.find((player) => player.playerId === demand.sourceId);
  const targetPlayer = players.find((player) => player.playerId === demand.playerId);
  const isTargetingSelf = !!selfPlayerId && demand.playerId === selfPlayerId;
  const isSelfSource = !!selfPlayerId && demand.sourceId === selfPlayerId;
  const sourceName = sourcePlayer?.displayName ?? demand.sourceId;
  const targetName = targetPlayer?.displayName ?? demand.playerId;
  const isSlyDealDemand =
    demand.demandSource === DemandSource.DEMAND_SOURCE_SLY_DEAL;
  const isForcedDealDemand =
    demand.demandSource === DemandSource.DEMAND_SOURCE_FORCED_DEAL;
  const hasReturnProperty = !!demand.propertyDemand?.sourceCardId;
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
  const selfInlineName = (
    <span className="md-demand-inline-player">
      <img
        className="md-demand-source__avatar md-demand-inline-player__avatar"
        src={sourcePlayer?.avatarUrl ?? ""}
        alt="You"
        loading="lazy"
        referrerPolicy="no-referrer"
      />
      <span className="md-demand-source__name">You</span>
    </span>
  );
  const demandEyebrow = isSlyDealDemand
    ? "Property Steal"
    : isForcedDealDemand
      ? "Property Swap"
      : "Property Demand";
  const demandLine = isTargetingSelf
    ? isSlyDealDemand
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
        : "said no to your property ask:"
    : isSlyDealDemand
      ? demand.isActive
        ? "is answering a property steal from"
        : "blocked a property steal from"
      : isForcedDealDemand
        ? demand.isActive
          ? "is answering a property swap from"
          : "blocked a property swap from"
        : demand.isActive
          ? "is answering a property demand from"
          : "blocked a property demand from";
  const demandActor = isTargetingSelf ? sourcePlayer : targetPlayer;
  const demandActorName = sourceName;
  const firstExchangeImageUrl = demand.isActive
    ? sourceCardImageUrl
    : targetCardImageUrl;
  const secondExchangeImageUrl = demand.isActive
    ? targetCardImageUrl
    : sourceCardImageUrl;
  const isDeniedBySelf = !isTargetingSelf && !demand.isActive && isSelfSource;

  const shouldAskForcedDealPlacement =
    isTargetingSelf &&
    isForcedDealDemand &&
    demand.isActive &&
    !!hasReturnProperty;

  const onClickComply = () => {
    if (!shouldAskForcedDealPlacement) {
      onComply(demand.id);
      return;
    }

    onStartForcedDealPlacementSelection(demand.id);
  };

  const onConfirmForcedDealPlacement = () => {
    onConfirmForcedDealPlacementSelection(demand.id);
  };

  return (
    <aside
      className="md-demand md-demand--payment"
      aria-live="polite"
      onPointerDown={(event) => event.stopPropagation()}
    >
      <p className="md-demand__eyebrow">{demandEyebrow}</p>
      {typeof demandTimerSeconds === "number" ? (
        <p className="md-demand__line md-demand__timer">
          Time left: {String(Math.floor(demandTimerSeconds / 60)).padStart(2, "0")}:{String(demandTimerSeconds % 60).padStart(2, "0")}
        </p>
      ) : null}
      {isTargetingSelf ? (
        <div className="md-demand-source">
          <img
            className="md-demand-source__avatar"
            src={demandActor?.avatarUrl ?? ""}
            alt={demandActorName}
            loading="lazy"
            referrerPolicy="no-referrer"
          />
          <p className="md-demand__line md-demand-source__message">
            <span className="md-demand-source__name">{demandActorName}</span> {demandLine}
          </p>
        </div>
      ) : (
        <p className="md-demand__line md-demand-source__message">
          {isDeniedBySelf
            ? isSlyDealDemand
              ? <>{selfInlineName} blocked a property steal from {targetInlineName}.</>
              : isForcedDealDemand
                ? <>{selfInlineName} blocked a property swap from {targetInlineName}.</>
                : <>{selfInlineName} blocked a property demand from {targetInlineName}.</>
            : <>{targetInlineName} {demandLine} {sourceReference}.</>}
        </p>
      )}
      {isTargetingSelf && hasReturnProperty ? (
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
      ) : isTargetingSelf && targetCardImageUrl ? (
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
      {isTargetingSelf && isChoosingForcedDealPlacement ? (
        <p className="md-demand__line">
          {selectedForcedDealPlacementSetId
            ? "Set selected. Click OK to accept swap."
            : "No set selected. Click OK to create a new set."}
        </p>
      ) : null}
      {isTargetingSelf ? (
        <div className="md-demand__actions">
          {isChoosingForcedDealPlacement ? (
            <>
              <button
                type="button"
                className="md-demand__button"
                onPointerDown={(event) => event.stopPropagation()}
                onClick={onConfirmForcedDealPlacement}
              >
                OK
              </button>
            </>
          ) : (
            <button
              type="button"
              className="md-demand__button"
              onPointerDown={(event) => event.stopPropagation()}
              onClick={onClickComply}
            >
              Accept
            </button>
          )}
          <button
            type="button"
            className="md-demand__button md-demand__button--secondary"
            onPointerDown={(event) => event.stopPropagation()}
            onClick={() => {
              onCancelForcedDealPlacementSelection();
              onDeny(demand.id);
            }}
            disabled={!canDeny}
            title={canDeny ? "Play Just Say No" : "Requires a Just Say No card"}
          >
            Just Say No
          </button>
        </div>
      ) : null}
    </aside>
  );
};

export default PropertyDemandOverlay;
