-- name: CreateChallenge :one
INSERT INTO challenges (name, description, invite_code, status, start_date, end_date, created_by, media_required, media_fine_amount)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING id, name, description, invite_code, status, start_date, end_date, created_by, created_at, updated_at, media_required, media_fine_amount;

-- name: GetChallengeByID :one
SELECT id, name, description, invite_code, status, start_date, end_date, created_by, created_at, updated_at, media_required, media_fine_amount
FROM challenges WHERE id = $1;

-- name: GetChallengeByInviteCode :one
SELECT id, name, description, invite_code, status, start_date, end_date, created_by, created_at, updated_at, media_required, media_fine_amount
FROM challenges WHERE invite_code = $1;

-- name: ListUserChallenges :many
SELECT
    c.id, c.name, c.description, c.invite_code, c.status,
    c.start_date, c.end_date, c.created_by, c.created_at, c.updated_at,
    c.media_required, c.media_fine_amount,
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

-- name: GetChallengeFinesSummary :many
SELECT
    u.id           AS user_id,
    u.display_name,
    cs.total_fines,
    COUNT(ds.id) FILTER (WHERE ds.submission_type = 'missed')::bigint AS missed_days,
    COALESCE(
        COUNT(ds.id) FILTER (WHERE ds.media_fine_applied_at IS NOT NULL)::numeric
            * c.media_fine_amount,
        0
    )::numeric(10,2) AS media_fines,
    (cs.total_fines - COALESCE(
        COUNT(ds.id) FILTER (WHERE ds.media_fine_applied_at IS NOT NULL)::numeric
            * c.media_fine_amount,
        0
    ))::numeric(10,2) AS missed_day_fines,
    COALESCE(
        string_agg(to_char(ds.date, 'YYYY-MM-DD'), ',' ORDER BY ds.date) FILTER (WHERE ds.submission_type = 'missed'),
        ''
    )::text AS missed_dates
FROM challenge_scores cs
JOIN user_challenges uc ON cs.user_challenge_id = uc.id
JOIN users u             ON uc.user_id = u.id
JOIN challenges c        ON c.id = uc.challenge_id
LEFT JOIN daily_submissions ds ON ds.user_challenge_id = uc.id
WHERE uc.challenge_id = sqlc.arg(challenge_id)::uuid
  AND uc.status = 'active'
GROUP BY u.id, u.display_name, cs.total_fines, c.media_fine_amount
ORDER BY cs.total_fines DESC;

-- name: UpdateChallengeStatus :one
UPDATE challenges
SET status     = sqlc.arg(status)::challenge_status,
    updated_at = NOW()
WHERE id       = sqlc.arg(id)::uuid
RETURNING id, name, description, invite_code, status, start_date, end_date, created_by, created_at, updated_at, media_required, media_fine_amount;
