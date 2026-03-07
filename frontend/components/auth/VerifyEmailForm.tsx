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
  const [sent, setSent] = useState<boolean | "failed">(false);
  const inputRef = useRef<HTMLInputElement>(null);

  // Send code on mount (already sent at register, but handles direct navigation too)
  useEffect(() => {
    clientFetch("/auth/verify/send", { method: "POST" })
      .then(() => {
        setSent(true);
        startCooldown();
      })
      .catch((err) => {
        if (err instanceof ApiResponseError && err.status === 409) {
          // Already verified — go to dashboard
          router.push("/dashboard");
        } else {
          // Failed to send (SMTP not configured, etc.) — stay on page, let user retry
          setSent("failed");
        }
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
        <h1 className="text-xl font-semibold text-[var(--text)] mb-1">Verify your email</h1>
        <p className="text-sm text-[var(--text-muted)]">
          {sent === true
            ? "We sent a 6-digit code to your email address."
            : sent === "failed"
            ? "Couldn't send a code. Use the button below to try again."
            : "Sending code…"}
        </p>
      </div>

      <form onSubmit={handleSubmit} className="flex flex-col gap-4">
        <div className="flex flex-col gap-1.5">
          <label
            htmlFor="code"
            className="text-[10px] font-semibold text-[var(--text-muted)] uppercase tracking-widest"
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
            className="block w-full rounded-xl border border-[var(--border)] bg-[var(--surface)] px-3.5 py-3 text-2xl font-mono tracking-[0.5em] text-center text-[var(--text)] placeholder-[var(--text-dim)] focus:outline-none focus:border-[var(--accent)] transition-colors"
          />
        </div>

        {error && (
          <p className="text-sm text-[var(--danger)] bg-[var(--danger-dim)] border border-[var(--danger)]/20 rounded-xl px-3.5 py-2.5">
            {error}
          </p>
        )}

        <Button type="submit" loading={loading} disabled={code.length < 6} className="w-full">
          Verify email
        </Button>
      </form>

      <div className="flex items-center justify-between mt-5">
        <p className="text-xs text-[var(--text-dim)]">Didn&apos;t receive it?</p>
        {cooldown > 0 ? (
          <span className="text-xs text-[var(--text-dim)] tabular-nums">Resend in {cooldown}s</span>
        ) : (
          <button
            type="button"
            onClick={handleResend}
            className="text-xs text-[var(--text-muted)] hover:text-[var(--text)] font-medium underline underline-offset-2 transition-colors"
          >
            Resend code
          </button>
        )}
      </div>
    </>
  );
}
