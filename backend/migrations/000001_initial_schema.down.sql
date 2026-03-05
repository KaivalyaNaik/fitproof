-- Drop tables in reverse FK dependency order
DROP TABLE IF EXISTS challenge_scores;
DROP TABLE IF EXISTS submission_metric_values;
DROP TABLE IF EXISTS daily_submissions;
DROP TABLE IF EXISTS challenge_metrics;
DROP TABLE IF EXISTS user_challenges;
DROP TABLE IF EXISTS challenges;
DROP TABLE IF EXISTS metrics;
DROP TABLE IF EXISTS refresh_tokens;
DROP TABLE IF EXISTS users;

-- Drop ENUMs
DROP TYPE IF EXISTS challenge_status;
DROP TYPE IF EXISTS submission_type;
DROP TYPE IF EXISTS metric_type;
DROP TYPE IF EXISTS user_challenge_status;
DROP TYPE IF EXISTS user_challenge_role;
