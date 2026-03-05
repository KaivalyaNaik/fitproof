-- name: CreateChallenge :one
INSERT INTO challenges (name, description, invite_code, status, start_date, end_date, created_by)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: GetChallengeByID :one
SELECT * FROM challenges WHERE id = $1;

-- name: GetChallengeByInviteCode :one
SELECT * FROM challenges WHERE invite_code = $1;

-- name: ListUserChallenges :many
SELECT
    c.id, c.name, c.description, c.invite_code, c.status,
    c.start_date, c.end_date, c.created_by, c.created_at, c.updated_at,
    uc.id        AS uc_id,
    uc.role      AS uc_role,
    uc.status    AS uc_status,
    uc.joined_at AS uc_joined_at
FROM challenges c
JOIN user_challenges uc ON c.id = uc.challenge_id
WHERE uc.user_id = $1 AND uc.status = 'active'
ORDER BY c.created_at DESC;

-- name: GetChallengeLeaderboard :many
SELECT
    u.id           AS user_id,
    u.display_name,
    cs.total_points,
    cs.total_fines,
    cs.last_submission_date,
    RANK() OVER (ORDER BY cs.total_points DESC, cs.total_fines ASC)::bigint AS rank
FROM challenge_scores cs
JOIN user_challenges uc ON cs.user_challenge_id = uc.id
JOIN users u             ON uc.user_id = u.id
WHERE uc.challenge_id = sqlc.arg(challenge_id)::uuid
  AND uc.status = 'active'
ORDER BY rank;

-- name: UpdateChallengeStatus :one
UPDATE challenges
SET status     = sqlc.arg(status)::challenge_status,
    updated_at = NOW()
WHERE id       = sqlc.arg(id)::uuid
RETURNING *;
