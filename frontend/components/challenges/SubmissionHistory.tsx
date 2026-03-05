"use client";

import { useEffect, useState } from "react";
import { clientFetch } from "@/lib/client-api";
import type { SubmissionHistoryItem } from "@/lib/types";
import { formatDate, formatTimestamp, formatPoints } from "@/lib/utils";

export function SubmissionHistory({ challengeId }: { challengeId: string }) {
  const [history, setHistory] = useState<SubmissionHistoryItem[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    clientFetch<SubmissionHistoryItem[]>(`/challenges/${challengeId}/submissions`)
      .then(setHistory)
      .finally(() => setLoading(false));
  }, [challengeId]);

  if (loading) {
    return (
      <div className="py-14 text-center text-zinc-400 text-sm">Loading…</div>
    );
  }

  if (history.length === 0) {
    return (
      <div className="py-14 text-center text-zinc-400 text-sm">
        No submissions yet.
      </div>
    );
  }

  return (
    <div className="flex flex-col gap-2">
      {history.map((item) => (
        <div
          key={item.id}
          className="bg-white rounded-2xl ring-1 ring-zinc-100 overflow-hidden"
        >
          <div className="flex items-center justify-between px-5 py-4">
            <div>
              <p className="text-sm font-semibold text-zinc-900">
                {formatDate(item.date)}
              </p>
              <p className="text-[11px] text-zinc-400 mt-0.5">
                {item.submission_type === "missed"
                  ? "Missed"
                  : `Submitted at ${formatTimestamp(item.submitted_at)}`}
              </p>
            </div>
            <div className="text-right flex items-center gap-2">
              {item.submission_type === "missed" && (
                <span className="text-[11px] bg-red-50 text-red-600 ring-1 ring-red-100 px-2 py-0.5 rounded-md font-medium">
                  Missed
                </span>
              )}
              <p className="font-mono font-semibold text-sm tabular-nums text-zinc-900">
                {formatPoints(item.total_points_earned)}
              </p>
            </div>
          </div>

          {item.metrics.length > 0 && (
            <div className="border-t border-zinc-50 px-5 py-3 flex flex-col gap-2.5">
              {item.metrics.map((mv) => (
                <div
                  key={mv.metric_id}
                  className="flex items-center justify-between"
                >
                  <div className="flex items-center gap-2">
                    <span
                      className={`w-1.5 h-1.5 rounded-full shrink-0 ${
                        mv.passed ? "bg-emerald-500" : "bg-red-400"
                      }`}
                    />
                    <span className="text-xs text-zinc-600">{mv.metric_name}</span>
                  </div>
                  <div className="flex items-center gap-3">
                    <span className="font-mono text-[11px] text-zinc-400 tabular-nums">
                      {mv.value}
                    </span>
                    <span
                      className={[
                        "text-[11px] font-medium tabular-nums",
                        mv.passed ? "text-emerald-600" : "text-zinc-300",
                      ].join(" ")}
                    >
                      {mv.passed ? `+${formatPoints(mv.points_awarded)}` : "0 pts"}
                    </span>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      ))}
    </div>
  );
}
