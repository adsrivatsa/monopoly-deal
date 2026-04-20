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
