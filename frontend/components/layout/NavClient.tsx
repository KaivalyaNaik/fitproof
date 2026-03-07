"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { clientFetch } from "@/lib/client-api";
import { Button } from "@/components/ui/Button";

export function NavClient({ displayName }: { displayName: string }) {
  const router = useRouter();
  const [loading, setLoading] = useState(false);

  async function handleLogout() {
    setLoading(true);
    await clientFetch("/auth/logout", { method: "POST" }).catch(() => {});
    router.push("/login");
  }

  return (
    <div className="flex items-center gap-3">
      <span className="text-xs text-[var(--text-muted)] font-medium hidden sm:block">
        {displayName}
      </span>
      <Button variant="ghost" size="sm" onClick={handleLogout} loading={loading}>
        Sign out
      </Button>
    </div>
  );
}
