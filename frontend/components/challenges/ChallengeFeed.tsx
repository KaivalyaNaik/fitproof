"use client";

import { useEffect, useState, useCallback } from "react";
import { clientFetch } from "@/lib/client-api";
import type { FeedItem } from "@/lib/types";
import { formatDate } from "@/lib/utils";

function isVideo(url: string) {
  return /\.(mp4|mov|webm|avi|m4v)(\?|$)/i.test(url);
}

type Lightbox = { media: string[]; index: number };

function getInitials(name: string) {
  return name
    .split(" ")
    .map((w) => w[0])
    .join("")
    .slice(0, 2)
    .toUpperCase();
}

export function ChallengeFeed({ challengeId }: { challengeId: string }) {
  const [feed, setFeed] = useState<FeedItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [lightbox, setLightbox] = useState<Lightbox | null>(null);

  useEffect(() => {
    clientFetch<FeedItem[]>(`/challenges/${challengeId}/feed`)
      .then(setFeed)
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
    return <div className="py-14 text-center text-[var(--text-muted)] text-sm">Loading…</div>;
  }

  if (feed.length === 0) {
    return (
      <div className="py-14 text-center text-[var(--text-muted)] text-sm">
        No proof media in the last 7 days.
      </div>
    );
  }

  return (
    <>
      <div className="flex flex-col gap-3">
        {feed.map((item) => (
          <div
            key={item.submission_id}
            className="bg-[var(--surface)] border border-[var(--border)] rounded-2xl overflow-hidden"
          >
            {/* Header: avatar + name + date */}
            <div className="flex items-center gap-3 px-5 py-3.5">
              <div className="w-8 h-8 rounded-full bg-[var(--accent-dim)] text-[var(--accent)] flex items-center justify-center text-[11px] font-bold shrink-0 border border-[var(--accent)]/20">
                {getInitials(item.display_name)}
              </div>
              <div className="min-w-0">
                <p className="text-sm font-semibold text-[var(--text)] truncate">
                  {item.display_name}
                </p>
                <p className="text-[11px] text-[var(--text-muted)]">{formatDate(item.date)}</p>
              </div>
            </div>

            {/* Media grid */}
            <div className="px-5 pb-4">
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
          </div>
        ))}
      </div>

      {lightbox && (
        <div
          className="fixed inset-0 z-50 flex items-center justify-center bg-black/90"
          onClick={closeLightbox}
        >
          <button
            onClick={closeLightbox}
            className="absolute top-4 right-4 z-10 text-white/70 hover:text-white text-2xl leading-none p-2 bg-white/10 hover:bg-white/20 rounded-lg transition-all"
            aria-label="Close"
          >
            ✕
          </button>

          {lightbox.index > 0 && (
            <button
              onClick={(e) => { e.stopPropagation(); navigate(-1); }}
              className="absolute left-4 z-10 text-white/70 hover:text-white text-3xl p-3 bg-white/10 hover:bg-white/20 rounded-lg transition-all"
              aria-label="Previous"
            >
              ‹
            </button>
          )}

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

          {lightbox.index < lightbox.media.length - 1 && (
            <button
              onClick={(e) => { e.stopPropagation(); navigate(1); }}
              className="absolute right-4 z-10 text-white/70 hover:text-white text-3xl p-3 bg-white/10 hover:bg-white/20 rounded-lg transition-all"
              aria-label="Next"
            >
              ›
            </button>
          )}

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
