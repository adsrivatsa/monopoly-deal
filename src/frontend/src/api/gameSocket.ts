import { appConfig } from "../config";
import {
  ClientMessage as GatewayClientMessage,
  ServerMessage,
} from "../generated/gateway";

export const connectGameSocket = () => {
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

export const decodeGameServerMessage = async (data: MessageEvent["data"]) => {
  const bytes = await toUint8Array(data);
  if (!bytes) {
    return null;
  }

  return ServerMessage.decode(bytes);
};

export const toGameServerMessageJson = (message: ServerMessage) => {
  return ServerMessage.toJSON(message);
};

export const sendGameChatMessage = (socket: WebSocket, payload: string) => {
  const message = GatewayClientMessage.encode({
    monopolyDealMessage: {
      chat: {
        payload,
      },
    },
  }).finish();

  socket.send(message);
};

export const sendGamePlayMoneyMessage = (socket: WebSocket, cardId: string) => {
  const message = GatewayClientMessage.encode({
    monopolyDealMessage: {
      playMoney: {
        cardId,
      },
    },
  }).finish();

  socket.send(message);
};

export const sendGamePlayPropertyMessage = (
  socket: WebSocket,
  payload: {
    cardId: string;
    propertySetId?: string;
    activeColor?: number;
  },
) => {
  const message = GatewayClientMessage.encode({
    monopolyDealMessage: {
      playProperty: {
        cardId: payload.cardId,
        propertySetId: payload.propertySetId,
        activeColor: payload.activeColor,
      },
    },
  }).finish();

  socket.send(message);
};

export const sendGamePlayHouseMessage = (
  socket: WebSocket,
  payload: {
    cardId: string;
    propertySetId: string;
  },
) => {
  const message = GatewayClientMessage.encode({
    monopolyDealMessage: {
      playHouse: {
        cardId: payload.cardId,
        propertySetId: payload.propertySetId,
      },
    },
  }).finish();

  socket.send(message);
};

export const sendGamePlayHotelMessage = (
  socket: WebSocket,
  payload: {
    cardId: string;
    propertySetId: string;
  },
) => {
  const message = GatewayClientMessage.encode({
    monopolyDealMessage: {
      playHotel: {
        cardId: payload.cardId,
        propertySetId: payload.propertySetId,
      },
    },
  }).finish();

  socket.send(message);
};

export const sendGameCompleteTurnMessage = (socket: WebSocket) => {
  const message = GatewayClientMessage.encode({
    monopolyDealMessage: {
      completeTurn: {},
    },
  }).finish();

  socket.send(message);
};

export const sendGamePlayPassGoMessage = (socket: WebSocket, cardId: string) => {
  const message = GatewayClientMessage.encode({
    monopolyDealMessage: {
      playPassGo: {
        cardId,
      },
    },
  }).finish();

  socket.send(message);
};

export const sendGamePlayItsMyBirthdayMessage = (
  socket: WebSocket,
  cardId: string,
) => {
  const message = GatewayClientMessage.encode({
    monopolyDealMessage: {
      playItsMyBirthday: {
        cardId,
      },
    },
  }).finish();

  socket.send(message);
};

export const sendGamePlayDebtCollectorMessage = (
  socket: WebSocket,
  payload: {
    cardId: string;
    targetId: string;
  },
) => {
  const message = GatewayClientMessage.encode({
    monopolyDealMessage: {
      playDebtCollector: {
        cardId: payload.cardId,
        targetId: payload.targetId,
      },
    },
  }).finish();

  socket.send(message);
};

export const sendGamePlayRentMessage = (socket: WebSocket, cardId: string) => {
  const message = GatewayClientMessage.encode({
    monopolyDealMessage: {
      playRent: {
        cardId,
      },
    },
  }).finish();

  socket.send(message);
};

export const sendGamePlayWildRentMessage = (
  socket: WebSocket,
  payload: {
    cardId: string;
    targetId: string;
  },
) => {
  const message = GatewayClientMessage.encode({
    monopolyDealMessage: {
      playWildRent: {
        cardId: payload.cardId,
        targetId: payload.targetId,
      },
    },
  }).finish();

  socket.send(message);
};

export const sendGamePlaySlyDealMessage = (
  socket: WebSocket,
  payload: {
    cardId: string;
    targetId: string;
    targetCardId: string;
  },
) => {
  const message = GatewayClientMessage.encode({
    monopolyDealMessage: {
      playSlyDeal: {
        cardId: payload.cardId,
        targetId: payload.targetId,
        targetCardId: payload.targetCardId,
      },
    },
  }).finish();

  socket.send(message);
};

export const sendGamePlayForcedDealMessage = (
  socket: WebSocket,
  payload: {
    cardId: string;
    targetId: string;
    sourceCardId: string;
    targetCardId: string;
  },
) => {
  const message = GatewayClientMessage.encode({
    monopolyDealMessage: {
      playForcedDeal: {
        cardId: payload.cardId,
        targetId: payload.targetId,
        sourceCardId: payload.sourceCardId,
        targetCardId: payload.targetCardId,
      },
    },
  }).finish();

  socket.send(message);
};

export const sendGamePlayDealBreakerMessage = (
  socket: WebSocket,
  payload: {
    cardId: string;
    targetId: string;
    propertySetId: string;
  },
) => {
  const message = GatewayClientMessage.encode({
    monopolyDealMessage: {
      playDealBreaker: {
        cardId: payload.cardId,
        targetId: payload.targetId,
        propertySetId: payload.propertySetId,
      },
    },
  }).finish();

  socket.send(message);
};

export const sendGamePlayDoubleTheRentMessage = (
  socket: WebSocket,
  cardId: string,
) => {
  const message = GatewayClientMessage.encode({
    monopolyDealMessage: {
      playDoubleTheRent: {
        cardId,
      },
    },
  }).finish();

  socket.send(message);
};

export const sendGameResolvePendingRentMessage = (socket: WebSocket) => {
  const message = GatewayClientMessage.encode({
    monopolyDealMessage: {
      resolvePendingRent: {},
    },
  }).finish();

  socket.send(message);
};

export const sendGameRearrangeCardMessage = (
  socket: WebSocket,
  payload: {
    cardId: string;
    propertySetId?: string;
    color?: number;
  },
) => {
  const message = GatewayClientMessage.encode({
    monopolyDealMessage: {
      rearrangeCard: {
        cardId: payload.cardId,
        propertySetId: payload.propertySetId,
        color: payload.color,
      },
    },
  }).finish();

  socket.send(message);
};

export const sendGameDiscardCardsMessage = (
  socket: WebSocket,
  cardIds: string[],
) => {
  const message = GatewayClientMessage.encode({
    monopolyDealMessage: {
      discardCards: {
        cardIds,
      },
    },
  }).finish();

  socket.send(message);
};

export const sendGameDenyDemandMessage = (
  socket: WebSocket,
  payload: {
    demandId: string;
    cardId: string;
  },
) => {
  const message = GatewayClientMessage.encode({
    monopolyDealMessage: {
      denyDemand: {
        demandId: payload.demandId,
        cardId: payload.cardId,
      },
    },
  }).finish();

  socket.send(message);
};

export const sendGameComplyPaymentDemandMessage = (
  socket: WebSocket,
  payload: {
    demandId: string;
    cardIds: string[];
  },
) => {
  const message = GatewayClientMessage.encode({
    monopolyDealMessage: {
      complyPaymentDemand: {
        demandId: payload.demandId,
        cardIds: payload.cardIds,
      },
    },
  }).finish();

  socket.send(message);
};

export const sendGameComplyPropertyDemandMessage = (
  socket: WebSocket,
  payload: {
    demandId: string;
    propertySetId?: string;
  },
) => {
  const message = GatewayClientMessage.encode({
    monopolyDealMessage: {
      complyPropertyDemand: {
        demandId: payload.demandId,
        propertySetId: payload.propertySetId,
      },
    },
  }).finish();

  socket.send(message);
};

export const sendGameComplyPropertySetDemandMessage = (
  socket: WebSocket,
  demandId: string,
) => {
  const message = GatewayClientMessage.encode({
    monopolyDealMessage: {
      complyPropertySetDemand: {
        demandId,
      },
    },
  }).finish();

  socket.send(message);
};
