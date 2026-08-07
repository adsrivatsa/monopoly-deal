import type { WonGame } from "../../../generated/deal_no_mercy";

export type GameChatMessage = {
  id: string;
  playerId: string;
  text: string;
};

// ActionHistoryEntry is a normalized, replay-friendly representation of one
// Action for the sidebar log. Each `kind` selects a renderer in the sidebar;
// the extra fields carry the card asset keys / player ids that renderer needs.
export type ActionHistoryEntry = {
  id: string;
  text: string;
  kind:
    | "generic"
    | "playMoney"
    | "playProperty"
    | "playShack"
    | "playAction"
    | "demandsCreated"
    | "denied"
    | "paymentComplied"
    | "propertySteal"
    | "propertySetSteal"
    | "bankSwap"
    | "distribution"
    | "pickpocket"
    | "debtTrap"
    | "debtSettled"
    | "bigPayday"
    | "goAgain"
    | "discardCards"
    | "maskedDiscardCards"
    | "startTurn"
    | "maskedStartTurn"
    | "turnDivider";
  playerId?: string;
  cardAssetKey?: number;
  cardAssetKeys?: number[];
  drawCount?: number;
  amount?: number;
  sourcePlayerId?: string;
  targetPlayerId?: string;
};

export type WonGameResult = WonGame & {
  displayName?: string;
  avatarUrl?: string;
};
