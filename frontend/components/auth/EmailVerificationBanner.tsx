"use client";

import { useState, useEffect } from "react";
import Link from "next/link";
import { clientFetch } from "@/lib/client-api";

const SESSION_KEY = "email_verify_banner_dismissed";

export function EmailVerificationBanner({ userEmail }: { userEmail: string }) {
  const [dismissed, setDismissed] = useState<boolean | null>(null);
  const [resent, setResent] = useState(false);
  const [resending, setResending] = useState(false);

  useEffect(() => {
    setDismissed(sessionStorage.getItem(SESSION_KEY) === "1");
  }, []);

  if (dismissed !== false) return null;

  function handleDismiss() {
    sessionStorage.setItem(SESSION_KEY, "1");
    setDismissed(true);
  }

  async function handleResend() {
    setResending(true);
    try {
      await clientFetch("/auth/verify/send", { method: "POST" });
      setResent(true);
    } catch {
      // 409 = already verified, banner should disappear on next navigation
    } finally {
      setResending(false);
    }
  }

  return (
    <div className="bg-[var(--warning-dim)] border border-[var(--warning)]/20 rounded-2xl px-5 py-4 mb-6 flex items-center justify-between gap-4">
      <div className="flex items-center gap-3 min-w-0">
        <div className="w-7 h-7 rounded-full bg-[var(--warning)]/15 flex items-center justify-center shrink-0">
          <svg width="14" height="14" viewBox="0 0 14 14" fill="none">
            <path
              d="M7 1.75v5.25M7 10.25v.875"
              stroke="var(--warning)"
              strokeWidth="1.5"
              strokeLinecap="round"
            />
          </svg>
        </div>
        <div className="min-w-0">
          <p className="text-sm font-medium text-[var(--warning)] truncate">
            Verify your email address
          </p>
          <p className="text-xs text-[var(--warning)]/70 truncate">{userEmail}</p>
        </div>
      </div>

      <div className="flex items-center gap-2 shrink-0">
        {resent ? (
          <span className="text-xs text-[var(--warning)] font-medium">Code sent!</span>
        ) : (
          <button
            onClick={handleResend}
            disabled={resending}
            className="text-xs text-[var(--warning)]/80 hover:text-[var(--warning)] font-medium underline underline-offset-2 transition-colors disabled:opacity-50"
          >
            {resending ? "Sending…" : "Resend code"}
          </button>
        )}
        <Link
          href="/verify-email"
          className="text-xs bg-[var(--warning)] text-[var(--accent-fg)] px-3 py-1.5 rounded-lg font-semibold hover:brightness-110 transition-all"
        >
          Verify
        </Link>
        <button
          onClick={handleDismiss}
          className="text-[var(--warning)]/60 hover:text-[var(--warning)] transition-colors"
          aria-label="Dismiss"
        >
          <svg width="14" height="14" viewBox="0 0 14 14" fill="none">
            <path
              d="M1 1l12 12M13 1L1 13"
              stroke="currentColor"
              strokeWidth="1.5"
              strokeLinecap="round"
            />
          </svg>
        </button>
      </div>
    </div>
  );
}
