-- name: ListMetrics :many
SELECT * FROM metrics ORDER BY name;

-- name: GetMetricByID :one
SELECT * FROM metrics WHERE id = $1;
