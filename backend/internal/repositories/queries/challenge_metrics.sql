-- name: CreateChallengeMetric :one
INSERT INTO challenge_metrics (challenge_id, metric_id, metric_type, target_value, points, fine_amount)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: ListChallengeMetrics :many
SELECT
    cm.id, cm.challenge_id, cm.metric_id, cm.metric_type,
    cm.target_value, cm.points, cm.fine_amount, cm.created_at,
    m.name AS metric_name,
    m.unit AS metric_unit
FROM challenge_metrics cm
JOIN metrics m ON cm.metric_id = m.id
WHERE cm.challenge_id = $1
ORDER BY cm.created_at;
