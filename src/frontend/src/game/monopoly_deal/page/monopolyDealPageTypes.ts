import type { WonGame } from "../../../generated/monopoly_deal";

export type GameChatMessage = {
  id: string;
  playerId: string;
  text: string;
};

export type ActionHistoryEntry = {
  id: string;
  text: string;
  kind:
    | "generic"
    | "playMoney"
    | "playProperty"
    | "demandsCreated"
    | "playedCards"
    | "paymentComplied"
    | "discardCards"
    | "maskedDiscardCards"
    | "propertySwap"
    | "propertySetSteal"
    | "propertySteal"
    | "startTurn"
    | "maskedStartTurn"
    | "turnDivider";
  playerId?: string;
  cardAssetKey?: number;
  cardAssetKeys?: number[];
  drawCount?: number;
  sourcePlayerId?: string;
  targetPlayerId?: string;
  sourceCardAssetKey?: number;
  targetCardAssetKey?: number;
};

export type WonGameResult = WonGame & {
  displayName?: string;
  avatarUrl?: string;
};
