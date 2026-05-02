-- name: CreateSubmissionMedia :one
INSERT INTO submission_media (submission_id, media_key)
VALUES ($1, $2)
RETURNING id, submission_id, media_key, created_at;

-- name: CountSubmissionMedia :one
SELECT COUNT(*) FROM submission_media WHERE submission_id = $1;

-- name: ListSubmissionMediaBySubmissions :many
SELECT id, submission_id, media_key, created_at
FROM submission_media
WHERE submission_id = ANY($1::uuid[])
ORDER BY created_at;

-- name: ListMediaKeysByChallenge :many
SELECT sm.media_key
FROM submission_media sm
JOIN daily_submissions ds ON ds.id = sm.submission_id
JOIN user_challenges uc   ON uc.id = ds.user_challenge_id
WHERE uc.challenge_id = $1;

-- name: ListSubmittedWithoutMedia :many
SELECT ds.id, ds.user_challenge_id, c.media_fine_amount
FROM daily_submissions ds
JOIN user_challenges uc ON uc.id = ds.user_challenge_id
JOIN challenges c        ON c.id  = uc.challenge_id
WHERE ds.date           = $1
  AND ds.submission_type = 'submitted'
  AND ds.media_fine_applied_at IS NULL
  AND c.status           = 'active'
  AND c.media_required   = true
  AND NOT EXISTS (
    SELECT 1 FROM submission_media sm WHERE sm.submission_id = ds.id
  );

-- name: MarkMediaFineApplied :exec
UPDATE daily_submissions
SET media_fine_applied_at = NOW()
WHERE id = $1
  AND media_fine_applied_at IS NULL;

-- name: ListExpiredSubmissionMedia :many
SELECT id, media_key FROM submission_media WHERE created_at < $1;

-- name: DeleteExpiredSubmissionMedia :exec
DELETE FROM submission_media WHERE created_at < $1;

-- name: ListChallengeFeed :many
SELECT ds.id AS submission_id, u.id AS user_id, u.display_name, ds.date, sm.media_key
FROM submission_media sm
JOIN daily_submissions ds ON ds.id = sm.submission_id
JOIN user_challenges uc   ON uc.id = ds.user_challenge_id
JOIN users u              ON u.id  = uc.user_id
WHERE uc.challenge_id = $1
  AND sm.created_at >= NOW() - INTERVAL '7 days'
ORDER BY ds.date DESC, u.display_name, sm.created_at;
