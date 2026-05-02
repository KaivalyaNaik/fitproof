import { cookies } from "next/headers";
import { redirect } from "next/navigation";

import { log } from "./logger";

export class ApiResponseError extends Error {
  status: number;
  constructor(status: number, message: string) {
    super(message);
    this.status = status;
    this.name = "ApiResponseError";
  }
}

// ─── Server fetch ────────────────────────────────────────────────────────────
// Used only in Server Components. Forwards cookies from next/headers to the
// backend. On 401, redirects to /login (cannot refresh from a Server Component).

export async function serverFetch<T>(
  path: string,
  options?: RequestInit
): Promise<T> {
  const cookieStore = await cookies();
  const cookieHeader = cookieStore
    .getAll()
    .map((c) => `${c.name}=${c.value}`)
    .join("; ");

  const res = await fetch(`${process.env.BACKEND_URL ?? "http://localhost:8080"}${path}`, {
    ...options,
    headers: {
      "Content-Type": "application/json",
      Cookie: cookieHeader,
      ...(options?.headers ?? {}),
    },
    cache: "no-store",
  });

  if (res.status === 401) {
    redirect("/login");
  }

  if (!res.ok) {
    const body = await res.json().catch(() => ({ error: "Unknown error" }));
    const message = (body as { error: string }).error ?? "Unknown error";
    log.error("api error", { source: "serverFetch", status: res.status, path, message });
    throw new ApiResponseError(res.status, message);
  }

  if (res.status === 204) return undefined as T;
  return res.json() as Promise<T>;
}

// ─── Client fetch ────────────────────────────────────────────────────────────
// Used in Client Components. Calls /api/* (proxied to backend) with
// credentials: "include" so the browser sends its cookies automatically.
// On 401: attempts POST /api/auth/refresh; retries once; if still 401 redirects to /login.

let isRefreshing = false;

export async function clientFetch<T>(
  path: string,
  options?: RequestInit
): Promise<T> {
  const res = await fetch(`/api${path}`, {
    ...options,
    credentials: "include",
    headers: {
      "Content-Type": "application/json",
      ...(options?.headers ?? {}),
    },
  });

  if (res.status === 401 && !isRefreshing) {
    isRefreshing = true;
    try {
      const refreshed = await fetch("/api/auth/refresh", {
        method: "POST",
        credentials: "include",
      });

      if (refreshed.ok) {
        isRefreshing = false;
        const retry = await fetch(`/api${path}`, {
          ...options,
          credentials: "include",
          headers: {
            "Content-Type": "application/json",
            ...(options?.headers ?? {}),
          },
        });
        if (retry.status === 401) {
          log.error("api 401 after retry", { path });
          window.location.href = "/login";
          throw new ApiResponseError(401, "Session expired");
        }
        if (!retry.ok) {
          const body = await retry.json().catch(() => ({ error: "Unknown error" }));
          const message = (body as { error: string }).error ?? "Unknown error";
          log.error("api error", { source: "clientFetch.retry", status: retry.status, path, message });
          throw new ApiResponseError(retry.status, message);
        }
        if (retry.status === 204) return undefined as T;
        return retry.json() as Promise<T>;
      } else {
        isRefreshing = false;
        log.error("refresh failed", { path, refreshStatus: refreshed.status });
        window.location.href = "/login";
        throw new ApiResponseError(401, "Session expired");
      }
    } catch (e) {
      isRefreshing = false;
      throw e;
    }
  }

  if (!res.ok) {
    const body = await res.json().catch(() => ({ error: "Unknown error" }));
    const message = (body as { error: string }).error ?? "Unknown error";
    log.error("api error", { source: "clientFetch", status: res.status, path, message });
    throw new ApiResponseError(res.status, message);
  }

  if (res.status === 204) return undefined as T;
  return res.json() as Promise<T>;
}
