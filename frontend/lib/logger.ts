type Level = "info" | "warn" | "error";
type Context = Record<string, unknown>;

const isDev = process.env.NODE_ENV === "development";

const recent = new Map<string, number>();

function shouldSend(key: string): boolean {
  const now = Date.now();
  const last = recent.get(key);
  if (last && now - last < 5000) return false;
  recent.set(key, now);
  if (recent.size > 100) {
    let oldestKey = "";
    let oldestTs = Infinity;
    for (const [k, ts] of recent) {
      if (ts < oldestTs) {
        oldestKey = k;
        oldestTs = ts;
      }
    }
    if (oldestKey) recent.delete(oldestKey);
  }
  return true;
}

async function send(level: Level, msg: string, context?: Context) {
  if (isDev) {
    const fn =
      level === "error" ? console.error : level === "warn" ? console.warn : console.log;
    fn(`[${level}]`, msg, context ?? "");
    return;
  }

  if (typeof window === "undefined") {
    const fn = level === "error" ? console.error : console.log;
    fn(JSON.stringify({ level, msg, ...context }));
    return;
  }

  if (!shouldSend(`${level}:${msg}`)) return;

  try {
    await fetch("/log", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        level,
        msg,
        context: {
          ...context,
          url: window.location.href,
          ts: new Date().toISOString(),
        },
      }),
      keepalive: true,
    });
  } catch {
    // Never throw from logger.
  }
}

export const log = {
  info: (m: string, c?: Context) => void send("info", m, c),
  warn: (m: string, c?: Context) => void send("warn", m, c),
  error: (m: string, c?: Context) => void send("error", m, c),
};
