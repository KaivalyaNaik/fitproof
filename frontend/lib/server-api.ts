import { cookies } from "next/headers";
import { redirect } from "next/navigation";
import { ApiResponseError } from "./client-api";

// Used only in Server Components. Forwards cookies from next/headers.
// On 401, redirects to /login (cannot refresh from a Server Component).
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
    throw new ApiResponseError(
      res.status,
      (body as { error: string }).error ?? "Unknown error"
    );
  }

  if (res.status === 204) return undefined as T;
  return res.json() as Promise<T>;
}
