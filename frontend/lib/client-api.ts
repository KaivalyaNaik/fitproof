// Used in Client Components. Calls /api/* (proxied to backend) with
// credentials: "include" so the browser sends its cookies automatically.
// On 401: attempts POST /api/auth/refresh; retries once; if still failing redirects to /login.

export class ApiResponseError extends Error {
  status: number;
  constructor(status: number, message: string) {
    super(message);
    this.status = status;
    this.name = "ApiResponseError";
  }
}

let refreshPromise: Promise<boolean> | null = null;

async function refreshToken(): Promise<boolean> {
  if (refreshPromise) return refreshPromise;
  refreshPromise = fetch("/api/auth/refresh", {
    method: "POST",
    credentials: "include",
  })
    .then((r) => r.ok)
    .finally(() => {
      refreshPromise = null;
    });
  return refreshPromise;
}

function buildHeaders(options?: RequestInit): HeadersInit {
  if (options?.body instanceof FormData) return options?.headers ?? {};
  return { "Content-Type": "application/json", ...(options?.headers ?? {}) };
}

export async function clientFetch<T>(
  path: string,
  options?: RequestInit
): Promise<T> {
  const res = await fetch(`/api${path}`, {
    ...options,
    credentials: "include",
    headers: buildHeaders(options),
  });

  if (res.status === 401) {
    const refreshed = await refreshToken();
    if (!refreshed) {
      window.location.href = "/login";
      throw new ApiResponseError(401, "Session expired");
    }
    const retry = await fetch(`/api${path}`, {
      ...options,
      credentials: "include",
      headers: buildHeaders(options),
    });
    if (retry.status === 401) {
      window.location.href = "/login";
      throw new ApiResponseError(401, "Session expired");
    }
    if (!retry.ok) {
      const body = await retry
        .json()
        .catch(() => ({ error: "Unknown error" }));
      throw new ApiResponseError(
        retry.status,
        (body as { error: string }).error ?? "Unknown error"
      );
    }
    if (retry.status === 204) return undefined as T;
    return retry.json() as Promise<T>;
  }

  if (!res.ok) {
    const body = await res.json().catch(() => ({ error: "Unknown error" }));
    throw new ApiResponseError(
      res.status,
      (body as { error: string }).error ?? "Unknown error"
    );
  }

  if (res.status === 204) return undefined as T;
  return res.json() as Promise<T>;
}
