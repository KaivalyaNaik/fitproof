"use client";

import { useState, FormEvent } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import { clientFetch, ApiResponseError } from "@/lib/client-api";
import { User } from "@/lib/types";
import { Button } from "@/components/ui/Button";
import { Input } from "@/components/ui/Input";
import { GoogleButton } from "./GoogleButton";

export function RegisterForm() {
  const router = useRouter();
  const [displayName, setDisplayName] = useState("");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  async function handleSubmit(e: FormEvent) {
    e.preventDefault();
    setError(null);
    setLoading(true);
    try {
      await clientFetch<User>("/auth/register", {
        method: "POST",
        body: JSON.stringify({ email, password, display_name: displayName }),
      });
      router.push("/verify-email");
    } catch (err) {
      if (err instanceof ApiResponseError) {
        setError(
          err.status === 409
            ? "Email already in use."
            : err.message
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
        <h1 className="text-xl font-semibold text-zinc-900 mb-1">Create account</h1>
        <p className="text-sm text-zinc-400">Join FitProof today</p>
      </div>

      <GoogleButton />

      <div className="flex items-center gap-3 my-5">
        <div className="flex-1 h-px bg-zinc-100" />
        <span className="text-xs text-zinc-400">or</span>
        <div className="flex-1 h-px bg-zinc-100" />
      </div>

      <form onSubmit={handleSubmit} className="flex flex-col gap-4">
        <Input
          id="display_name"
          label="Display name"
          type="text"
          autoComplete="name"
          value={displayName}
          onChange={(e) => setDisplayName(e.target.value)}
          required
        />
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
          autoComplete="new-password"
          value={password}
          onChange={(e) => setPassword(e.target.value)}
          required
          minLength={8}
        />

        {error && (
          <p className="text-sm text-red-600 bg-red-50 rounded-xl px-3.5 py-2.5">
            {error}
          </p>
        )}

        <Button type="submit" loading={loading} className="w-full mt-1">
          Create account
        </Button>
      </form>

      <p className="text-xs text-zinc-400 text-center mt-6">
        Already have an account?{" "}
        <Link
          href="/login"
          className="text-zinc-700 hover:text-zinc-900 font-medium underline underline-offset-2 transition-colors"
        >
          Sign in
        </Link>
      </p>
    </>
  );
}
