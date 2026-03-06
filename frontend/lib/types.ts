// Auth
export interface User {
  id: string;
  email: string;
  display_name: string;
  email_verified: boolean;
  created_at: string;
}

export interface UserStats {
  challenges_joined: number;
  total_points: string;
  total_fines: string;
  total_submissions: number;
  missed_submissions: number;
}

// Challenges
export type ChallengeStatus = "draft" | "active" | "completed" | "cancelled";
export type UserRole = "host" | "cohost" | "participant";

export interface Membership {
  id: string;
  role: UserRole;
  joined_at: string;
}

export interface Challenge {
  id: string;
  name: string;
  description: string;
  invite_code: string;
  status: ChallengeStatus;
  start_date: string;
  end_date: string;
  created_at: string;
  media_required: boolean;
  media_fine_amount: string;
}

export interface ChallengeListItem extends Challenge {
  membership: Membership;
}

export interface ChallengeMetric {
  id: string;
  metric_id: string;
  metric_name: string;
  metric_unit: string;
  metric_type: "min" | "max";
  target_value: string;
  points: string;
  fine_amount: string;
}

export interface ChallengeDetail extends Challenge {
  metrics: ChallengeMetric[];
  membership: Membership;
}

// Metrics catalog
export interface Metric {
  id: string;
  name: string;
  unit: string;
  description: string;
}

// Leaderboard
export interface LeaderboardEntry {
  rank: number;
  user_id: string;
  display_name: string;
  total_points: string;
  total_fines: string;
  last_submission_date?: string;
}

// Submissions
export interface MetricResult {
  metric_id: string;
  value: string;
  passed: boolean;
  points_awarded: string;
}

export interface SubmissionResult {
  id: string;
  date: string;
  metrics: MetricResult[];
  total_points_earned: string;
}

export interface MetricValueDetail {
  metric_id: string;
  metric_name: string;
  value: string;
  passed: boolean;
  points_awarded: string;
}

export interface SubmissionHistoryItem {
  id: string;
  date: string;
  submission_type: "submitted" | "missed";
  submitted_at: string;
  metrics: MetricValueDetail[];
  total_points_earned: string;
  media: string[];
}
