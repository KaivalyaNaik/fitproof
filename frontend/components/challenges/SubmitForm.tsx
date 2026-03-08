"use client";

import { useEffect, useRef, useState, FormEvent, ChangeEvent } from "react";
import { useRouter } from "next/navigation";
import imageCompression from "browser-image-compression";
import { clientFetch, ApiResponseError } from "@/lib/client-api";
import type { ChallengeMetric, SubmissionResult, SubmissionHistoryItem } from "@/lib/types";
import { Button } from "@/components/ui/Button";
import { Input } from "@/components/ui/Input";
import { formatPoints } from "@/lib/utils";

interface SubmitFormProps {
  challengeId: string;
  metrics: ChallengeMetric[];
  mediaRequired?: boolean;
  mediaFineAmount?: string;
}

export function SubmitForm({ challengeId, metrics, mediaRequired, mediaFineAmount }: SubmitFormProps) {
  const router = useRouter();
  const [alreadySubmitted, setAlreadySubmitted] = useState(false);
  const [todaySubmissionId, setTodaySubmissionId] = useState<string | null>(null);
  const [checkingHistory, setCheckingHistory] = useState(true);
  const [values, setValues] = useState<Record<string, string>>(
    Object.fromEntries(metrics.map((m) => [m.metric_id, ""]))
  );
  const [result, setResult] = useState<SubmissionResult | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);
  const [mediaFiles, setMediaFiles] = useState<File[]>([]);
  const [mediaUploading, setMediaUploading] = useState(false);
  const [uploadedCount, setUploadedCount] = useState(0);
  const [mediaError, setMediaError] = useState<string | null>(null);
  const fileInputRef = useRef<HTMLInputElement>(null);

  const today = new Date().toISOString().split("T")[0];

  useEffect(() => {
    clientFetch<SubmissionHistoryItem[]>(`/challenges/${challengeId}/submissions`)
      .then((history) => {
        const todayEntry = history.find(
          (h) => h.date === today && h.submission_type === "submitted"
        );
        if (todayEntry) {
          setAlreadySubmitted(true);
          setTodaySubmissionId(todayEntry.id);
          setUploadedCount(todayEntry.media.length);
        }
      })
      .finally(() => setCheckingHistory(false));
  }, [challengeId, today]);

  async function handleSubmit(e: FormEvent) {
    e.preventDefault();
    setError(null);
    setLoading(true);
    try {
      const payload = metrics.map((m) => ({
        metric_id: m.metric_id,
        value: values[m.metric_id],
      }));
      const res = await clientFetch<SubmissionResult>(
        `/challenges/${challengeId}/submissions`,
        { method: "POST", body: JSON.stringify({ metrics: payload }) }
      );
      setResult(res);
      setAlreadySubmitted(true);
      router.refresh();
    } catch (err) {
      if (err instanceof ApiResponseError && err.status === 409) {
        setAlreadySubmitted(true);
      } else {
        setError(
          err instanceof ApiResponseError ? err.message : "Something went wrong."
        );
      }
    } finally {
      setLoading(false);
    }
  }

  function handleFileChange(e: ChangeEvent<HTMLInputElement>) {
    const files = Array.from(e.target.files ?? []);
    const remaining = 4 - uploadedCount;
    setMediaFiles((prev) => [...prev, ...files].slice(0, remaining));
    setMediaError(null);
    e.target.value = "";
  }

  async function handleMediaUpload() {
    const submissionId = result?.id ?? todaySubmissionId;
    if (!submissionId || mediaFiles.length === 0) return;
    setMediaError(null);
    setMediaUploading(true);
    let uploaded = 0;
    try {
      for (const file of mediaFiles) {
        let fileToUpload: File = file;
        if (file.type.startsWith("image/")) {
          fileToUpload = await imageCompression(file, {
            maxSizeMB: 2,
            maxWidthOrHeight: 1920,
            useWebWorker: true,
          });
        }
        const form = new FormData();
        form.append("media", fileToUpload, file.name);
        await clientFetch(
          `/challenges/${challengeId}/submissions/${submissionId}/media`,
          { method: "POST", body: form }
        );
        uploaded++;
      }
      setUploadedCount((c) => c + uploaded);
      setMediaFiles([]);
    } catch (err) {
      setMediaError(
        err instanceof ApiResponseError ? err.message : "Upload failed."
      );
    } finally {
      setMediaUploading(false);
    }
  }

  if (checkingHistory) {
    return (
      <div className="py-14 text-center text-[var(--text-muted)] text-sm">Loading…</div>
    );
  }

  if (metrics.length === 0) {
    return (
      <div className="py-14 text-center text-[var(--text-muted)] text-sm">
        No metrics configured for this challenge yet.
      </div>
    );
  }

  if (alreadySubmitted && !result) {
    return (
      <div className="flex flex-col gap-4">
        <div className="text-center py-6">
          <div className="w-10 h-10 rounded-full bg-[var(--success-dim)] flex items-center justify-center mx-auto mb-3">
            <svg width="18" height="18" viewBox="0 0 18 18" fill="none">
              <path d="M3.5 9l4 4 7-7" stroke="var(--success)" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" />
            </svg>
          </div>
          <p className="text-sm font-semibold text-[var(--text)]">Already submitted today</p>
          <p className="text-xs text-[var(--text-muted)] mt-1">You can still add more proof below.</p>
        </div>

        {todaySubmissionId && uploadedCount < 4 && (
          <div className="bg-[var(--surface)] border border-[var(--border)] rounded-2xl p-5">
            <div className="flex items-start justify-between mb-1">
              <p className="text-xs font-semibold text-[var(--text)]">Add proof</p>
              <span className="text-[11px] text-[var(--text-muted)] tabular-nums font-mono-nums">{uploadedCount}/4</span>
            </div>
            <p className="text-[11px] text-[var(--text-muted)] mb-3">
              Photo or video · max 50 MB · images are auto-compressed
            </p>
            <div className="flex flex-col gap-2.5">
              <input
                ref={fileInputRef}
                type="file"
                accept="image/*,video/*"
                multiple
                onChange={handleFileChange}
                className="hidden"
              />
              <button
                type="button"
                onClick={() => fileInputRef.current?.click()}
                className="w-full border-2 border-dashed border-[var(--border)] rounded-xl py-4 text-xs text-[var(--text-muted)] hover:border-[var(--accent)]/50 hover:text-[var(--accent)] transition-colors"
              >
                {mediaFiles.length > 0
                  ? `${mediaFiles.length} file${mediaFiles.length > 1 ? "s" : ""} selected`
                  : `Tap to add files (up to ${4 - uploadedCount})`}
              </button>
              {uploadedCount > 0 && (
                <p className="text-[11px] text-[var(--success)]">{uploadedCount} file{uploadedCount > 1 ? "s" : ""} already uploaded</p>
              )}
              {mediaFiles.length > 0 && (
                <Button type="button" size="sm" onClick={handleMediaUpload} loading={mediaUploading}>
                  Upload {mediaFiles.length} file{mediaFiles.length > 1 ? "s" : ""}
                </Button>
              )}
              {mediaError && (
                <p className="text-xs text-[var(--danger)]">{mediaError}</p>
              )}
            </div>
          </div>
        )}

        {uploadedCount >= 4 && (
          <div className="flex items-center gap-2 text-[var(--success)] text-xs font-medium px-1">
            <svg width="14" height="14" viewBox="0 0 14 14" fill="none">
              <path d="M2.5 7l3 3 6-6" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round"/>
            </svg>
            Maximum uploads reached (4/4)
          </div>
        )}
      </div>
    );
  }

  if (result) {
    return (
      <div className="flex flex-col gap-4">
        <div className="text-center py-6">
          <div className="w-12 h-12 rounded-full bg-[var(--accent-dim)] flex items-center justify-center mx-auto mb-4">
            <svg width="22" height="22" viewBox="0 0 22 22" fill="none">
              <path
                d="M4 11l5 5 9-9"
                stroke="var(--accent)"
                strokeWidth="2.5"
                strokeLinecap="round"
                strokeLinejoin="round"
              />
            </svg>
          </div>
          <p className="text-sm font-semibold text-[var(--text)]">Submission recorded</p>
          <p className="text-2xl font-bold text-[var(--accent)] mt-1 tabular-nums font-mono-nums">
            +{formatPoints(result.total_points_earned)}
          </p>
        </div>
        <div className="bg-[var(--surface)] border border-[var(--border)] rounded-2xl overflow-hidden">
          {result.metrics.map((m, i) => {
            const metric = metrics.find((cm) => cm.metric_id === m.metric_id);
            return (
              <div
                key={m.metric_id}
                className={[
                  "flex items-center justify-between px-5 py-3.5",
                  i > 0 ? "border-t border-[var(--border-subtle)]" : "",
                ].join(" ")}
              >
                <div className="flex items-center gap-2.5">
                  <span
                    className={`w-1.5 h-1.5 rounded-full shrink-0 ${
                      m.passed ? "bg-[var(--success)]" : "bg-[var(--danger)]"
                    }`}
                  />
                  <span className="text-xs text-[var(--text-muted)]">
                    {metric?.metric_name ?? m.metric_id}
                  </span>
                </div>
                <div className="flex items-center gap-3">
                  <span className="font-mono-nums text-[11px] text-[var(--text-dim)] tabular-nums">
                    {m.value}
                  </span>
                  <span
                    className={[
                      "text-xs font-semibold tabular-nums font-mono-nums",
                      m.passed ? "text-[var(--success)]" : "text-[var(--text-dim)]",
                    ].join(" ")}
                  >
                    {m.passed ? `+${formatPoints(m.points_awarded)}` : "0 pts"}
                  </span>
                </div>
              </div>
            );
          })}
        </div>

        {/* Proof upload */}
        <div className="bg-[var(--surface)] border border-[var(--border)] rounded-2xl p-5">
          <div className="flex items-start justify-between mb-1">
            <p className="text-xs font-semibold text-[var(--text)]">
              {mediaRequired ? "Upload proof (required)" : "Attach proof (optional)"}
            </p>
            <span className="text-[11px] text-[var(--text-muted)] tabular-nums font-mono-nums">{uploadedCount}/4</span>
          </div>
          <p className="text-[11px] text-[var(--text-muted)] mb-3">
            Photo or video · max 50 MB · images are auto-compressed
            {mediaRequired && mediaFineAmount && parseFloat(mediaFineAmount) > 0 && (
              <span className="ml-1 text-[var(--warning)]">· Fine: {mediaFineAmount} pts if skipped</span>
            )}
          </p>

          {uploadedCount >= 4 ? (
            <div className="flex items-center gap-2 text-[var(--success)] text-xs font-medium">
              <svg width="14" height="14" viewBox="0 0 14 14" fill="none">
                <path d="M2.5 7l3 3 6-6" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round"/>
              </svg>
              Maximum uploads reached (4/4)
            </div>
          ) : (
            <div className="flex flex-col gap-2.5">
              <input
                ref={fileInputRef}
                type="file"
                accept="image/*,video/*"
                multiple
                onChange={handleFileChange}
                className="hidden"
              />
              <button
                type="button"
                onClick={() => fileInputRef.current?.click()}
                className="w-full border-2 border-dashed border-[var(--border)] rounded-xl py-4 text-xs text-[var(--text-muted)] hover:border-[var(--accent)]/50 hover:text-[var(--accent)] transition-colors"
              >
                {mediaFiles.length > 0
                  ? `${mediaFiles.length} file${mediaFiles.length > 1 ? "s" : ""} selected`
                  : `Tap to choose files (up to ${4 - uploadedCount})`}
              </button>
              {uploadedCount > 0 && uploadedCount < 4 && (
                <p className="text-[11px] text-[var(--success)]">{uploadedCount} file{uploadedCount > 1 ? "s" : ""} uploaded successfully</p>
              )}
              {mediaFiles.length > 0 && (
                <Button
                  type="button"
                  size="sm"
                  onClick={handleMediaUpload}
                  loading={mediaUploading}
                >
                  Upload {mediaFiles.length} file{mediaFiles.length > 1 ? "s" : ""}
                </Button>
              )}
              {mediaError && (
                <p className="text-xs text-[var(--danger)]">{mediaError}</p>
              )}
            </div>
          )}
        </div>
      </div>
    );
  }

  return (
    <form onSubmit={handleSubmit} className="flex flex-col gap-3">
      {metrics.map((m) => (
        <div
          key={m.metric_id}
          className="bg-[var(--surface)] border border-[var(--border)] rounded-2xl p-5"
        >
          <div className="flex items-start justify-between mb-4">
            <div>
              <p className="text-sm font-semibold text-[var(--text)]">{m.metric_name}</p>
              <p className="text-xs text-[var(--text-muted)] mt-0.5">
                {m.metric_type === "min" ? "Min" : "Max"} {m.target_value} {m.metric_unit} → {formatPoints(m.points)}
              </p>
            </div>
            <span className="text-[10px] text-[var(--warning)] bg-[var(--warning-dim)] border border-[var(--warning)]/20 px-2 py-0.5 rounded-md font-semibold">
              Fine {formatPoints(m.fine_amount)}
            </span>
          </div>
          <Input
            type="number"
            min="0"
            step="any"
            placeholder={`${m.metric_name} in ${m.metric_unit}`}
            value={values[m.metric_id]}
            onChange={(e) =>
              setValues((prev) => ({ ...prev, [m.metric_id]: e.target.value }))
            }
            required
          />
        </div>
      ))}

      {error && (
        <p className="text-sm text-[var(--danger)] bg-[var(--danger-dim)] border border-[var(--danger)]/20 rounded-xl px-3.5 py-2.5">
          {error}
        </p>
      )}

      <Button type="submit" loading={loading} className="mt-1">
        Submit Today&#39;s Activity
      </Button>
    </form>
  );
}
