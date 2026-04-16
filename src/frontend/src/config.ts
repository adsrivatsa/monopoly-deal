const DEFAULT_BACKEND_URL = "http://mdeal.io:4000";
const DEFAULT_ROOM_SOCKET_URL = "ws://127.0.0.1:4000/room/socket";

const trimTrailingSlash = (value: string) => value.replace(/\/+$/, "");

const backendUrl = trimTrailingSlash(
  import.meta.env.VITE_BACKEND_URL ?? DEFAULT_BACKEND_URL,
);

const roomSocketUrl = import.meta.env.VITE_ROOM_SOCKET_URL ?? DEFAULT_ROOM_SOCKET_URL;

export const appConfig = {
  backendUrl,
  auth: {
    googleLoginUrl: `${backendUrl}/auth/google/login`,
  },
  room: {
    socketUrl: roomSocketUrl,
    create: {
      games: ["monopoly_deal"],
      capacity: {
        min: 2,
        max: 5,
      },
    },
  },
} as const;
