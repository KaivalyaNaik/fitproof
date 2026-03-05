"use client";

import { useState, FormEvent } from "react";
import { useRouter } from "next/navigation";
import { clientFetch, ApiResponseError } from "@/lib/client-api";
import type { Challenge, Metric } from "@/lib/types";
import { Button } from "@/components/ui/Button";
import { Input } from "@/components/ui/Input";

interface MetricRow {
  metric_id: string;
  metric_type: "min" | "max";
  target_value: string;
  points: string;
  fine_amount: string;
}

function StepIndicator({ step }: { step: 1 | 2 }) {
  return (
    <div className="flex items-center gap-3 mb-10">
      <div className="flex items-center gap-2.5">
        <div
          className={[
            "w-6 h-6 rounded-full flex items-center justify-center text-xs font-semibold shrink-0",
            step >= 1
              ? "bg-zinc-900 text-white"
              : "border-2 border-zinc-200 text-zinc-400",
          ].join(" ")}
        >
          {step > 1 ? (
            <svg width="10" height="10" viewBox="0 0 10 10" fill="none">
              <path d="M1.5 5l2.5 2.5 4.5-4.5" stroke="white" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round"/>
            </svg>
          ) : "1"}
        </div>
        <span className="text-xs font-semibold text-zinc-900">
          Challenge Info
        </span>
      </div>
      <div className="flex-1 h-px bg-zinc-100 mx-1" />
      <div className="flex items-center gap-2.5">
        <div
          className={[
            "w-6 h-6 rounded-full flex items-center justify-center text-xs font-semibold shrink-0",
            step >= 2
              ? "bg-zinc-900 text-white"
              : "border-2 border-zinc-200 text-zinc-400",
          ].join(" ")}
        >
          2
        </div>
        <span
          className={[
            "text-xs font-semibold",
            step >= 2 ? "text-zinc-900" : "text-zinc-400",
          ].join(" ")}
        >
          Add Metrics
        </span>
      </div>
    </div>
  );
}

export function CreateChallengeForm() {
  const router = useRouter();

  // Step 1 state
  const [step, setStep] = useState<1 | 2>(1);
  const [name, setName] = useState("");
  const [description, setDescription] = useState("");
  const [startDate, setStartDate] = useState("");
  const [endDate, setEndDate] = useState("");
  const [step1Error, setStep1Error] = useState<string | null>(null);
  const [step1Loading, setStep1Loading] = useState(false);

  // Created challenge (after step 1)
  const [createdChallenge, setCreatedChallenge] = useState<Challenge | null>(null);

  // Step 2 state
  const [metrics, setMetrics] = useState<Metric[]>([]);
  const [metricRows, setMetricRows] = useState<MetricRow[]>([
    { metric_id: "", metric_type: "min", target_value: "", points: "", fine_amount: "" },
  ]);
  const [step2Error, setStep2Error] = useState<string | null>(null);
  const [step2Loading, setStep2Loading] = useState(false);

  async function handleStep1Submit(e: FormEvent) {
    e.preventDefault();
    if (endDate <= startDate) {
      setStep1Error("End date must be after start date.");
      return;
    }
    setStep1Error(null);
    setStep1Loading(true);
    try {
      const challenge = await clientFetch<Challenge>("/challenges", {
        method: "POST",
        body: JSON.stringify({
          name,
          description,
          start_date: startDate,
          end_date: endDate,
        }),
      });
      setCreatedChallenge(challenge);
      const catalog = await clientFetch<Metric[]>("/metrics");
      setMetrics(catalog);
      setStep(2);
    } catch (err) {
      setStep1Error(
        err instanceof ApiResponseError ? err.message : "Something went wrong."
      );
    } finally {
      setStep1Loading(false);
    }
  }

  function addRow() {
    setMetricRows((prev) => [
      ...prev,
      { metric_id: "", metric_type: "min", target_value: "", points: "", fine_amount: "" },
    ]);
  }

  function removeRow(i: number) {
    setMetricRows((prev) => prev.filter((_, idx) => idx !== i));
  }

  function updateRow(i: number, field: keyof MetricRow, value: string) {
    setMetricRows((prev) =>
      prev.map((row, idx) => (idx === i ? { ...row, [field]: value } : row))
    );
  }

  async function handleStep2Submit() {
    const payload = metricRows.filter((r) => r.metric_id !== "");
    if (payload.length === 0) {
      router.push(`/challenges/${createdChallenge!.id}`);
      return;
    }
    setStep2Error(null);
    setStep2Loading(true);
    try {
      await clientFetch(`/challenges/${createdChallenge!.id}/metrics`, {
        method: "POST",
        body: JSON.stringify(payload),
      });
      router.push(`/challenges/${createdChallenge!.id}`);
    } catch (err) {
      setStep2Error(
        err instanceof ApiResponseError ? err.message : "Something went wrong."
      );
    } finally {
      setStep2Loading(false);
    }
  }

  const today = new Date().toISOString().split("T")[0];

  const selectClass =
    "block w-full rounded-xl border border-zinc-200 px-3.5 py-2.5 text-sm text-zinc-900 bg-white focus:outline-none focus:border-zinc-900 transition-colors";

  return (
    <div className="max-w-xl mx-auto">
      <h1 className="text-xl font-semibold text-zinc-900 mb-8">New Challenge</h1>
      <StepIndicator step={step} />

      {step === 1 && (
        <form onSubmit={handleStep1Submit} className="flex flex-col gap-5">
          <Input
            id="name"
            label="Challenge name"
            placeholder="e.g. March Step Challenge"
            value={name}
            onChange={(e) => setName(e.target.value)}
            required
          />
          <div className="flex flex-col gap-1.5">
            <label
              htmlFor="description"
              className="text-[11px] font-semibold text-zinc-400 uppercase tracking-widest"
            >
              Description{" "}
              <span className="normal-case tracking-normal text-zinc-300">
                (optional)
              </span>
            </label>
            <textarea
              id="description"
              rows={3}
              className="block w-full rounded-xl border border-zinc-200 px-3.5 py-2.5 text-sm text-zinc-900 placeholder-zinc-400 focus:outline-none focus:border-zinc-900 resize-none transition-colors"
              placeholder="What's this challenge about?"
              value={description}
              onChange={(e) => setDescription(e.target.value)}
            />
          </div>
          <div className="grid grid-cols-2 gap-4">
            <Input
              id="start_date"
              label="Start date"
              type="date"
              min={today}
              value={startDate}
              onChange={(e) => setStartDate(e.target.value)}
              required
            />
            <Input
              id="end_date"
              label="End date"
              type="date"
              min={startDate || today}
              value={endDate}
              onChange={(e) => setEndDate(e.target.value)}
              required
            />
          </div>

          {step1Error && (
            <p className="text-sm text-red-600 bg-red-50 rounded-xl px-3.5 py-2.5">
              {step1Error}
            </p>
          )}

          <Button type="submit" loading={step1Loading}>
            Continue
          </Button>
        </form>
      )}

      {step === 2 && (
        <div className="flex flex-col gap-4">
          <div className="mb-2">
            <h2 className="text-sm font-semibold text-zinc-900 mb-1">
              Add metrics to track
            </h2>
            <p className="text-xs text-zinc-400">
              Define what participants log each day and how they earn points.
            </p>
          </div>

          {metricRows.map((row, i) => (
            <div
              key={i}
              className="bg-white rounded-2xl ring-1 ring-zinc-100 p-5 flex flex-col gap-4"
            >
              <div className="flex items-center justify-between">
                <span className="text-[11px] font-semibold text-zinc-400 uppercase tracking-widest">
                  Metric {i + 1}
                </span>
                {metricRows.length > 1 && (
                  <button
                    type="button"
                    onClick={() => removeRow(i)}
                    className="text-xs text-red-500 hover:text-red-700 font-medium"
                  >
                    Remove
                  </button>
                )}
              </div>

              <div className="flex flex-col gap-1.5">
                <label className="text-[11px] font-semibold text-zinc-400 uppercase tracking-widest">
                  Metric
                </label>
                <select
                  className={selectClass}
                  value={row.metric_id}
                  onChange={(e) => updateRow(i, "metric_id", e.target.value)}
                >
                  <option value="">Select a metric…</option>
                  {metrics.map((m) => (
                    <option key={m.id} value={m.id}>
                      {m.name} ({m.unit})
                    </option>
                  ))}
                </select>
              </div>

              <div className="flex items-center gap-1 p-1 bg-zinc-100 rounded-lg w-fit">
                {(["min", "max"] as const).map((type) => (
                  <label
                    key={type}
                    className={[
                      "px-3 py-1 text-xs font-semibold rounded-md cursor-pointer transition-all",
                      row.metric_type === type
                        ? "bg-white text-zinc-900 shadow-sm"
                        : "text-zinc-500",
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

              <div className="grid grid-cols-3 gap-3">
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
            className="text-xs text-zinc-500 hover:text-zinc-900 font-medium self-start transition-colors"
          >
            + Add another metric
          </button>

          {step2Error && (
            <p className="text-sm text-red-600 bg-red-50 rounded-xl px-3.5 py-2.5">
              {step2Error}
            </p>
          )}

          <div className="flex items-center justify-between pt-2">
            <button
              type="button"
              onClick={() => router.push(`/challenges/${createdChallenge!.id}`)}
              className="text-xs text-zinc-400 hover:text-zinc-700 font-medium transition-colors"
            >
              Skip for now
            </button>
            <Button onClick={handleStep2Submit} loading={step2Loading}>
              Save & Open Challenge
            </Button>
          </div>
        </div>
      )}
    </div>
  );
}
