"use client";

import { useState, useEffect, FormEvent, useRef } from "react";
import { useRouter } from "next/navigation";
import { clientFetch, ApiResponseError } from "@/lib/client-api";
import { Button } from "@/components/ui/Button";

const RESEND_COOLDOWN = 60;

export function VerifyEmailForm() {
  const router = useRouter();
  const [code, setCode] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [cooldown, setCooldown] = useState(0);
  const [sent, setSent] = useState(false);
  const inputRef = useRef<HTMLInputElement>(null);

  // Send code on mount (already sent at register, but handles direct navigation too)
  useEffect(() => {
    clientFetch("/auth/verify/send", { method: "POST" })
      .then(() => {
        setSent(true);
        startCooldown();
      })
      .catch(() => {
        // 409 = already verified → go to dashboard
        router.push("/dashboard");
      });
    inputRef.current?.focus();
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  function startCooldown() {
    setCooldown(RESEND_COOLDOWN);
  }

  useEffect(() => {
    if (cooldown <= 0) return;
    const t = setInterval(() => setCooldown((c) => c - 1), 1000);
    return () => clearInterval(t);
  }, [cooldown]);

  async function handleResend() {
    setError(null);
    try {
      await clientFetch("/auth/verify/send", { method: "POST" });
      setSent(true);
      startCooldown();
    } catch (err) {
      if (err instanceof ApiResponseError && err.status === 409) {
        router.push("/dashboard");
      } else {
        setError("Failed to resend. Please try again.");
      }
    }
  }

  async function handleSubmit(e: FormEvent) {
    e.preventDefault();
    setError(null);
    setLoading(true);
    try {
      await clientFetch("/auth/verify", {
        method: "POST",
        body: JSON.stringify({ code: code.trim() }),
      });
      router.refresh();
      router.push("/dashboard");
    } catch (err) {
      if (err instanceof ApiResponseError && err.status === 422) {
        setError("Invalid or expired code. Check your inbox or resend.");
      } else if (err instanceof ApiResponseError && err.status === 409) {
        router.push("/dashboard");
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
        <h1 className="text-xl font-semibold text-zinc-900 mb-1">Verify your email</h1>
        <p className="text-sm text-zinc-400">
          {sent
            ? "We sent a 6-digit code to your email address."
            : "Sending code…"}
        </p>
      </div>

      <form onSubmit={handleSubmit} className="flex flex-col gap-4">
        <div className="flex flex-col gap-1.5">
          <label
            htmlFor="code"
            className="text-[11px] font-semibold text-zinc-400 uppercase tracking-widest"
          >
            Verification code
          </label>
          <input
            ref={inputRef}
            id="code"
            type="text"
            inputMode="numeric"
            pattern="[0-9]*"
            maxLength={6}
            placeholder="000000"
            value={code}
            onChange={(e) => setCode(e.target.value.replace(/\D/g, "").slice(0, 6))}
            required
            className="block w-full rounded-xl border border-zinc-200 px-3.5 py-3 text-2xl font-mono tracking-[0.5em] text-center text-zinc-900 placeholder-zinc-300 focus:outline-none focus:border-zinc-900 transition-colors"
          />
        </div>

        {error && (
          <p className="text-sm text-red-600 bg-red-50 rounded-xl px-3.5 py-2.5">
            {error}
          </p>
        )}

        <Button type="submit" loading={loading} disabled={code.length < 6} className="w-full">
          Verify email
        </Button>
      </form>

      <div className="flex items-center justify-between mt-5">
        <p className="text-xs text-zinc-400">Didn&apos;t receive it?</p>
        {cooldown > 0 ? (
          <span className="text-xs text-zinc-400 tabular-nums">Resend in {cooldown}s</span>
        ) : (
          <button
            type="button"
            onClick={handleResend}
            className="text-xs text-zinc-700 hover:text-zinc-900 font-medium underline underline-offset-2 transition-colors"
          >
            Resend code
          </button>
        )}
      </div>
    </>
  );
}
