import {
  apiFetch,
  ApiErrorPayload,
  isTokenErrorCode,
  readApiError,
} from "./client";
import { Game } from "./models";

export type ListRoomsParams = {
  limit: number;
  offset: number;
  search: string | null;
  game: Game | null;
};

export type ShortPlayer = {
  player_id: string;
  display_name: string;
  image_url: string;
};

export type RoomPlayer = ShortPlayer & {
  is_host: boolean;
  is_ready: boolean;
};

export type RoomListItem = {
  room_id: string;
  display_name: string;
  capacity: number;
  occupied: number;
  players: ShortPlayer[];
  game: Game;
};

export type ListRoomsResponse = {
  total_count: number;
  rooms: RoomListItem[];
};

export type CreateRoomParams = {
  display_name: string;
  capacity: number;
  game: Game;
  settings: string;
};

export type CreateRoomResponse = {
  room_id: string;
  display_name: string;
  capacity: number;
  occupied: number;
  created_at: string;
};

export type RoomResponse = {
  room_id: string;
  display_name: string;
  capacity: number;
  occupied: number;
  game: Game;
  settings: string;
  players: RoomPlayer[];
};

export type ListRoomsResult =
  | { ok: true; data: ListRoomsResponse }
  | { ok: false; error: ApiErrorPayload | null; isTokenError: boolean };

export type JoinRoomResult =
  | { ok: true }
  | { ok: false; error: ApiErrorPayload | null; isTokenError: boolean };

export type CreateRoomResult =
  | { ok: true; data: CreateRoomResponse }
  | { ok: false; error: ApiErrorPayload | null; isTokenError: boolean };

export type GetRoomResult =
  | { ok: true; data: RoomResponse }
  | { ok: false; error: ApiErrorPayload | null; isTokenError: boolean };

export type LeaveRoomResult =
  | { ok: true }
  | { ok: false; error: ApiErrorPayload | null; isTokenError: boolean };

export type ReadyRoomResult =
  | { ok: true }
  | { ok: false; error: ApiErrorPayload | null; isTokenError: boolean };

export type UpdateRoomSettingsParams = {
  capacity: number;
  game: Game;
  settings: string;
};

export type UpdateRoomSettingsResult =
  | { ok: true }
  | { ok: false; error: ApiErrorPayload | null; isTokenError: boolean };

const defaultListRoomsParams: ListRoomsParams = {
  limit: 10,
  offset: 0,
  search: null,
  game: null,
};

export const listRooms = async (
  params: Partial<ListRoomsParams> = {},
): Promise<ListRoomsResult> => {
  const body: ListRoomsParams = {
    ...defaultListRoomsParams,
    ...params,
  };

  const response = await apiFetch("/room/list", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body),
  });

  if (response.status !== 200) {
    const error = await readApiError(response);
    return {
      ok: false,
      error,
      isTokenError: isTokenErrorCode(error?.code),
    };
  }

  return {
    ok: true,
    data: (await response.json()) as ListRoomsResponse,
  };
};

export const joinRoom = async (roomId: string): Promise<JoinRoomResult> => {
  const response = await apiFetch(`/room/join/${roomId}`, {
    method: "PATCH",
  });

  if (response.status !== 200) {
    const error = await readApiError(response);
    return {
      ok: false,
      error,
      isTokenError: isTokenErrorCode(error?.code),
    };
  }

  return { ok: true };
};

export const createRoom = async (
  params: CreateRoomParams,
): Promise<CreateRoomResult> => {
  const response = await apiFetch("/room/", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(params),
  });

  if (response.status !== 200) {
    const error = await readApiError(response);
    return {
      ok: false,
      error,
      isTokenError: isTokenErrorCode(error?.code),
    };
  }

  return { ok: true, data: (await response.json()) as CreateRoomResponse };
};

export const getRoom = async (): Promise<GetRoomResult> => {
  const response = await apiFetch("/room", {
    method: "GET",
  });

  if (response.status !== 200) {
    const error = await readApiError(response);
    return {
      ok: false,
      error,
      isTokenError: isTokenErrorCode(error?.code),
    };
  }

  return { ok: true, data: (await response.json()) as RoomResponse };
};

export const leaveRoom = async (): Promise<LeaveRoomResult> => {
  const response = await apiFetch("/room/leave", {
    method: "PATCH",
  });

  if (response.status !== 200) {
    const error = await readApiError(response);
    return {
      ok: false,
      error,
      isTokenError: isTokenErrorCode(error?.code),
    };
  }

  return { ok: true };
};

export const readyRoom = async (): Promise<ReadyRoomResult> => {
  const response = await apiFetch("/room/ready", {
    method: "PATCH",
  });

  if (response.status !== 200) {
    const error = await readApiError(response);
    return {
      ok: false,
      error,
      isTokenError: isTokenErrorCode(error?.code),
    };
  }

  return { ok: true };
};

export const updateRoomSettings = async (
  params: UpdateRoomSettingsParams,
): Promise<UpdateRoomSettingsResult> => {
  const response = await apiFetch("/room/settings", {
    method: "PUT",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(params),
  });

  if (response.status !== 200) {
    const error = await readApiError(response);
    return {
      ok: false,
      error,
      isTokenError: isTokenErrorCode(error?.code),
    };
  }

  return { ok: true };
};
