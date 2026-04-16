import { apiFetch } from "./client";

export type GetPlayerParams = {
  player_id: string | null;
  email: string | null;
};

const defaultGetPlayerParams: GetPlayerParams = {
  player_id: null,
  email: null,
};

export const getPlayer = (params: Partial<GetPlayerParams> = {}) => {
  const body: GetPlayerParams = {
    ...defaultGetPlayerParams,
    ...params,
  };

  return apiFetch("/player", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body),
  });
};

export type UpdatePlayerParams = {
  display_name: string;
};

export const updatePlayer = (params: UpdatePlayerParams) => {
  return apiFetch("/player", {
    method: "PUT",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(params),
  });
};
