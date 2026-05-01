import {
  type Card,
  DemandKind,
  DemandSource,
  type Demand,
  type Player,
} from "../../../../generated/monopoly_deal";
import PaymentDemandOverlay from "./PaymentDemandOverlay";
import PropertyDemandOverlay from "./PropertyDemandOverlay";
import PropertySetDemandOverlay from "./PropertySetDemandOverlay";

type DemandOverlayProps = {
  demands: Demand[];
  players: Player[];
  selfPlayerId?: string;
  demandTimerSeconds?: number;
  targetCardImageById: Record<string, string>;
  propertySetCardsById: Record<string, Card[]>;
  canDeny: boolean;
  isSelectingCards: boolean;
  selectingDemandId?: string;
  canConfirmSelection: boolean;
  selectedPaymentTotal: number;
  onComply: (demandId: string, propertySetId?: string) => void;
  forcedDealPlacementDemandId?: string;
  selectedForcedDealPlacementSetId?: string;
  onStartForcedDealPlacementSelection: (demandId: string) => void;
  onConfirmForcedDealPlacementSelection: (demandId: string) => void;
  onCancelForcedDealPlacementSelection: () => void;
  onDeny: (demandId: string) => void;
};

const DemandOverlay = ({
  demands,
  players,
  selfPlayerId,
  demandTimerSeconds,
  targetCardImageById,
  propertySetCardsById,
  canDeny,
  isSelectingCards,
  selectingDemandId,
  canConfirmSelection,
  selectedPaymentTotal,
  onComply,
  forcedDealPlacementDemandId,
  selectedForcedDealPlacementSetId,
  onStartForcedDealPlacementSelection,
  onConfirmForcedDealPlacementSelection,
  onCancelForcedDealPlacementSelection,
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
              selfPlayerId={selfPlayerId}
              canDeny={canDeny}
              isDemandActive={demand.isActive}
              isSelectingCards={isThisDemandSelecting}
              canConfirmSelection={canConfirmSelection}
              selectedPaymentTotal={selectedPaymentTotal}
              demandTimerSeconds={demandTimerSeconds}
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
              selfPlayerId={selfPlayerId}
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
              isChoosingForcedDealPlacement={
                demand.demandSource === DemandSource.DEMAND_SOURCE_FORCED_DEAL &&
                forcedDealPlacementDemandId === demand.id
              }
              selectedForcedDealPlacementSetId={selectedForcedDealPlacementSetId}
              demandTimerSeconds={demandTimerSeconds}
              onStartForcedDealPlacementSelection={
                onStartForcedDealPlacementSelection
              }
              onConfirmForcedDealPlacementSelection={
                onConfirmForcedDealPlacementSelection
              }
              onCancelForcedDealPlacementSelection={
                onCancelForcedDealPlacementSelection
              }
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
              selfPlayerId={selfPlayerId}
              targetCardImageById={targetCardImageById}
              requestedCards={
                demand.propertySetDemand?.propertySetId
                  ? propertySetCardsById[demand.propertySetDemand.propertySetId] ?? []
                  : []
              }
              demandTimerSeconds={demandTimerSeconds}
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
