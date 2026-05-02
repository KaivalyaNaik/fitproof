"use client";

import { useEffect } from "react";

import { log } from "@/lib/logger";

export function ErrorReporter() {
  useEffect(() => {
    const onError = (e: ErrorEvent) =>
      log.error("window.onerror", {
        message: e.message,
        filename: e.filename,
        lineno: e.lineno,
        colno: e.colno,
        stack: e.error?.stack,
      });
    const onRejection = (e: PromiseRejectionEvent) =>
      log.error("unhandledrejection", {
        reason: String(e.reason),
        stack: (e.reason as Error)?.stack,
      });

    window.addEventListener("error", onError);
    window.addEventListener("unhandledrejection", onRejection);
    return () => {
      window.removeEventListener("error", onError);
      window.removeEventListener("unhandledrejection", onRejection);
    };
  }, []);

  return null;
}
