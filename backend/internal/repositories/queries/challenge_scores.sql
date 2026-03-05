-- name: CreateChallengeScore :one
INSERT INTO challenge_scores (user_challenge_id)
VALUES ($1)
RETURNING *;

-- name: AddChallengeScorePoints :one
UPDATE challenge_scores
SET total_points         = total_points + sqlc.arg(amount)::numeric,
    last_submission_date = sqlc.arg(date)::date,
    updated_at           = NOW()
WHERE user_challenge_id  = sqlc.arg(user_challenge_id)::uuid
RETURNING *;

-- name: AddChallengeScoreFines :one
UPDATE challenge_scores
SET total_fines         = total_fines + sqlc.arg(amount)::numeric,
    updated_at          = NOW()
WHERE user_challenge_id = sqlc.arg(user_challenge_id)::uuid
RETURNING *;
