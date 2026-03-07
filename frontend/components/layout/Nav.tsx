import Link from "next/link";
import { getCurrentUser } from "@/lib/auth";
import { NavClient } from "./NavClient";

export async function Nav() {
  const user = await getCurrentUser();

  return (
    <nav className="bg-[var(--surface)] border-b border-[var(--border)] sticky top-0 z-40 backdrop-blur-sm">
      <div className="max-w-5xl mx-auto px-4 sm:px-6 h-[72px] flex items-center justify-between">
        <Link href="/dashboard" className="flex items-center hover:opacity-80 transition-opacity">
          {/* eslint-disable-next-line @next/next/no-img-element */}
          <img
            src="/logo-full.png"
            alt="FitProof"
            style={{ height: 68, width: "auto", display: "block" }}
          />
        </Link>
        {user && <NavClient displayName={user.display_name} />}
      </div>
    </nav>
  );
}
