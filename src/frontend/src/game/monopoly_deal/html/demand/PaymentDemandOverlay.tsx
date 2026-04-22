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

type PaymentDemandOverlayProps = {
  demand: Demand;
  players: Player[];
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
  canDeny,
  isDemandActive,
  isSelectingCards,
  canConfirmSelection,
  selectedPaymentTotal,
  onComply,
  onDeny,
}: PaymentDemandOverlayProps) => {
  const sourcePlayer = players.find((player) => player.playerId === demand.sourceId);
  const amount = demand.paymentDemand?.amount;
  const isBirthdayDemand = isItsMyBirthdayDemand(demand);
  const isDebtCollector = isDebtCollectorDemand(demand);
  const isRent = isRentDemand(demand);
  const sourceName = sourcePlayer?.displayName ?? demand.sourceId;
  const demandLine = isBirthdayDemand
    ? isDemandActive
      ? `is throwing a party. Pay $${typeof amount === "number" ? amount : "-"}M to attend.`
      : `doesn't want to come to your party :(.`
    : isDebtCollector
      ? isDemandActive
        ? `wants you to settle up. Pay them $${typeof amount === "number" ? amount : "-"}M.`
        : `wants to settle up $${typeof amount === "number" ? amount : "-"}M later (probably never).`
      : isRent
        ? isDemandActive
        ? `wants you to pay your rent - $${typeof amount === "number" ? amount : "-"}M.`
        : `doesn't want to pay you rent - $${typeof amount === "number" ? amount : "-"}M.`
      : isDemandActive
        ? `wants $${typeof amount === "number" ? amount : "-"}.`
        : `said no to paying you $${typeof amount === "number" ? amount : "-"}.`;
  const demandLineText =
    isSelectingCards && typeof selectedPaymentTotal === "number"
      ? `${demandLine.replace(/\.$/, "")} ($${selectedPaymentTotal}M selected).`
      : demandLine;

  return (
    <aside
      className="md-demand md-demand--payment"
      aria-live="polite"
      onPointerDown={(event) => event.stopPropagation()}
    >
      <h3 className="md-demand__title">
        {isDemandActive ? "Payment requested" : "Payment denied"}
      </h3>
      <div className="md-demand-source">
        <img
          className="md-demand-source__avatar"
          src={sourcePlayer?.avatarUrl ?? ""}
          alt={sourcePlayer?.displayName ?? demand.sourceId}
          loading="lazy"
          referrerPolicy="no-referrer"
        />
        <p className="md-demand__line md-demand-source__message">
          <span className="md-demand-source__name">{sourceName}</span> {demandLineText}
        </p>
      </div>
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
    </aside>
  );
};

export default PaymentDemandOverlay;
