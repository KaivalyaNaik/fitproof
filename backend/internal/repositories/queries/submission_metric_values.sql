-- name: CreateSubmissionMetricValue :one
INSERT INTO submission_metric_values (submission_id, metric_id, value, passed, points_awarded)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: ListMetricValuesBySubmissions :many
SELECT
    smv.id, smv.submission_id, smv.metric_id,
    smv.value, smv.passed, smv.points_awarded, smv.created_at,
    m.name AS metric_name
FROM submission_metric_values smv
JOIN metrics m ON m.id = smv.metric_id
WHERE smv.submission_id = ANY(sqlc.arg(submission_ids)::uuid[])
ORDER BY smv.submission_id, smv.created_at;
