import Link from "next/link";
import { getCurrentUser } from "@/lib/auth";
import { NavClient } from "./NavClient";

export async function Nav() {
  const user = await getCurrentUser();

  return (
    <nav className="bg-white border-b border-zinc-100 sticky top-0 z-40">
      <div className="max-w-5xl mx-auto px-4 sm:px-6 h-14 flex items-center justify-between">
        <Link
          href="/dashboard"
          className="text-zinc-900 font-semibold text-sm tracking-tight hover:text-zinc-700 transition-colors"
        >
          FitProof
        </Link>
        {user && <NavClient displayName={user.display_name} />}
      </div>
    </nav>
  );
}
