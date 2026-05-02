import { NextResponse } from "next/server";
import type { NextRequest } from "next/server";

// /log is the server-side Loki relay — publicly reachable by design (Origin
// check inside the route handler is the real gate). Must NOT be auth-gated or
// browser logs from logged-out users (the login page itself, error reporter on
// public pages) would 307 to /login and never reach Loki.
const PUBLIC_PATHS = ["/login", "/register", "/verify-email", "/log"];

export function proxy(request: NextRequest) {
  const { pathname } = request.nextUrl;

  if (pathname === "/" || PUBLIC_PATHS.some((p) => pathname.startsWith(p))) {
    return NextResponse.next();
  }

  const token = request.cookies.get("access_token");
  if (!token) {
    const url = new URL("/login", request.url);
    url.searchParams.set("from", pathname);
    return NextResponse.redirect(url);
  }

  return NextResponse.next();
}

export const config = {
  matcher: [
    "/((?!_next/static|_next/image|favicon.ico|api/|.*\\.(?:png|jpg|jpeg|gif|svg|ico|webp|mp4|mov|avi|webm|woff2?|ttf|otf)$).*)",
  ],
};
