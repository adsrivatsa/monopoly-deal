import { type Color, type GameState } from "../../generated/monopoly_deal";
import MonopolyDealHtmlBoard from "./html/MonopolyDealHtmlBoard";

type MonopolyDealGameMountProps = {
  initialGameState: GameState | null;
  assetImageByKey: Record<number, string>;
  selfPlayerId?: string;
  onPlayMoneyCard: (cardId: string) => void;
  onPlayPassGoCard: (cardId: string) => void;
  onPlayDebtCollectorCard: (cardId: string, targetPlayerId: string) => void;
  onPlayWildRentCard: (cardId: string, targetPlayerId: string) => void;
  onPlaySlyDealCard: (
    cardId: string,
    targetPlayerId: string,
    targetCardId: string,
  ) => void;
  onPlayDealBreakerCard: (
    cardId: string,
    targetPlayerId: string,
    propertySetId: string,
  ) => void;
  onPlayForcedDealCard: (
    cardId: string,
    targetPlayerId: string,
    sourceCardId: string,
    targetCardId: string,
  ) => void;
  onPlayDoubleTheRentCard: () => void;
  onResolvePendingRent: () => void;
  onRearrangeCard: (
    cardId: string,
    propertySetId?: string,
    color?: Color,
  ) => void;
  onDenyDemand: (demandId: string) => void;
  isDiscardRequired: boolean;
  requiredDiscardCount: number;
  selectedDiscardCardIds: ReadonlySet<string>;
  onToggleDiscardCard: (cardId: string) => void;
  onPlayPropertyCard: (
    cardId: string,
    propertySetId?: string,
    activeColor?: Color,
  ) => void;
  onComplyPaymentDemand: (demandId: string, cardIds: string[]) => void;
  onComplyPropertyDemand: (demandId: string) => void;
  onComplyPropertySetDemand: (demandId: string) => void;
  onClientError?: (error: unknown) => void;
  onGameError?: (error: { message: string; code: string }) => void;
};

const MonopolyDealGameMount = ({
  initialGameState,
  assetImageByKey,
  selfPlayerId,
  onPlayMoneyCard,
  onPlayPassGoCard,
  onPlayDebtCollectorCard,
  onPlayWildRentCard,
  onPlaySlyDealCard,
  onPlayDealBreakerCard,
  onPlayForcedDealCard,
  onPlayDoubleTheRentCard,
  onResolvePendingRent,
  onRearrangeCard,
  onDenyDemand,
  isDiscardRequired,
  requiredDiscardCount,
  selectedDiscardCardIds,
  onToggleDiscardCard,
  onPlayPropertyCard,
  onComplyPaymentDemand,
  onComplyPropertyDemand,
  onComplyPropertySetDemand,
  onClientError,
  onGameError,
}: MonopolyDealGameMountProps) => {
  return (
    <MonopolyDealHtmlBoard
      gameState={initialGameState}
      assetImageByKey={assetImageByKey}
      selfPlayerId={selfPlayerId}
      onPlayMoneyCard={onPlayMoneyCard}
      onPlayPassGoCard={onPlayPassGoCard}
      onPlayDebtCollectorCard={onPlayDebtCollectorCard}
      onPlayWildRentCard={onPlayWildRentCard}
      onPlaySlyDealCard={onPlaySlyDealCard}
      onPlayDealBreakerCard={onPlayDealBreakerCard}
      onPlayForcedDealCard={onPlayForcedDealCard}
      onPlayDoubleTheRentCard={onPlayDoubleTheRentCard}
      onResolvePendingRent={onResolvePendingRent}
      onRearrangeCard={onRearrangeCard}
      onDenyDemand={onDenyDemand}
      isDiscardRequired={isDiscardRequired}
      requiredDiscardCount={requiredDiscardCount}
      selectedDiscardCardIds={selectedDiscardCardIds}
      onToggleDiscardCard={onToggleDiscardCard}
      onPlayPropertyCard={onPlayPropertyCard}
      onComplyPaymentDemand={onComplyPaymentDemand}
      onComplyPropertyDemand={onComplyPropertyDemand}
      onComplyPropertySetDemand={onComplyPropertySetDemand}
      onClientError={onClientError}
      onGameError={onGameError}
    />
  );
};

export default MonopolyDealGameMount;
