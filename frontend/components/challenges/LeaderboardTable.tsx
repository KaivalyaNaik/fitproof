import type { LeaderboardEntry } from "@/lib/types";
import { formatPoints, formatFines, formatDate } from "@/lib/utils";

export function LeaderboardTable({ entries }: { entries: LeaderboardEntry[] }) {
  if (entries.length === 0) {
    return (
      <div className="text-center py-14 text-zinc-400 text-sm">
        No submissions yet — be the first!
      </div>
    );
  }

  return (
    <div className="overflow-x-auto">
      <table className="w-full text-sm">
        <thead>
          <tr className="border-b border-zinc-100">
            <th className="px-5 py-3.5 text-left text-[11px] font-semibold text-zinc-400 uppercase tracking-widest w-14">
              #
            </th>
            <th className="px-5 py-3.5 text-left text-[11px] font-semibold text-zinc-400 uppercase tracking-widest">
              Name
            </th>
            <th className="px-5 py-3.5 text-right text-[11px] font-semibold text-zinc-400 uppercase tracking-widest">
              Points
            </th>
            <th className="px-5 py-3.5 text-right text-[11px] font-semibold text-zinc-400 uppercase tracking-widest">
              Fines
            </th>
            <th className="px-5 py-3.5 text-right text-[11px] font-semibold text-zinc-400 uppercase tracking-widest hidden sm:table-cell">
              Last sub
            </th>
          </tr>
        </thead>
        <tbody>
          {entries.map((entry, i) => (
            <tr
              key={entry.user_id}
              className={[
                "border-b border-zinc-50 last:border-0 hover:bg-zinc-50/60 transition-colors",
                entry.rank === 1 ? "bg-indigo-50/40" : "",
              ].join(" ")}
            >
              <td className="px-5 py-3.5">
                <span
                  className={[
                    "text-sm tabular-nums font-semibold",
                    entry.rank === 1
                      ? "text-indigo-600"
                      : "text-zinc-400",
                  ].join(" ")}
                >
                  {entry.rank}
                </span>
              </td>
              <td className="px-5 py-3.5 font-medium text-zinc-800 text-sm">
                {entry.display_name}
                {entry.rank === 1 && (
                  <span className="ml-2 text-sm">🏆</span>
                )}
              </td>
              <td className="px-5 py-3.5 text-right font-mono text-zinc-700 tabular-nums text-xs">
                {formatPoints(entry.total_points)}
              </td>
              <td className="px-5 py-3.5 text-right font-mono tabular-nums text-xs">
                <span className={parseFloat(entry.total_fines) > 0 ? "text-red-500" : "text-zinc-300"}>
                  {parseFloat(entry.total_fines) > 0
                    ? formatFines(entry.total_fines)
                    : "—"}
                </span>
              </td>
              <td className="px-5 py-3.5 text-right text-zinc-400 text-xs hidden sm:table-cell tabular-nums">
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
