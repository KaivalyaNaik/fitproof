-- name: CreateRefreshToken :one
INSERT INTO refresh_tokens (user_id, token_hash, device_id, ip_address, user_agent, expires_at)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetRefreshTokenByHash :one
SELECT * FROM refresh_tokens WHERE token_hash = $1 LIMIT 1;

-- name: RevokeRefreshToken :exec
UPDATE refresh_tokens SET revoked_at = NOW() WHERE id = $1;

-- name: RevokeDeviceTokens :exec
UPDATE refresh_tokens
SET revoked_at = NOW()
WHERE user_id = $1
  AND device_id = $2
  AND revoked_at IS NULL
  AND expires_at > NOW();
