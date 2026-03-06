import { redirect } from "next/navigation";
import { serverFetch } from "@/lib/server-api";
import type { ChallengeListItem, LeaderboardEntry } from "@/lib/types";
import { ChallengeDetailClient } from "@/components/challenges/ChallengeDetailClient";

interface ChallengeDetailResponse {
  id: string;
  name: string;
  description: string;
  invite_code: string;
  status: string;
  start_date: string;
  end_date: string;
  created_at: string;
  media_required: boolean;
  media_fine_amount: string;
  metrics: {
    id: string;
    metric_id: string;
    metric_name: string;
    metric_unit: string;
    metric_type: "min" | "max";
    target_value: string;
    points: string;
    fine_amount: string;
  }[];
}

export default async function ChallengePage({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  const { id } = await params;

  let detail: ChallengeDetailResponse;
  let allChallenges: ChallengeListItem[];
  let leaderboard: LeaderboardEntry[];
  try {
    [detail, allChallenges, leaderboard] = await Promise.all([
      serverFetch<ChallengeDetailResponse>(`/challenges/${id}`),
      serverFetch<ChallengeListItem[]>("/challenges"),
      serverFetch<LeaderboardEntry[]>(`/challenges/${id}/leaderboard`),
    ]);
  } catch {
    redirect("/dashboard");
  }

  const membership = allChallenges.find((c) => c.id === id)?.membership;
  if (!membership) redirect("/dashboard");

  return (
    <ChallengeDetailClient
      challenge={{ ...detail, membership, status: detail.status as import("@/lib/types").ChallengeStatus }}
      leaderboard={leaderboard}
    />
  );
}
