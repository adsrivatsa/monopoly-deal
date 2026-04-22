import {
  type Card,
  DemandKind,
  type Demand,
  type Player,
} from "../../../../generated/monopoly_deal";
import PaymentDemandOverlay from "./PaymentDemandOverlay";
import PropertyDemandOverlay from "./PropertyDemandOverlay";
import PropertySetDemandOverlay from "./PropertySetDemandOverlay";

type DemandOverlayProps = {
  demands: Demand[];
  players: Player[];
  targetCardImageById: Record<string, string>;
  propertySetCardsById: Record<string, Card[]>;
  canDeny: boolean;
  isSelectingCards: boolean;
  selectingDemandId?: string;
  canConfirmSelection: boolean;
  selectedPaymentTotal: number;
  onComply: (demandId: string) => void;
  onDeny: (demandId: string) => void;
};

const DemandOverlay = ({
  demands,
  players,
  targetCardImageById,
  propertySetCardsById,
  canDeny,
  isSelectingCards,
  selectingDemandId,
  canConfirmSelection,
  selectedPaymentTotal,
  onComply,
  onDeny,
}: DemandOverlayProps) => {
  if (demands.length === 0) {
    return null;
  }

  return (
    <div className="md-demand-stack">
      {demands.map((demand) => {
        if (demand.demandKind === DemandKind.DEMAND_KIND_PAYMENT) {
          const isThisDemandSelecting = isSelectingCards && selectingDemandId === demand.id;
          return (
            <PaymentDemandOverlay
              key={demand.id}
              demand={demand}
              players={players}
              canDeny={canDeny}
              isDemandActive={demand.isActive}
              isSelectingCards={isThisDemandSelecting}
              canConfirmSelection={canConfirmSelection}
              selectedPaymentTotal={selectedPaymentTotal}
              onComply={onComply}
              onDeny={onDeny}
            />
          );
        }

        if (demand.demandKind === DemandKind.DEMAND_KIND_PROPERTY) {
          return (
            <PropertyDemandOverlay
              key={demand.id}
              demand={demand}
              players={players}
              targetCardImageUrl={
                demand.propertyDemand?.targetCardId
                  ? targetCardImageById[demand.propertyDemand.targetCardId]
                  : undefined
              }
              sourceCardImageUrl={
                demand.propertyDemand?.sourceCardId
                  ? targetCardImageById[demand.propertyDemand.sourceCardId]
                  : undefined
              }
              canDeny={canDeny}
              onComply={onComply}
              onDeny={onDeny}
            />
          );
        }

        if (demand.demandKind === DemandKind.DEMAND_KIND_PROPERTY_SET) {
          return (
            <PropertySetDemandOverlay
              key={demand.id}
              demand={demand}
              players={players}
              targetCardImageById={targetCardImageById}
              requestedCards={
                demand.propertySetDemand?.propertySetId
                  ? propertySetCardsById[demand.propertySetDemand.propertySetId] ?? []
                  : []
              }
              canDeny={canDeny}
              onComply={onComply}
              onDeny={onDeny}
            />
          );
        }

        return null;
      })}
    </div>
  );
};

export default DemandOverlay;
