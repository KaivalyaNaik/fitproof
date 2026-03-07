import Link from "next/link";
import Image from "next/image";
import { getCurrentUser } from "@/lib/auth";
import { NavClient } from "./NavClient";

export async function Nav() {
  const user = await getCurrentUser();

  return (
    <nav className="bg-white border-b border-zinc-100 sticky top-0 z-40">
      <div className="max-w-5xl mx-auto px-4 sm:px-6 h-14 flex items-center justify-between">
        <Link href="/dashboard" className="flex items-center hover:opacity-80 transition-opacity">
          <Image src="/logo-full.png" alt="FitProof" width={120} height={32} priority />
        </Link>
        {user && <NavClient displayName={user.display_name} />}
      </div>
    </nav>
  );
}
