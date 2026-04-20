/// <reference types="vite/client" />

interface ImportMetaEnv {
  readonly VITE_BACKEND_URL?: string;
  readonly VITE_WS_DOMAIN?: string;
  readonly VITE_ROOM_SOCKET_URL?: string;
  readonly VITE_GAME_SOCKET_URL?: string;
}

interface ImportMeta {
  readonly env: ImportMetaEnv;
}
