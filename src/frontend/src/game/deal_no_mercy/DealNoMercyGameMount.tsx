// DealNoMercyGameMount is a thin pass-through wrapper around DealNoMercyBoard,
// mirroring MonopolyDealGameMount: the page owns all interaction logic and the
// socket lifecycle, and threads the play/comply/deny callbacks down through the
// mount to the presentational board. Keeping this layer thin means the board
// stays a pure view over GameState + callbacks.
import type {
  Color,
  DistributionAssignment,
  GameState,
  PropertyPick,
  StealCategory,
} from "../../generated/deal_no_mercy";
import DealNoMercyBoard from "./board/DealNoMercyBoard";

export type DealNoMercyBoardViewMode = "expanded" | "compact";

export type DealNoMercyGameMountProps = {
  gameState: GameState | null;
  assetImageByKey: Record<number, string>;
  viewMode: DealNoMercyBoardViewMode;
  selfPlayerId: string | null;
  demandDeadlineMsById: Record<string, number>;
  turnDeadlineMs: number;

  // Turn lifecycle.
  onPassTurn: () => void;
  onSubmitDiscard: () => void;
  isDiscardRequired: boolean;
  requiredDiscardCount: number;
  selectedDiscardCardIds: ReadonlySet<string>;
  onToggleDiscardCard: (cardId: string) => void;

  // Simple / property plays.
  onPlayMoney: (cardId: string) => void;
  onPlayProperty: (cardId: string, propertySetId?: string, activeColor?: Color) => void;
  onRearrangeCard: (cardId: string, propertySetId?: string, color?: Color) => void;
  onPlayShack: (cardId: string, propertySetId: string) => void;
  onPlayBigPayday: (cardId: string) => void;
  onPlayGoAgain: (cardId: string) => void;

  // Targeted / choice action plays.
  onPlaySetSnatcher: (cardId: string, targetId: string, propertySetId: string) => void;
  onPlayDebtTrap: (cardId: string, targetId: string) => void;
  onPlayYoink: (cardId: string, targetId: string) => void;
  onPlayBankSwap: (cardId: string, targetId: string) => void;
  onPlayRepoMan: (cardId: string, targetId: string) => void;
  onPlayTaxDay: (cardId: string, targetId: string) => void;
  onPlayPickpocket: (cardId: string, targetId: string, category: StealCategory) => void;
  onPlayPropertyRaid: (cardId: string, color: Color) => void;
  onPlayRent: (cardId: string, color: Color) => void;
  onPlayHeist: (cardId: string, picks: PropertyPick[]) => void;
  onPlayMarketCrash: (cardId: string, picks: PropertyPick[]) => void;

  // Demand resolution.
  onComplyPaymentDemand: (demandId: string, cardIds: string[]) => void;
  onComplyPropertyDemand: (demandId: string) => void;
  onComplyPropertySetDemand: (demandId: string) => void;
  onComplyColorPropertiesDemand: (demandId: string) => void;
  onComplyBankCardDemand: (demandId: string) => void;
  onComplyBankSwapDemand: (demandId: string) => void;
  onComplyDebtTrapDemand: (demandId: string) => void;
  onComplyPickpocketDemand: (demandId: string) => void;
  onComplyRepoManDemand: (
    demandId: string,
    keepCardId: string,
    distribution: DistributionAssignment[],
  ) => void;
  onComplyTaxDayDemand: (
    demandId: string,
    keepCardId: string,
    distribution: DistributionAssignment[],
  ) => void;
  onDenyDemand: (demandId: string, cardId: string) => void;

  // Debt settlement.
  onSettleDebt: (debtId: string, cardId: string) => void;

  onGameError?: (error: { message: string; code: string }) => void;
};

const DealNoMercyGameMount = (props: DealNoMercyGameMountProps) => {
  const { gameState, viewMode, selfPlayerId, ...rest } = props;
  return (
    <DealNoMercyBoard
      gameState={gameState}
      layoutMode={viewMode}
      selfPlayerId={selfPlayerId ?? undefined}
      {...rest}
    />
  );
};

export default DealNoMercyGameMount;
