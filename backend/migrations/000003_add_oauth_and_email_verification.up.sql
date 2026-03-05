-- Make password_hash nullable so OAuth-only users can exist
ALTER TABLE users ALTER COLUMN password_hash DROP NOT NULL;

-- Add Google OAuth identity column and email verification flag
ALTER TABLE users
    ADD COLUMN google_id      TEXT UNIQUE,
    ADD COLUMN email_verified BOOLEAN NOT NULL DEFAULT FALSE;

-- Email verification OTP tokens
CREATE TABLE email_verification_tokens (
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    code       TEXT        NOT NULL,       -- bcrypt hash of the 6-digit OTP
    expires_at TIMESTAMPTZ NOT NULL,
    used_at    TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_evt_user_id ON email_verification_tokens(user_id);
