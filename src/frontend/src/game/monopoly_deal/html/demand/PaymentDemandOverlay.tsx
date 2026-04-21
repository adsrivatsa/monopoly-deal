import { type Demand, type Player } from "../../../../generated/monopoly_deal";

type PaymentDemandOverlayProps = {
  demand: Demand;
  players: Player[];
  canDeny: boolean;
  isDemandActive: boolean;
  isSelectingCards: boolean;
  canConfirmSelection: boolean;
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
  onComply,
  onDeny,
}: PaymentDemandOverlayProps) => {
  const sourcePlayer = players.find((player) => player.playerId === demand.sourceId);
  const amount = demand.paymentDemand?.amount;

  return (
    <aside
      className="md-demand md-demand--payment"
      aria-live="polite"
      onPointerDown={(event) => event.stopPropagation()}
    >
      <p className="md-demand__eyebrow">Active demand</p>
      <h3 className="md-demand__title">Payment demand</h3>
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
      <p className="md-demand__line">
        Amount: ${typeof amount === "number" ? amount : "-"}
      </p>
      {!isDemandActive ? (
        <p className="md-demand__line">Status: Inactive (Comply sends no card transfer)</p>
      ) : null}
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
          {isSelectingCards ? "OK" : "Comply"}
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

export default PaymentDemandOverlay;
