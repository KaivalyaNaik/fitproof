"use client";

import { useEffect, useState, useCallback } from "react";
import { clientFetch } from "@/lib/client-api";
import type { SubmissionHistoryItem } from "@/lib/types";
import { formatDate, formatTimestamp, formatPoints } from "@/lib/utils";

function isVideo(url: string) {
  return /\.(mp4|mov|webm|avi|m4v)(\?|$)/i.test(url);
}

type Lightbox = { media: string[]; index: number };

export function SubmissionHistory({ challengeId }: { challengeId: string }) {
  const [history, setHistory] = useState<SubmissionHistoryItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [lightbox, setLightbox] = useState<Lightbox | null>(null);

  useEffect(() => {
    clientFetch<SubmissionHistoryItem[]>(`/challenges/${challengeId}/submissions`)
      .then(setHistory)
      .finally(() => setLoading(false));
  }, [challengeId]);

  const closeLightbox = useCallback(() => setLightbox(null), []);

  const navigate = useCallback((dir: 1 | -1) => {
    setLightbox((lb) => {
      if (!lb) return null;
      const next = lb.index + dir;
      if (next < 0 || next >= lb.media.length) return lb;
      return { ...lb, index: next };
    });
  }, []);

  useEffect(() => {
    if (!lightbox) return;
    const handler = (e: KeyboardEvent) => {
      if (e.key === "Escape") closeLightbox();
      if (e.key === "ArrowRight") navigate(1);
      if (e.key === "ArrowLeft") navigate(-1);
    };
    window.addEventListener("keydown", handler);
    return () => window.removeEventListener("keydown", handler);
  }, [lightbox, closeLightbox, navigate]);

  if (loading) {
    return (
      <div className="py-14 text-center text-[var(--text-muted)] text-sm">Loading…</div>
    );
  }

  if (history.length === 0) {
    return (
      <div className="py-14 text-center text-[var(--text-muted)] text-sm">
        No submissions yet.
      </div>
    );
  }

  return (
    <>
      <div className="flex flex-col gap-2">
        {history.map((item) => (
          <div
            key={item.id}
            className="bg-[var(--surface)] border border-[var(--border)] rounded-2xl overflow-hidden"
          >
            <div className="flex items-center justify-between px-5 py-4">
              <div>
                <p className="text-sm font-semibold text-[var(--text)]">
                  {formatDate(item.date)}
                </p>
                <p className="text-[11px] text-[var(--text-muted)] mt-0.5">
                  {item.submission_type === "missed"
                    ? "Missed"
                    : `Submitted at ${formatTimestamp(item.submitted_at)}`}
                </p>
              </div>
              <div className="text-right flex items-center gap-2">
                {item.submission_type === "missed" && (
                  <span className="text-[10px] bg-[var(--danger-dim)] text-[var(--danger)] border border-[var(--danger)]/20 px-2 py-0.5 rounded-md font-semibold">
                    Missed
                  </span>
                )}
                <p className="font-mono-nums font-semibold text-sm tabular-nums text-[var(--text)]">
                  {formatPoints(item.total_points_earned)}
                </p>
              </div>
            </div>

            {item.metrics.length > 0 && (
              <div className="border-t border-[var(--border-subtle)] px-5 py-3 flex flex-col gap-2.5">
                {item.metrics.map((mv) => (
                  <div
                    key={mv.metric_id}
                    className="flex items-center justify-between"
                  >
                    <div className="flex items-center gap-2">
                      <span
                        className={`w-1.5 h-1.5 rounded-full shrink-0 ${
                          mv.passed ? "bg-[var(--success)]" : "bg-[var(--danger)]"
                        }`}
                      />
                      <span className="text-xs text-[var(--text-muted)]">{mv.metric_name}</span>
                    </div>
                    <div className="flex items-center gap-3">
                      <span className="font-mono-nums text-[11px] text-[var(--text-dim)] tabular-nums">
                        {mv.value}
                      </span>
                      <span
                        className={[
                          "text-[11px] font-medium tabular-nums font-mono-nums",
                          mv.passed ? "text-[var(--success)]" : "text-[var(--text-dim)]",
                        ].join(" ")}
                      >
                        {mv.passed ? `+${formatPoints(mv.points_awarded)}` : "0 pts"}
                      </span>
                    </div>
                  </div>
                ))}
              </div>
            )}

            {item.media.length > 0 && (
              <div className="border-t border-[var(--border-subtle)] px-5 py-3">
                <div className="flex gap-1.5">
                  {item.media.slice(0, 4).map((url, idx) => (
                    <button
                      key={url}
                      onClick={() => setLightbox({ media: item.media, index: idx })}
                      className="w-14 h-14 rounded-lg overflow-hidden shrink-0 focus:outline-none focus:ring-2 focus:ring-[var(--accent)] ring-offset-1 ring-offset-[var(--surface)]"
                    >
                      {isVideo(url) ? (
                        <video
                          src={url}
                          preload="metadata"
                          muted
                          playsInline
                          className="w-full h-full object-cover"
                        />
                      ) : (
                        <img
                          src={url}
                          alt=""
                          loading="lazy"
                          className="w-full h-full object-cover"
                        />
                      )}
                    </button>
                  ))}
                </div>
              </div>
            )}
          </div>
        ))}
      </div>

      {lightbox && (
        <div
          className="fixed inset-0 z-50 flex items-center justify-center bg-black/90"
          onClick={closeLightbox}
        >
          {/* Close button */}
          <button
            onClick={closeLightbox}
            className="absolute top-4 right-4 z-10 text-white/70 hover:text-white text-2xl leading-none p-2 bg-white/10 hover:bg-white/20 rounded-lg transition-all"
            aria-label="Close"
          >
            ✕
          </button>

          {/* Prev arrow */}
          {lightbox.index > 0 && (
            <button
              onClick={(e) => { e.stopPropagation(); navigate(-1); }}
              className="absolute left-4 z-10 text-white/70 hover:text-white text-3xl p-3 bg-white/10 hover:bg-white/20 rounded-lg transition-all"
              aria-label="Previous"
            >
              ‹
            </button>
          )}

          {/* Media */}
          <div
            className="max-w-full max-h-[90vh] flex items-center justify-center"
            onClick={(e) => e.stopPropagation()}
          >
            {isVideo(lightbox.media[lightbox.index]) ? (
              <video
                key={lightbox.media[lightbox.index]}
                src={lightbox.media[lightbox.index]}
                controls
                autoPlay
                className="max-h-[85vh] max-w-full rounded-xl"
              />
            ) : (
              <img
                key={lightbox.media[lightbox.index]}
                src={lightbox.media[lightbox.index]}
                alt=""
                className="max-h-[90vh] max-w-full object-contain rounded-xl"
              />
            )}
          </div>

          {/* Next arrow */}
          {lightbox.index < lightbox.media.length - 1 && (
            <button
              onClick={(e) => { e.stopPropagation(); navigate(1); }}
              className="absolute right-4 z-10 text-white/70 hover:text-white text-3xl p-3 bg-white/10 hover:bg-white/20 rounded-lg transition-all"
              aria-label="Next"
            >
              ›
            </button>
          )}

          {/* Dot indicators */}
          {lightbox.media.length > 1 && (
            <div className="absolute bottom-4 flex gap-1.5">
              {lightbox.media.map((_, i) => (
                <span
                  key={i}
                  className={`w-1.5 h-1.5 rounded-full transition-all ${
                    i === lightbox.index ? "bg-white scale-125" : "bg-white/30"
                  }`}
                />
              ))}
            </div>
          )}
        </div>
      )}
    </>
  );
}
