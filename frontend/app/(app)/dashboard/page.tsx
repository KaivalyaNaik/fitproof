import Link from "next/link";
import { serverFetch } from "@/lib/server-api";
import type { ChallengeListItem, UserStats, User } from "@/lib/types";
import { ChallengeCard } from "@/components/challenges/ChallengeCard";
import { JoinChallengeButton } from "@/components/challenges/JoinChallengeButton";
import { Button } from "@/components/ui/Button";
import { EmailVerificationBanner } from "@/components/auth/EmailVerificationBanner";
import { formatPoints } from "@/lib/utils";

export default async function DashboardPage() {
  const [challenges, stats, user] = await Promise.all([
    serverFetch<ChallengeListItem[]>("/challenges"),
    serverFetch<UserStats>("/me/stats"),
    serverFetch<User>("/me"),
  ]);

  return (
    <main className="max-w-5xl mx-auto px-4 sm:px-6 py-6 sm:py-10">
      {!user.email_verified && (
        <EmailVerificationBanner userEmail={user.email} />
      )}

      {/* Stats */}
      <div className="grid grid-cols-3 gap-2 sm:gap-3 mb-8 sm:mb-12">
        <StatTile label="Challenges" value={String(stats.challenges_joined)} />
        <StatTile label="Points" value={formatPoints(stats.total_points)} />
        <StatTile
          label="Submissions"
          value={`${stats.total_submissions}/${stats.total_submissions + stats.missed_submissions}`}
        />
      </div>

      {/* Header */}
      <div className="flex flex-wrap items-center justify-between gap-3 mb-6">
        <h2 className="text-[11px] font-semibold text-[var(--text-muted)] uppercase tracking-widest">
          Your Challenges
        </h2>
        <div className="flex gap-2">
          <JoinChallengeButton />
          <Link href="/challenges/new">
            <Button size="sm">New challenge</Button>
          </Link>
        </div>
      </div>

      {/* Challenge grid */}
      {challenges.length === 0 ? (
        <div className="text-center py-24 text-[var(--text-muted)]">
          <p className="text-sm font-medium mb-1">No challenges yet</p>
          <p className="text-xs text-[var(--text-dim)]">Create one or join with an invite code.</p>
        </div>
      ) : (
        <div className="grid gap-3 sm:grid-cols-2">
          {challenges.map((c) => (
            <ChallengeCard key={c.id} challenge={c} />
          ))}
        </div>
      )}
    </main>
  );
}

function StatTile({ label, value }: { label: string; value: string }) {
  return (
    <div className="bg-[var(--surface)] border border-[var(--border)] rounded-2xl p-4 sm:p-5">
      <p className="text-[10px] font-semibold text-[var(--text-muted)] uppercase tracking-widest mb-2">
        {label}
      </p>
      <p className="text-xl font-semibold tabular-nums truncate font-mono-nums text-[var(--text)]">
        {value}
      </p>
    </div>
  );
}
