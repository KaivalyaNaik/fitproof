DROP TABLE IF EXISTS email_verification_tokens;

ALTER TABLE users DROP COLUMN IF EXISTS email_verified;
ALTER TABLE users DROP COLUMN IF EXISTS google_id;

ALTER TABLE users ALTER COLUMN password_hash SET NOT NULL;
