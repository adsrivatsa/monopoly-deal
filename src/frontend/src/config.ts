const DEFAULT_BACKEND_URL = "http://127.0.0.1:4000";
const DEFAULT_WS_DOMAIN = "ws://127.0.0.1:4000";

const trimTrailingSlash = (value: string) => value.replace(/\/+$/, "");

const backendUrl = trimTrailingSlash(
  import.meta.env.VITE_BACKEND_URL ?? DEFAULT_BACKEND_URL,
);

const wsDomain = trimTrailingSlash(
  import.meta.env.VITE_WS_DOMAIN ?? DEFAULT_WS_DOMAIN,
);

const roomSocketUrl =
  import.meta.env.VITE_ROOM_SOCKET_URL ?? `${wsDomain}/room/socket`;
const gameSocketUrl =
  import.meta.env.VITE_GAME_SOCKET_URL ?? `${wsDomain}/game/socket`;
const buildQuickPlaySocketUrl = (game: string) =>
  `${wsDomain}/room/quick-play/${game}/socket`;

export const appConfig = {
  backendUrl,
  auth: {
    googleLoginUrl: `${backendUrl}/auth/google/login`,
  },
  room: {
    socketUrl: roomSocketUrl,
    quickPlaySocketUrl: buildQuickPlaySocketUrl,
    create: {
      games: ["monopoly_deal"],
      capacity: {
        min: 2,
        max: 5,
      },
    },
  },
  game: {
    socketUrl: gameSocketUrl,
  },
} as const;
