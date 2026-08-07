// Socket wrapper for "Monopoly Deal: No Mercy". Mirrors gameSocket.ts (the
// classic wrapper) but targets the gateway's `dealNoMercyMessage` arm and the
// generated types in generated/deal_no_mercy.ts. There is exactly one send
// function per client message the service's HandleEvent switch dispatches
// (see src/backend/internal/service/deal-no-mercy/event.go).
import { appConfig } from "../config";
import {
  ClientMessage as GatewayClientMessage,
  ServerMessage as GatewayServerMessage,
} from "../generated/gateway";
import {
  Color,
  ServerMessage,
  StealCategory,
  type PropertyPick,
  type DistributionAssignment,
} from "../generated/deal_no_mercy";

// The No Mercy game shares the classic `/game/socket` endpoint; the server
// resolves the game type from the player's active game, so the URL builder is
// identical to the classic wrapper.
export const connectDealNoMercyGameSocket = () => {
  const socket = new WebSocket(appConfig.game.socketUrl);
  socket.binaryType = "arraybuffer";
  return socket;
};

const toUint8Array = async (
  data: MessageEvent["data"],
): Promise<Uint8Array | null> => {
  if (data instanceof ArrayBuffer) {
    return new Uint8Array(data);
  }

  if (data instanceof Blob) {
    const buffer = await data.arrayBuffer();
    return new Uint8Array(buffer);
  }

  return null;
};

// The server always sends a gateway ServerMessage wrapping the deal_no_mercy
// payload in its `dealNoMercyMessage` arm, so decode the gateway envelope
// first and return the unwrapped inner message (null when the arm is absent).
export const decodeDealNoMercyServerMessage = async (
  data: MessageEvent["data"],
) => {
  const bytes = await toUint8Array(data);
  if (!bytes) {
    return null;
  }

  const gateway = GatewayServerMessage.decode(bytes);
  return gateway.dealNoMercyMessage ?? null;
};

export const toDealNoMercyServerMessageJson = (message: ServerMessage) => {
  return ServerMessage.toJSON(message);
};

// send wraps the dealNoMercyMessage payload in the gateway ClientMessage and
// writes the encoded bytes to the socket. Every send function below funnels
// through here.
const send = (
  socket: WebSocket,
  message: NonNullable<GatewayClientMessage["dealNoMercyMessage"]>,
) => {
  const encoded = GatewayClientMessage.encode({
    dealNoMercyMessage: message,
  }).finish();

  socket.send(encoded);
};

// ---------------------------------------------------------------------------
// Chat
// ---------------------------------------------------------------------------

export const sendDealNoMercyChatMessage = (
  socket: WebSocket,
  payload: string,
) => {
  send(socket, { chat: { payload } });
};

// ---------------------------------------------------------------------------
// Turn lifecycle
// ---------------------------------------------------------------------------

export const sendDealNoMercyCompleteTurnMessage = (socket: WebSocket) => {
  send(socket, { completeTurn: {} });
};

export const sendDealNoMercyDiscardCardsMessage = (
  socket: WebSocket,
  cardIds: string[],
) => {
  send(socket, { discardCards: { cardIds } });
};

export const sendDealNoMercyRearrangeCardMessage = (
  socket: WebSocket,
  payload: {
    cardId: string;
    propertySetId?: string;
    color?: Color;
  },
) => {
  send(socket, {
    rearrangeCard: {
      cardId: payload.cardId,
      propertySetId: payload.propertySetId,
      color: payload.color,
    },
  });
};

// ---------------------------------------------------------------------------
// Plays
// ---------------------------------------------------------------------------

export const sendDealNoMercyPlayMoneyMessage = (
  socket: WebSocket,
  cardId: string,
) => {
  send(socket, { playMoney: { cardId } });
};

export const sendDealNoMercyPlayPropertyMessage = (
  socket: WebSocket,
  payload: {
    cardId: string;
    propertySetId?: string;
    activeColor?: Color;
  },
) => {
  send(socket, {
    playProperty: {
      cardId: payload.cardId,
      propertySetId: payload.propertySetId,
      activeColor: payload.activeColor,
    },
  });
};

export const sendDealNoMercyPlayShackMessage = (
  socket: WebSocket,
  payload: {
    cardId: string;
    propertySetId: string;
  },
) => {
  send(socket, {
    playShack: {
      cardId: payload.cardId,
      propertySetId: payload.propertySetId,
    },
  });
};

export const sendDealNoMercyPlayBigPaydayMessage = (
  socket: WebSocket,
  cardId: string,
) => {
  send(socket, { playBigPayday: { cardId } });
};

export const sendDealNoMercyPlayGoAgainMessage = (
  socket: WebSocket,
  cardId: string,
) => {
  send(socket, { playGoAgain: { cardId } });
};

export const sendDealNoMercyPlaySetSnatcherMessage = (
  socket: WebSocket,
  payload: {
    cardId: string;
    targetId: string;
    propertySetId: string;
  },
) => {
  send(socket, {
    playSetSnatcher: {
      cardId: payload.cardId,
      targetId: payload.targetId,
      propertySetId: payload.propertySetId,
    },
  });
};

export const sendDealNoMercyPlayDebtTrapMessage = (
  socket: WebSocket,
  payload: {
    cardId: string;
    targetId: string;
  },
) => {
  send(socket, {
    playDebtTrap: {
      cardId: payload.cardId,
      targetId: payload.targetId,
    },
  });
};

export const sendDealNoMercyPlayHeistMessage = (
  socket: WebSocket,
  payload: {
    cardId: string;
    // One pick per target player: the card that player must surrender.
    picks: PropertyPick[];
  },
) => {
  send(socket, {
    playHeist: {
      cardId: payload.cardId,
      picks: payload.picks,
    },
  });
};

export const sendDealNoMercyPlayMarketCrashMessage = (
  socket: WebSocket,
  payload: {
    cardId: string;
    // One pick per target player: the property that player must surrender.
    picks: PropertyPick[];
  },
) => {
  send(socket, {
    playMarketCrash: {
      cardId: payload.cardId,
      picks: payload.picks,
    },
  });
};

export const sendDealNoMercyPlayRepoManMessage = (
  socket: WebSocket,
  payload: {
    cardId: string;
    targetId: string;
  },
) => {
  send(socket, {
    playRepoMan: {
      cardId: payload.cardId,
      targetId: payload.targetId,
    },
  });
};

export const sendDealNoMercyPlayPropertyRaidMessage = (
  socket: WebSocket,
  payload: {
    cardId: string;
    color: Color;
  },
) => {
  send(socket, {
    playPropertyRaid: {
      cardId: payload.cardId,
      color: payload.color,
    },
  });
};

export const sendDealNoMercyPlayTaxDayMessage = (
  socket: WebSocket,
  payload: {
    cardId: string;
    targetId: string;
  },
) => {
  send(socket, {
    playTaxDay: {
      cardId: payload.cardId,
      targetId: payload.targetId,
    },
  });
};

export const sendDealNoMercyPlayPickpocketMessage = (
  socket: WebSocket,
  payload: {
    cardId: string;
    targetId: string;
    category: StealCategory;
  },
) => {
  send(socket, {
    playPickpocket: {
      cardId: payload.cardId,
      targetId: payload.targetId,
      category: payload.category,
    },
  });
};

export const sendDealNoMercyPlayBankSwapMessage = (
  socket: WebSocket,
  payload: {
    cardId: string;
    targetId: string;
  },
) => {
  send(socket, {
    playBankSwap: {
      cardId: payload.cardId,
      targetId: payload.targetId,
    },
  });
};

export const sendDealNoMercyPlayYoinkMessage = (
  socket: WebSocket,
  payload: {
    cardId: string;
    targetId: string;
  },
) => {
  send(socket, {
    playYoink: {
      cardId: payload.cardId,
      targetId: payload.targetId,
    },
  });
};

export const sendDealNoMercyPlayRentMessage = (
  socket: WebSocket,
  payload: {
    cardId: string;
    color: Color;
  },
) => {
  send(socket, {
    playRent: {
      cardId: payload.cardId,
      color: payload.color,
    },
  });
};

// ---------------------------------------------------------------------------
// Demand resolution (comply / deny for every demand kind)
// ---------------------------------------------------------------------------

export const sendDealNoMercyComplyPaymentDemandMessage = (
  socket: WebSocket,
  payload: {
    demandId: string;
    cardIds: string[];
  },
) => {
  send(socket, {
    complyPaymentDemand: {
      demandId: payload.demandId,
      cardIds: payload.cardIds,
    },
  });
};

export const sendDealNoMercyComplyPropertyDemandMessage = (
  socket: WebSocket,
  demandId: string,
) => {
  send(socket, { complyPropertyDemand: { demandId } });
};

export const sendDealNoMercyComplyPropertySetDemandMessage = (
  socket: WebSocket,
  demandId: string,
) => {
  send(socket, { complyPropertySetDemand: { demandId } });
};

export const sendDealNoMercyComplyColorPropertiesDemandMessage = (
  socket: WebSocket,
  demandId: string,
) => {
  send(socket, { complyColorPropertiesDemand: { demandId } });
};

export const sendDealNoMercyComplyBankCardDemandMessage = (
  socket: WebSocket,
  demandId: string,
) => {
  send(socket, { complyBankCardDemand: { demandId } });
};

export const sendDealNoMercyComplyBankSwapDemandMessage = (
  socket: WebSocket,
  demandId: string,
) => {
  send(socket, { complyBankSwapDemand: { demandId } });
};

export const sendDealNoMercyComplyDebtTrapDemandMessage = (
  socket: WebSocket,
  demandId: string,
) => {
  send(socket, { complyDebtTrapDemand: { demandId } });
};

// REPO MAN comply: keep one card, distribute the rest (card -> recipient).
export const sendDealNoMercyComplyRepoManDemandMessage = (
  socket: WebSocket,
  payload: {
    demandId: string;
    keepCardId: string;
    distribution: DistributionAssignment[];
  },
) => {
  send(socket, {
    complyRepoManDemand: {
      demandId: payload.demandId,
      keepCardId: payload.keepCardId,
      distribution: payload.distribution,
    },
  });
};

// TAX DAY comply: keep one card, distribute the rest (card -> recipient).
export const sendDealNoMercyComplyTaxDayDemandMessage = (
  socket: WebSocket,
  payload: {
    demandId: string;
    keepCardId: string;
    distribution: DistributionAssignment[];
  },
) => {
  send(socket, {
    complyTaxDayDemand: {
      demandId: payload.demandId,
      keepCardId: payload.keepCardId,
      distribution: payload.distribution,
    },
  });
};

export const sendDealNoMercyComplyPickpocketDemandMessage = (
  socket: WebSocket,
  demandId: string,
) => {
  send(socket, { complyPickpocketDemand: { demandId } });
};

// NAH! — deny a demand with a NAH! card from hand.
export const sendDealNoMercyDenyDemandMessage = (
  socket: WebSocket,
  payload: {
    demandId: string;
    cardId: string;
  },
) => {
  send(socket, {
    denyDemand: {
      demandId: payload.demandId,
      cardId: payload.cardId,
    },
  });
};

// ---------------------------------------------------------------------------
// Debt settlement
// ---------------------------------------------------------------------------

export const sendDealNoMercySettleDebtMessage = (
  socket: WebSocket,
  payload: {
    debtId: string;
    cardId: string;
  },
) => {
  send(socket, {
    settleDebt: {
      debtId: payload.debtId,
      cardId: payload.cardId,
    },
  });
};
