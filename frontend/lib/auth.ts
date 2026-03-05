import { serverFetch } from "@/lib/server-api";
import type { User } from "@/lib/types";

export async function getCurrentUser(): Promise<User | null> {
  try {
    return await serverFetch<User>("/me");
  } catch {
    return null;
  }
}
