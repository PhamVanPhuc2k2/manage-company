import { isTokenExpiringSoon, useAuthStore } from "./auth/useAuth";

export const API_BASE = process.env.NEXT_PUBLIC_API_URL || "/api/v1";

export class ApiError extends Error {
  constructor(
    public status: number,
    public code: string,
    message: string,
    public details?: unknown,
    public requestId?: string,
  ) {
    super(message);
    this.name = "ApiError";
  }
}

type Envelope<T> = {
  data?: T;
  meta?: unknown;
  error?: {
    code: string;
    message: string;
    details?: unknown;
    request_id?: string;
  };
};

export type PageMeta = {
  page: number;
  pagq_size: number;
  total_items: number;
  total_page: number;
};

let refreshPromise: Promise<string | null> | null = null;

async function doRefresh(): Promise<string | null> {
  const auth = useAuthStore.getState();

  try {
    const res = await fetch(`${API_BASE}/auth/refresh`, {
      method: "POST",
      credentials: "include",
      headers: { "Content-Type": "application/json" },
    });

    if (!res.ok) {
      auth.clear();
      return null;
    }

    const body = (await res.json()) as Envelope<{
      access_token: string;
      expires_at: string;
    }>;

    if (!body.data?.access_token) {
      auth.clear();
      return null;
    }

    auth.setToken(body.data.access_token, body.data.expires_at);
    return body.data.access_token;
  } catch {
    auth.clear();
    return null;
  }
}

export async function refreshAccessToken(): Promise<string | null> {
  refreshPromise ??= doRefresh().finally(() => {
    refreshPromise = null;
  });
  return refreshPromise;
}

async function getValidToken(): Promise<string | null> {
  const token = useAuthStore.getState().accessToken;
  if (token && !isTokenExpiringSoon()) return token;
  return refreshAccessToken();
}

type RequestOptions = Omit<RequestInit, "body"> & {
  body?: unknown;
  skipAuth?: boolean;
};

async function request<T>(
  path: string,
  opts: RequestOptions = {},
): Promise<{ data: T; meta?: PageMeta }> {
  const { body, skipAuth, headers, ...rest } = opts;
  const send = async (token: string | null): Promise<Response> => {
    const h = new Headers(headers);
    const isForm = body instanceof FormData;
    if (!isForm) h.set("Content-Type", "application/json");
    if (token) h.set("Authorization", `Bearer ${token}`);

    return fetch(`${API_BASE}${path}`, {
      ...rest,
      headers: h,
      credentials: "include",
      body:
        body === undefined ? undefined : isForm ? body : JSON.stringify(body),
    });
  };

  let token = skipAuth ? null : await getValidToken();
  let res = await send(token);

  if (res.status === 401 && !skipAuth) {
    token = await refreshAccessToken();
    if (token) res = await send(token);
  }

  const raw = await res.text();

  let parsed: Envelope<T> = {};
  if (raw) {
    try {
      parsed = JSON.parse(raw) as Envelope<T>;
    } catch {
      if (res.ok) {
        throw new ApiError(
          res.status,
          "BAD_RESPONSE",
          "Máy chủ trả về dữ liệu không đọc được",
        );
      }
    }
  }

  if (!res.ok) {
    throw new ApiError(
      res.status,
      parsed.error?.code ?? gatewayCode(res.status),
      parsed.error?.message ?? gatewayMessage(res.status),
      parsed.error?.details,
      parsed.error?.request_id,
    );
  }

  return { data: parsed.data as T, meta: parsed.meta as PageMeta | undefined };
}

export async function downloadFile(
  path: string,
  fallbackName: string,
): Promise<void> {
  const send = (token: string | null) =>
    fetch(`${API_BASE}${path}`, {
      headers: token ? { Authorization: `Bearer ${token}` } : {},
      credentials: "include",
    });

  let res = await send(await getValidToken());
  if (res.status === 401) {
    const token = await refreshAccessToken();
    if (token) res = await send(token);
  }

  if (!res.ok) {
    let parsed: Envelope<unknown> = {};
    try {
      parsed = (await res.json()) as Envelope<unknown>;
    } catch {}
    throw new ApiError(
      res.status,
      parsed.error?.code ?? gatewayCode(res.status),
      parsed.error?.message ?? gatewayMessage(res.status),
    );
  }

  const cd = res.headers.get("Content-Disposition") ?? "";
  const star = /filename\*=UTF-8''([^;]+)/i.exec(cd)?.[1];
  const plain = /filename="([^"]+)"/i.exec(cd)?.[1];
  const name = star ? decodeURIComponent(star) : (plain ?? fallbackName);

  const url = URL.createObjectURL(await res.blob());
  const a = document.createElement("a");
  a.href = url;
  a.download = name;
  a.click();
  URL.revokeObjectURL(url);
}

export const api = {
  get: <T>(path: string, opts?: RequestOptions) =>
    request<T>(path, { ...opts, method: "GET" }),

  post: <T>(path: string, body?: unknown, opts?: RequestOptions) =>
    request<T>(path, { ...opts, method: "POST", body }),

  put: <T>(path: string, body?: unknown, opts?: RequestOptions) =>
    request<T>(path, { ...opts, method: "PUT", body }),

  patch: <T>(path: string, body?: unknown, opts?: RequestOptions) =>
    request<T>(path, { ...opts, method: "PATCH", body }),

  delete: <T>(path: string, opts?: RequestOptions) =>
    request<T>(path, { ...opts, method: "DELETE" }),
};

function gatewayCode(status: number): string {
  switch (status) {
    case 429:
      return "TOO_MANY_REQUEST";
    case 413:
      return "PAYLOAD_TOO_LARGE";
    case 502:
    case 503:
    case 504:
      return "SERVICE_UNAVALIABLE";
    default:
      return "UNKNOW";
  }
}

function gatewayMessage(status: number): string {
  switch (status) {
    case 429:
      return "Bạn thao tác quá nhanh. Chờ khoảng một phút rồi thử lại.";
    case 413:
      return "Tệp hoặc dữ liệu gửi lại quá lớn.";
    case 502:
    case 503:
    case 504:
      return "Máy chủ đang bận hoặc vừa khởi động lại. Thử lại sau một lát.";
    default:
      return "Đã có lỗi xảy ra";
  }
}
