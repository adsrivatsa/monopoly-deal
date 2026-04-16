import { appConfig } from "../config";
import { ClientMessage, ServerMessage } from "../generated/message";

export const connectRoomSocket = () => {
  const socket = new WebSocket(appConfig.room.socketUrl);
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

export const decodeRoomServerMessage = async (data: MessageEvent["data"]) => {
  const bytes = await toUint8Array(data);
  if (!bytes) {
    return null;
  }

  return ServerMessage.decode(bytes);
};

export const sendRoomChatMessage = (socket: WebSocket, payload: string) => {
  const message = ClientMessage.encode({
    roomMessage: {
      chat: {
        payload,
      },
    },
  }).finish();

  socket.send(message);
};
