"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import type { ChallengeDetail, LeaderboardEntry } from "@/lib/types";
import { clientFetch, ApiResponseError } from "@/lib/client-api";
import { Badge } from "@/components/ui/Badge";
import { Button } from "@/components/ui/Button";
import { Modal } from "@/components/ui/Modal";
import { LeaderboardTable } from "./LeaderboardTable";
import { SubmissionHistory } from "./SubmissionHistory";
import { SubmitForm } from "./SubmitForm";
import { AddMetricsModal } from "./AddMetricsModal";
import { formatDate, statusLabel } from "@/lib/utils";

type Tab = "leaderboard" | "history" | "submit";

interface Props {
  challenge: ChallengeDetail;
  leaderboard: LeaderboardEntry[];
}

export function ChallengeDetailClient({ challenge, leaderboard }: Props) {
  const router = useRouter();
  const [activeTab, setActiveTab] = useState<Tab>("leaderboard");
  const [closeModalOpen, setCloseModalOpen] = useState(false);
  const [leaveModalOpen, setLeaveModalOpen] = useState(false);
  const [addMetricsOpen, setAddMetricsOpen] = useState(false);
  const [copyLabel, setCopyLabel] = useState<string | null>(null);
  const [actionLoading, setActionLoading] = useState(false);
  const [actionError, setActionError] = useState<string | null>(null);

  const isHost =
    challenge.membership.role === "host" ||
    challenge.membership.role === "cohost";
  const isActive = challenge.status === "active";

  function handleCopyInvite() {
    navigator.clipboard.writeText(challenge.invite_code).then(() => {
      setCopyLabel("Copied!");
      setTimeout(() => setCopyLabel(null), 2000);
    });
  }

  async function handleClose(status: "completed" | "cancelled") {
    setActionError(null);
    setActionLoading(true);
    try {
      await clientFetch(`/challenges/${challenge.id}/status`, {
        method: "PATCH",
        body: JSON.stringify({ status }),
      });
      setCloseModalOpen(false);
      router.refresh();
    } catch (err) {
      setActionError(
        err instanceof ApiResponseError ? err.message : "Something went wrong."
      );
    } finally {
      setActionLoading(false);
    }
  }

  async function handleLeave() {
    setActionError(null);
    setActionLoading(true);
    try {
      await clientFetch(`/challenges/${challenge.id}/leave`, { method: "POST" });
      router.push("/dashboard");
    } catch (err) {
      setActionError(
        err instanceof ApiResponseError ? err.message : "Something went wrong."
      );
    } finally {
      setActionLoading(false);
    }
  }

  const tabs: { id: Tab; label: string }[] = [
    { id: "leaderboard", label: "Leaderboard" },
    { id: "history", label: "My History" },
    { id: "submit", label: "Submit Today" },
  ];

  return (
    <main className="max-w-5xl mx-auto px-4 sm:px-6 py-10">
      {/* Header */}
      <div className="mb-8">
        <div className="flex items-start justify-between gap-4">
          <div className="min-w-0">
            <div className="flex items-center gap-2.5 mb-1.5 flex-wrap">
              <h1 className="text-xl font-semibold text-zinc-900 truncate">
                {challenge.name}
              </h1>
              <Badge
                variant={challenge.status as "active" | "completed" | "cancelled" | "draft"}
                label={statusLabel(challenge.status)}
              />
            </div>
            {challenge.description && (
              <p className="text-sm text-zinc-400 leading-relaxed">{challenge.description}</p>
            )}
            <p className="text-[11px] text-zinc-400 mt-1.5 tabular-nums">
              {formatDate(challenge.start_date)} – {formatDate(challenge.end_date)}
            </p>
          </div>

          <div className="flex gap-2 shrink-0 flex-wrap justify-end">
            {isHost && (
              <Button
                variant="secondary"
                size="sm"
                onClick={handleCopyInvite}
                title="Copy invite code"
              >
                <svg width="12" height="12" viewBox="0 0 12 12" fill="none" className="shrink-0">
                  <rect x="4" y="4" width="7" height="7" rx="1.5" stroke="currentColor" strokeWidth="1.25"/>
                  <path d="M3 8H2a1 1 0 01-1-1V2a1 1 0 011-1h5a1 1 0 011 1v1" stroke="currentColor" strokeWidth="1.25" strokeLinecap="round"/>
                </svg>
                {copyLabel ?? `Invite · ${challenge.invite_code}`}
              </Button>
            )}
            {isHost && isActive && (
              <>
                <Button
                  variant="secondary"
                  size="sm"
                  onClick={() => setAddMetricsOpen(true)}
                >
                  Add Metrics
                </Button>
                <Button
                  variant="secondary"
                  size="sm"
                  onClick={() => setCloseModalOpen(true)}
                >
                  Close
                </Button>
              </>
            )}
            {!isHost && (
              <Button
                variant="ghost"
                size="sm"
                onClick={() => setLeaveModalOpen(true)}
              >
                Leave
              </Button>
            )}
          </div>
        </div>

        {/* Metrics chips */}
        {challenge.metrics.length > 0 && (
          <div className="mt-4 flex flex-wrap gap-1.5">
            {challenge.metrics.map((m) => (
              <span
                key={m.id}
                className="text-[11px] bg-zinc-50 text-zinc-500 border border-zinc-100 px-2.5 py-1 rounded-full font-medium"
              >
                {m.metric_name} {m.metric_type === "min" ? "≥" : "≤"} {m.target_value} {m.metric_unit} · {m.points} pts
              </span>
            ))}
          </div>
        )}
      </div>

      {/* Pill tabs */}
      <div className="flex gap-1 p-1 bg-zinc-100 rounded-xl w-fit mb-8">
        {tabs.map((tab) => (
          <button
            key={tab.id}
            onClick={() => setActiveTab(tab.id)}
            className={[
              "px-4 py-1.5 text-xs font-semibold rounded-lg transition-all duration-150",
              activeTab === tab.id
                ? "bg-white text-zinc-900 shadow-sm"
                : "text-zinc-500 hover:text-zinc-700",
            ].join(" ")}
          >
            {tab.label}
          </button>
        ))}
      </div>

      {/* Tab content */}
      <div>
        {activeTab === "leaderboard" && (
          <div className="bg-white rounded-2xl ring-1 ring-zinc-100 overflow-hidden">
            <LeaderboardTable entries={leaderboard} />
          </div>
        )}
        {activeTab === "history" && (
          <SubmissionHistory challengeId={challenge.id} />
        )}
        {activeTab === "submit" && (
          <SubmitForm challengeId={challenge.id} metrics={challenge.metrics} />
        )}
      </div>

      {/* Add metrics modal */}
      <AddMetricsModal
        open={addMetricsOpen}
        onClose={() => setAddMetricsOpen(false)}
        challengeId={challenge.id}
      />

      {/* Close challenge modal */}
      <Modal
        open={closeModalOpen}
        onClose={() => setCloseModalOpen(false)}
        title="Close Challenge"
      >
        <p className="text-sm text-zinc-500 mb-5">
          How do you want to close this challenge?
        </p>
        {actionError && (
          <p className="text-sm text-red-500 mb-3">{actionError}</p>
        )}
        <div className="flex flex-col gap-2">
          <Button
            onClick={() => handleClose("completed")}
            loading={actionLoading}
            className="w-full"
          >
            Mark as Completed
          </Button>
          <Button
            variant="danger"
            onClick={() => handleClose("cancelled")}
            loading={actionLoading}
            className="w-full"
          >
            Cancel Challenge
          </Button>
          <Button
            variant="ghost"
            onClick={() => setCloseModalOpen(false)}
            className="w-full"
          >
            Never mind
          </Button>
        </div>
      </Modal>

      {/* Leave modal */}
      <Modal
        open={leaveModalOpen}
        onClose={() => setLeaveModalOpen(false)}
        title="Leave Challenge"
      >
        <p className="text-sm text-zinc-500 mb-5">
          Are you sure you want to leave{" "}
          <strong className="text-zinc-800">{challenge.name}</strong>? You
          won&apos;t be able to rejoin without a new invite code.
        </p>
        {actionError && (
          <p className="text-sm text-red-500 mb-3">{actionError}</p>
        )}
        <div className="flex justify-end gap-2">
          <Button variant="secondary" onClick={() => setLeaveModalOpen(false)}>
            Cancel
          </Button>
          <Button
            variant="danger"
            onClick={handleLeave}
            loading={actionLoading}
          >
            Leave Challenge
          </Button>
        </div>
      </Modal>
    </main>
  );
}
