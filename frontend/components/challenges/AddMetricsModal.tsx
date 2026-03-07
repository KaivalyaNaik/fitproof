"use client";

import { useState, useEffect } from "react";
import { useRouter } from "next/navigation";
import { clientFetch, ApiResponseError } from "@/lib/client-api";
import type { Metric } from "@/lib/types";
import { Modal } from "@/components/ui/Modal";
import { Button } from "@/components/ui/Button";
import { Input } from "@/components/ui/Input";

interface MetricRow {
  metric_id: string;
  metric_type: "min" | "max";
  target_value: string;
  points: string;
  fine_amount: string;
}

interface AddMetricsModalProps {
  open: boolean;
  onClose: () => void;
  challengeId: string;
}

export function AddMetricsModal({ open, onClose, challengeId }: AddMetricsModalProps) {
  const router = useRouter();
  const [catalog, setCatalog] = useState<Metric[]>([]);
  const [rows, setRows] = useState<MetricRow[]>([
    { metric_id: "", metric_type: "min", target_value: "", points: "", fine_amount: "" },
  ]);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    if (open && catalog.length === 0) {
      clientFetch<Metric[]>("/metrics").then(setCatalog).catch(() => {});
    }
  }, [open, catalog.length]);

  function addRow() {
    setRows((prev) => [
      ...prev,
      { metric_id: "", metric_type: "min", target_value: "", points: "", fine_amount: "" },
    ]);
  }

  function removeRow(i: number) {
    setRows((prev) => prev.filter((_, idx) => idx !== i));
  }

  function updateRow(i: number, field: keyof MetricRow, value: string) {
    setRows((prev) =>
      prev.map((row, idx) => (idx === i ? { ...row, [field]: value } : row))
    );
  }

  async function handleSubmit() {
    const payload = rows.filter((r) => r.metric_id !== "");
    if (payload.length === 0) {
      onClose();
      return;
    }
    setError(null);
    setLoading(true);
    try {
      await clientFetch(`/challenges/${challengeId}/metrics`, {
        method: "POST",
        body: JSON.stringify(payload),
      });
      setRows([{ metric_id: "", metric_type: "min", target_value: "", points: "", fine_amount: "" }]);
      onClose();
      router.refresh();
    } catch (err) {
      setError(err instanceof ApiResponseError ? err.message : "Something went wrong.");
    } finally {
      setLoading(false);
    }
  }

  const selectClass =
    "block w-full rounded-xl border border-[var(--border)] pl-3.5 pr-9 py-2.5 text-sm text-[var(--text)] bg-[var(--surface)] focus:outline-none focus:border-[var(--accent)] transition-colors appearance-none cursor-pointer";

  return (
    <Modal open={open} onClose={onClose} title="Add Metrics">
      <div className="flex flex-col gap-4">
        {rows.map((row, i) => (
          <div
            key={i}
            className="bg-[var(--surface-raised)] border border-[var(--border)] rounded-xl p-4 flex flex-col gap-3.5"
          >
            <div className="flex items-center justify-between">
              <span className="text-[10px] font-semibold text-[var(--text-muted)] uppercase tracking-widest">
                Metric {i + 1}
              </span>
              {rows.length > 1 && (
                <button
                  type="button"
                  onClick={() => removeRow(i)}
                  className="text-xs text-[var(--danger)] hover:text-[var(--danger)]/80 font-medium transition-colors"
                >
                  Remove
                </button>
              )}
            </div>

            <div className="flex flex-col gap-1.5">
              <label className="text-[10px] font-semibold text-[var(--text-muted)] uppercase tracking-widest">
                Metric
              </label>
              <div className="relative">
                <select
                  className={selectClass}
                  value={row.metric_id}
                  onChange={(e) => updateRow(i, "metric_id", e.target.value)}
                >
                  <option value="">Select a metric…</option>
                  {catalog.map((m) => (
                    <option key={m.id} value={m.id}>
                      {m.name} ({m.unit})
                    </option>
                  ))}
                </select>
                <div className="pointer-events-none absolute inset-y-0 right-3 flex items-center">
                  <svg width="12" height="12" viewBox="0 0 12 12" fill="none">
                    <path d="M2 4l4 4 4-4" stroke="var(--text-muted)" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round"/>
                  </svg>
                </div>
              </div>
            </div>

            <div className="flex items-center gap-1 p-1 bg-[var(--surface)] border border-[var(--border)] rounded-lg w-fit">
              {(["min", "max"] as const).map((type) => (
                <label
                  key={type}
                  className={[
                    "px-3 py-1 text-xs font-semibold rounded-md cursor-pointer transition-all",
                    row.metric_type === type
                      ? "bg-[var(--accent)] text-[var(--accent-fg)]"
                      : "text-[var(--text-muted)] hover:text-[var(--text)]",
                  ].join(" ")}
                >
                  <input
                    type="radio"
                    value={type}
                    checked={row.metric_type === type}
                    onChange={() => updateRow(i, "metric_type", type)}
                    className="sr-only"
                  />
                  {type === "min" ? "Minimum" : "Maximum"}
                </label>
              ))}
            </div>

            <div className="grid grid-cols-3 gap-2">
              <Input
                label="Target"
                type="number"
                min="0"
                step="any"
                placeholder="10000"
                value={row.target_value}
                onChange={(e) => updateRow(i, "target_value", e.target.value)}
              />
              <Input
                label="Points"
                type="number"
                min="0"
                step="any"
                placeholder="10"
                value={row.points}
                onChange={(e) => updateRow(i, "points", e.target.value)}
              />
              <Input
                label="Fine"
                type="number"
                min="0"
                step="any"
                placeholder="2"
                value={row.fine_amount}
                onChange={(e) => updateRow(i, "fine_amount", e.target.value)}
              />
            </div>
          </div>
        ))}

        <button
          type="button"
          onClick={addRow}
          className="text-xs text-[var(--text-muted)] hover:text-[var(--accent)] font-medium self-start transition-colors"
        >
          + Add another metric
        </button>

        {error && (
          <p className="text-sm text-[var(--danger)] bg-[var(--danger-dim)] border border-[var(--danger)]/20 rounded-xl px-3.5 py-2.5">
            {error}
          </p>
        )}

        <div className="flex justify-end gap-2 pt-1">
          <Button variant="secondary" onClick={onClose}>
            Cancel
          </Button>
          <Button onClick={handleSubmit} loading={loading}>
            Save Metrics
          </Button>
        </div>
      </div>
    </Modal>
  );
}
