// IMPORTANT: this route must NOT live under /api/* — next.config.ts proxies the
// entire /api/* namespace to the Go backend, which has no /log route, so a
// /api/log handler here would silently 404 in production. Do not move it.
// Also do NOT rename to /_log: Next.js opts _-prefixed folders out of routing.

import { NextRequest, NextResponse } from "next/server";

const LOKI_URL = process.env.GRAFANA_LOKI_URL;
const LOKI_USER = process.env.GRAFANA_LOKI_USER;
const LOKI_TOKEN = process.env.GRAFANA_LOKI_TOKEN;

const ALLOWED_ORIGINS = new Set([
  "http://localhost:3000",
  "https://fitproof-six.vercel.app",
]);

function originAllowed(origin: string): boolean {
  if (!origin) return false;
  if (ALLOWED_ORIGINS.has(origin)) return true;
  if (process.env.VERCEL_URL && origin === `https://${process.env.VERCEL_URL}`) {
    return true;
  }
  return false;
}

type ClientLogPayload = {
  level?: unknown;
  msg?: unknown;
  context?: Record<string, unknown>;
};

export async function POST(req: NextRequest) {
  if (!originAllowed(req.headers.get("origin") ?? "")) {
    return NextResponse.json({ error: "forbidden" }, { status: 403 });
  }

  if (!LOKI_URL || !LOKI_USER || !LOKI_TOKEN) {
    return NextResponse.json({ ok: true, sink: "noop" });
  }

  let payload: ClientLogPayload;
  try {
    payload = (await req.json()) as ClientLogPayload;
  } catch {
    return NextResponse.json({ error: "invalid json" }, { status: 400 });
  }

  if (
    typeof payload.level !== "string" ||
    !["info", "warn", "error"].includes(payload.level) ||
    typeof payload.msg !== "string" ||
    payload.msg.length === 0 ||
    payload.msg.length > 4000
  ) {
    return NextResponse.json({ error: "invalid payload" }, { status: 400 });
  }

  const line = JSON.stringify({
    level: payload.level.toUpperCase(),
    msg: payload.msg,
    ...(payload.context ?? {}),
    user_agent: req.headers.get("user-agent") ?? "",
    ip: req.headers.get("x-forwarded-for")?.split(",")[0]?.trim() ?? "",
  });

  const body = {
    streams: [
      {
        stream: {
          service: "fitproof-web",
          env: process.env.VERCEL_ENV ?? "development",
          source: "client",
        },
        values: [[(Date.now() * 1_000_000).toString(), line]],
      },
    ],
  };

  try {
    await fetch(`${LOKI_URL}/loki/api/v1/push`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        Authorization: `Basic ${btoa(`${LOKI_USER}:${LOKI_TOKEN}`)}`,
      },
      body: JSON.stringify(body),
      signal: AbortSignal.timeout(3000),
    });
  } catch {
    // Never fail the client — logging must not block UX.
  }

  return NextResponse.json({ ok: true });
}
