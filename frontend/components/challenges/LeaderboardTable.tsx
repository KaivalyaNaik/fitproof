import type { LeaderboardEntry } from "@/lib/types";
import { formatPoints, formatDate } from "@/lib/utils";

export function LeaderboardTable({ entries }: { entries: LeaderboardEntry[] }) {
  if (entries.length === 0) {
    return (
      <div className="text-center py-14 text-[var(--text-muted)] text-sm">
        No submissions yet — be the first!
      </div>
    );
  }

  return (
    <div className="overflow-x-auto">
      <table className="w-full text-sm">
        <thead>
          <tr className="border-b border-[var(--border)]">
            <th className="px-5 py-3.5 text-left text-[10px] font-semibold text-[var(--text-muted)] uppercase tracking-widest w-14">
              #
            </th>
            <th className="px-5 py-3.5 text-left text-[10px] font-semibold text-[var(--text-muted)] uppercase tracking-widest">
              Name
            </th>
            <th className="px-5 py-3.5 text-right text-[10px] font-semibold text-[var(--text-muted)] uppercase tracking-widest">
              Points
            </th>
            <th className="px-5 py-3.5 text-right text-[10px] font-semibold text-[var(--text-muted)] uppercase tracking-widest hidden sm:table-cell">
              Last sub
            </th>
          </tr>
        </thead>
        <tbody>
          {entries.map((entry) => (
            <tr
              key={entry.user_id}
              className={[
                "border-b border-[var(--border-subtle)] last:border-0 transition-colors",
                entry.rank === 1
                  ? "bg-[var(--accent-dim)]"
                  : "hover:bg-[var(--surface-raised)]",
              ].join(" ")}
            >
              <td className="px-5 py-3.5">
                <span
                  className={[
                    "text-sm tabular-nums font-semibold font-mono-nums",
                    entry.rank === 1
                      ? "text-[var(--accent)]"
                      : "text-[var(--text-dim)]",
                  ].join(" ")}
                >
                  {entry.rank}
                </span>
              </td>
              <td className="px-5 py-3.5 font-medium text-[var(--text)] text-sm">
                {entry.display_name}
                {entry.rank === 1 && (
                  <span className="ml-2 text-sm">🏆</span>
                )}
              </td>
              <td className="px-5 py-3.5 text-right font-mono text-[var(--text)] tabular-nums text-xs font-mono-nums">
                {formatPoints(entry.total_points)}
              </td>
              <td className="px-5 py-3.5 text-right text-[var(--text-muted)] text-xs hidden sm:table-cell tabular-nums font-mono-nums">
                {entry.last_submission_date
                  ? formatDate(entry.last_submission_date)
                  : "—"}
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
