-- name: CreateUser :one
INSERT INTO users (email, password_hash, display_name)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1 LIMIT 1;

-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1 LIMIT 1;

-- name: CreateUserOAuth :one
INSERT INTO users (email, display_name, google_id, email_verified)
VALUES ($1, $2, $3, TRUE)
RETURNING *;

-- name: GetUserByGoogleID :one
SELECT * FROM users WHERE google_id = $1 LIMIT 1;

-- name: UpdateUserGoogleID :one
UPDATE users
SET google_id  = $2,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: SetEmailVerified :exec
UPDATE users
SET email_verified = TRUE,
    updated_at     = NOW()
WHERE id = $1;

-- name: GetUserStats :one
SELECT
    COUNT(DISTINCT uc.id)::bigint                                        AS challenges_joined,
    COALESCE(SUM(cs.total_points), 0)::numeric                           AS total_points,
    COALESCE(SUM(cs.total_fines),  0)::numeric                           AS total_fines,
    COUNT(ds.id) FILTER (WHERE ds.submission_type = 'submitted')::bigint AS total_submissions,
    COUNT(ds.id) FILTER (WHERE ds.submission_type = 'missed')::bigint    AS missed_submissions
FROM user_challenges uc
LEFT JOIN challenge_scores  cs ON cs.user_challenge_id = uc.id
LEFT JOIN daily_submissions ds ON ds.user_challenge_id = uc.id
WHERE uc.user_id = sqlc.arg(user_id)::uuid
  AND uc.status  = 'active';
