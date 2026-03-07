import Link from "next/link";
import type { ChallengeListItem } from "@/lib/types";
import { Badge } from "@/components/ui/Badge";
import { formatDate, statusLabel } from "@/lib/utils";

export function ChallengeCard({ challenge }: { challenge: ChallengeListItem }) {
  const { id, name, description, status, start_date, end_date, invite_code, membership } =
    challenge;

  return (
    <Link href={`/challenges/${id}`}>
      <div className="bg-[var(--surface)] border border-[var(--border)] rounded-2xl p-5 hover:border-[var(--text-dim)] hover:bg-[var(--surface-raised)] transition-all duration-200 cursor-pointer h-full group">
        <div className="flex items-start justify-between gap-3 mb-2">
          <h3 className="font-semibold text-[var(--text)] text-sm leading-snug group-hover:text-[var(--accent)] transition-colors">
            {name}
          </h3>
          <div className="flex gap-1.5 shrink-0">
            <Badge
              variant={status as "active" | "completed" | "cancelled" | "draft"}
              label={statusLabel(status)}
            />
            <Badge
              variant={membership.role as "host" | "cohost" | "participant"}
              label={membership.role.charAt(0).toUpperCase() + membership.role.slice(1)}
            />
          </div>
        </div>

        {description && (
          <p className="text-xs text-[var(--text-muted)] mb-4 line-clamp-2 leading-relaxed">
            {description}
          </p>
        )}

        <div className="flex items-center justify-between text-[11px] text-[var(--text-dim)] mt-auto pt-2">
          <span className="tabular-nums font-mono-nums">
            {formatDate(start_date)} – {formatDate(end_date)}
          </span>
          {(membership.role === "host" || membership.role === "cohost") && (
            <span className="font-mono bg-[var(--surface-raised)] border border-[var(--border)] px-1.5 py-0.5 rounded-md text-[var(--text-muted)] text-[10px]">
              {invite_code}
            </span>
          )}
        </div>
      </div>
    </Link>
  );
}
