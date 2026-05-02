"use client";

import { useState, FormEvent } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import Link from "next/link";
import { clientFetch, ApiResponseError } from "@/lib/client-api";
import { User } from "@/lib/types";
import { Button } from "@/components/ui/Button";
import { Input } from "@/components/ui/Input";
import { GoogleButton } from "./GoogleButton";

export function LoginForm() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  const oauthError = searchParams.get("error") === "oauth_failed";

  function safeRedirect(raw: string | null): string {
    if (!raw) return "/dashboard";
    if (!raw.startsWith("/") || raw.startsWith("//")) return "/dashboard";
    if (raw.includes(":") || raw.includes("\\")) return "/dashboard";
    return raw;
  }

  async function handleSubmit(e: FormEvent) {
    e.preventDefault();
    setError(null);
    setLoading(true);
    try {
      await clientFetch<User>("/auth/login", {
        method: "POST",
        body: JSON.stringify({ email, password }),
      });
      router.push(safeRedirect(searchParams.get("from")));
    } catch (err) {
      if (err instanceof ApiResponseError) {
        setError(
          err.status === 401 ? "Invalid email or password." : err.message
        );
      } else {
        setError("Something went wrong. Please try again.");
      }
    } finally {
      setLoading(false);
    }
  }

  return (
    <>
      <div className="mb-7">
        <h1 className="text-xl font-semibold text-[var(--text)] mb-1">Sign in</h1>
        <p className="text-sm text-[var(--text-muted)]">Welcome back</p>
      </div>

      <GoogleButton />

      <div className="flex items-center gap-3 my-5">
        <div className="flex-1 h-px bg-[var(--border)]" />
        <span className="text-xs text-[var(--text-dim)]">or</span>
        <div className="flex-1 h-px bg-[var(--border)]" />
      </div>

      <form onSubmit={handleSubmit} className="flex flex-col gap-4">
        <Input
          id="email"
          label="Email"
          type="email"
          autoComplete="email"
          value={email}
          onChange={(e) => setEmail(e.target.value)}
          required
        />
        <Input
          id="password"
          label="Password"
          type="password"
          autoComplete="current-password"
          value={password}
          onChange={(e) => setPassword(e.target.value)}
          required
        />

        {(error || oauthError) && (
          <p className="text-sm text-[var(--danger)] bg-[var(--danger-dim)] border border-[var(--danger)]/20 rounded-xl px-3.5 py-2.5">
            {error ?? "Google sign-in failed. Please try again."}
          </p>
        )}

        <Button type="submit" loading={loading} className="w-full mt-1">
          Sign in
        </Button>
      </form>

      <p className="text-xs text-[var(--text-dim)] text-center mt-6">
        No account?{" "}
        <Link
          href="/register"
          className="text-[var(--text-muted)] hover:text-[var(--text)] font-medium underline underline-offset-2 transition-colors"
        >
          Create one
        </Link>
      </p>
    </>
  );
}
