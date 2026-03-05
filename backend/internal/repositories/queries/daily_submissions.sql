-- name: CreateDailySubmission :one
INSERT INTO daily_submissions (user_challenge_id, date, submission_type)
VALUES ($1, $2, $3)
RETURNING *;

-- name: ListMembersWithoutSubmission :many
SELECT
    uc.id                            AS user_challenge_id,
    COALESCE(SUM(cm.fine_amount), 0) AS total_fine
FROM user_challenges uc
JOIN challenges c              ON uc.challenge_id = c.id
LEFT JOIN challenge_metrics cm ON cm.challenge_id  = uc.challenge_id
WHERE uc.status = 'active'
  AND c.status  = 'active'
  AND NOT EXISTS (
      SELECT 1 FROM daily_submissions ds
      WHERE ds.user_challenge_id = uc.id
        AND ds.date = sqlc.arg(date)::date
  )
GROUP BY uc.id;

-- name: ListUserSubmissions :many
SELECT id, user_challenge_id, date, submission_type, submitted_at
FROM daily_submissions
WHERE user_challenge_id = sqlc.arg(user_challenge_id)::uuid
ORDER BY date DESC;
