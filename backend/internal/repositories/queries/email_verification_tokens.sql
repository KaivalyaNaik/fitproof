-- name: CreateEmailVerificationToken :one
INSERT INTO email_verification_tokens (user_id, code, expires_at)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetLatestTokenForUser :one
SELECT * FROM email_verification_tokens
WHERE user_id  = $1
  AND used_at  IS NULL
  AND expires_at > NOW()
ORDER BY created_at DESC
LIMIT 1;

-- name: MarkTokenUsed :exec
UPDATE email_verification_tokens
SET used_at = NOW()
WHERE id = $1;

-- name: DeleteUserTokens :exec
DELETE FROM email_verification_tokens
WHERE user_id = $1;
