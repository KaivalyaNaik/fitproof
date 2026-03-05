-- name: CreateUserChallenge :one
INSERT INTO user_challenges (user_id, challenge_id, role)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetUserChallenge :one
SELECT * FROM user_challenges
WHERE user_id = $1 AND challenge_id = $2;

-- name: LeaveChallenge :one
UPDATE user_challenges
SET status = 'left'
WHERE user_id      = sqlc.arg(user_id)::uuid
  AND challenge_id = sqlc.arg(challenge_id)::uuid
RETURNING *;
