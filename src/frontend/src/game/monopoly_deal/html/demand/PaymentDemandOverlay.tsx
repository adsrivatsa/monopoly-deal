import {
  DemandSource,
  type Demand,
  type Player,
} from "../../../../generated/monopoly_deal";

const isItsMyBirthdayDemand = (demand: Demand): boolean => {
  return demand.demandSource === DemandSource.DEMAND_SOURCE_ITS_MY_BIRTHDAY;
};

const isDebtCollectorDemand = (demand: Demand): boolean => {
  return demand.demandSource === DemandSource.DEMAND_SOURCE_DEBT_COLLECTOR;
};

const isRentDemand = (demand: Demand): boolean => {
  return demand.demandSource === DemandSource.DEMAND_SOURCE_RENT;
};

export type PaymentDemandOverlayProps = {
  demand: Demand;
  players: Player[];
  selfPlayerId?: string;
  canDeny: boolean;
  isDemandActive: boolean;
  isSelectingCards: boolean;
  canConfirmSelection: boolean;
  selectedPaymentTotal?: number;
  onComply: (demandId: string) => void;
  onDeny: (demandId: string) => void;
};

const PaymentDemandOverlay = ({
  demand,
  players,
  selfPlayerId,
  canDeny,
  isDemandActive,
  isSelectingCards,
  canConfirmSelection,
  selectedPaymentTotal,
  onComply,
  onDeny,
}: PaymentDemandOverlayProps) => {
  const sourcePlayer = players.find((player) => player.playerId === demand.sourceId);
  const targetPlayer = players.find((player) => player.playerId === demand.playerId);
  const amount = demand.paymentDemand?.amount;
  const isBirthdayDemand = isItsMyBirthdayDemand(demand);
  const isDebtCollector = isDebtCollectorDemand(demand);
  const isRent = isRentDemand(demand);
  const sourceName = sourcePlayer?.displayName ?? demand.sourceId;
  const targetName = targetPlayer?.displayName ?? demand.playerId;
  const amountLabel = `$${typeof amount === "number" ? amount : "-"}M`;
  const isTargetingSelf = !!selfPlayerId && demand.playerId === selfPlayerId;
  const isSelfSource = !!selfPlayerId && demand.sourceId === selfPlayerId;
  const demandActor = isTargetingSelf ? sourcePlayer : targetPlayer;
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
  const sourceInlineActor = isSelfSource ? (
    <span className="md-demand-inline-player">
      <img
        className="md-demand-source__avatar md-demand-inline-player__avatar"
        src={sourcePlayer?.avatarUrl ?? ""}
        alt={sourceName}
        loading="lazy"
        referrerPolicy="no-referrer"
      />
      <span className="md-demand-source__name">You</span>
    </span>
  ) : sourceInlineName;
  const demandEyebrow = isBirthdayDemand
    ? "Party Bill"
    : isDebtCollector
      ? "Settle Up"
      : isRent
        ? "Rent"
        : "Payment Demand";
  const demandLine = isTargetingSelf
    ? isBirthdayDemand
      ? isDemandActive
        ? `is throwing a party. Pay ${amountLabel} to attend.`
        : `doesn't want to come to your party :(.`
      : isDebtCollector
        ? isDemandActive
          ? `wants you to settle up. Pay them ${amountLabel}.`
          : `wants to settle up ${amountLabel} later (probably never).`
        : isRent
          ? isDemandActive
            ? `wants you to pay your rent - ${amountLabel}.`
            : `doesn't want to pay you rent - ${amountLabel}.`
          : isDemandActive
            ? `wants ${amountLabel}.`
            : `said no to paying you ${amountLabel}.`
    : isBirthdayDemand
      ? isDemandActive
        ? `is paying their party bill to ${sourceName} - ${amountLabel}.`
        : `dodged ${sourceName}'s party bill.`
      : isDebtCollector
        ? isDemandActive
          ? `is settling a debt with ${sourceName} - ${amountLabel}.`
          : `didn't settle the debt with ${sourceName}.`
        : isRent
          ? isDemandActive
            ? `is paying rent to ${sourceName} - ${amountLabel}.`
            : `didn't pay rent to ${sourceName}.`
          : isDemandActive
            ? `is answering a payment demand from ${sourceName} - ${amountLabel}.`
            : `didn't pay ${sourceName}.`;
  const demandLineText =
    isTargetingSelf && isSelectingCards && typeof selectedPaymentTotal === "number"
      ? `${demandLine.replace(/\.$/, "")} ($${selectedPaymentTotal}M selected).`
      : demandLine;
  const spectatorDemandLine = isBirthdayDemand
    ? isDemandActive
      ? (
          <>
            is paying their party bill to {sourceReference} - {amountLabel}.
          </>
        )
      : (
          <>
            dodged the party bill from {sourceReference}, they're responding.
          </>
        )
    : isDebtCollector
      ? isDemandActive
        ? (
            <>
              is settling a debt with {sourceReference} - {amountLabel}.
            </>
          )
        : (
            <>
              didn't settle the debt with {sourceReference}.
            </>
          )
      : isRent
        ? isDemandActive
          ? (
              <>
                is paying rent to {sourceReference} - {amountLabel}.
              </>
            )
          : (
              <>
                {isSelfSource
                  ? <>didn't pay rent to {targetName}, they're responding.</>
                  : <>didn't pay rent to {sourceReference}.</>}
              </>
            )
        : isDemandActive
          ? (
              <>
                is answering a payment demand from {sourceReference} - {amountLabel}.
              </>
            )
          : (
              <>
                didn't pay {sourceReference}.
              </>
            );
  const demandActorName = sourceName;
  const shouldUseSelfInactiveRentNarration = isRent && !isDemandActive && isSelfSource;

  return (
    <aside
      className="md-demand md-demand--payment"
      aria-live="polite"
      onPointerDown={(event) => event.stopPropagation()}
    >
      <p className="md-demand__eyebrow">{demandEyebrow}</p>
      {isTargetingSelf && isDemandActive ? (
        <div className="md-demand-source">
          <img
            className="md-demand-source__avatar"
            src={demandActor?.avatarUrl ?? ""}
            alt={demandActorName}
            loading="lazy"
            referrerPolicy="no-referrer"
          />
          <p className="md-demand__line md-demand-source__message">
            <span className="md-demand-source__name">{demandActorName}</span> {demandLineText}
          </p>
        </div>
      ) : isTargetingSelf ? (
        <p className="md-demand__line md-demand-source__message">
          {sourceInlineName} {demandLineText}
        </p>
      ) : (
        <p className="md-demand__line md-demand-source__message">
          {shouldUseSelfInactiveRentNarration ? (
            <>
              {sourceInlineActor} didn't pay rent to {targetInlineName}, they're responding.
            </>
          ) : (
            <>
              {targetInlineName} {spectatorDemandLine}
            </>
          )}
        </p>
      )}
      {isTargetingSelf ? (
        <div className="md-demand__actions">
          <button
            type="button"
            className="md-demand__button"
            onPointerDown={(event) => event.stopPropagation()}
            onClick={() => onComply(demand.id)}
            disabled={isSelectingCards && !canConfirmSelection}
            title={
              isSelectingCards && !canConfirmSelection
                ? "Selected cards must cover the payment amount"
                : undefined
            }
          >
            {isDemandActive ? "Pay" : "OK"}
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

export default PaymentDemandOverlay;
