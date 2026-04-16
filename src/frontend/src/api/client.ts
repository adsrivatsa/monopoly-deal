import { appConfig } from "../config";

export type ApiErrorPayload = {
  message: string;
  status: number;
  code: string;
};

const buildApiUrl = (path: string) => {
  return new URL(path, appConfig.backendUrl).toString();
};

const requestApi = (path: string, init?: RequestInit) => {
  return fetch(buildApiUrl(path), {
    credentials: "include",
    ...init,
  });
};

export const isTokenErrorCode = (code?: string | null) => {
  return typeof code === "string" && code.startsWith("TOK");
};

export const readApiError = async (
  response: Response,
): Promise<ApiErrorPayload | null> => {
  if (response.ok) {
    return null;
  }

  try {
    const payload = (await response.clone().json()) as Partial<ApiErrorPayload>;
    if (
      typeof payload.message !== "string" ||
      typeof payload.status !== "number" ||
      typeof payload.code !== "string"
    ) {
      return null;
    }

    return {
      message: payload.message,
      status: payload.status,
      code: payload.code,
    };
  } catch {
    return null;
  }
};

export const refreshAuthToken = async () => {
  const response = await requestApi("/auth/refresh", { method: "GET" });
  return response.ok;
};

export const logoutGoogleAuth = async () => {
  await requestApi("/auth/google/logout", { method: "GET" });
};

export const apiFetch = async (path: string, init?: RequestInit) => {
  const response = await requestApi(path, init);
  const errorPayload = await readApiError(response);
  const hasTokenError = isTokenErrorCode(errorPayload?.code);

  if (
    !hasTokenError ||
    path === "/auth/refresh" ||
    path === "/auth/google/logout"
  ) {
    return response;
  }

  const didRefresh = await refreshAuthToken();

  if (!didRefresh) {
    await logoutGoogleAuth();
    return response;
  }

  return requestApi(path, init);
};
