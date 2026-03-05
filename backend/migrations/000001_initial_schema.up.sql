-- Enable pgcrypto for gen_random_uuid()
CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- ============================================================
-- ENUMs
-- ============================================================

CREATE TYPE user_challenge_role AS ENUM ('host', 'cohost', 'participant');
CREATE TYPE user_challenge_status AS ENUM ('active', 'left');
CREATE TYPE metric_type AS ENUM ('min', 'max');
CREATE TYPE submission_type AS ENUM ('submitted', 'missed');
CREATE TYPE challenge_status AS ENUM ('draft', 'active', 'completed', 'cancelled');

-- ============================================================
-- AUTH
-- ============================================================

CREATE TABLE users (
    id            UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    email         TEXT        NOT NULL UNIQUE,
    password_hash TEXT        NOT NULL,
    display_name  TEXT        NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE refresh_tokens (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash  TEXT        NOT NULL UNIQUE,
    device_id   TEXT        NOT NULL,
    ip_address  TEXT,
    user_agent  TEXT,
    expires_at  TIMESTAMPTZ NOT NULL,
    revoked_at  TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_refresh_tokens_user_id ON refresh_tokens(user_id);

-- ============================================================
-- DOMAIN
-- ============================================================

CREATE TABLE metrics (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    name        TEXT        NOT NULL,
    unit        TEXT        NOT NULL,
    description TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE challenges (
    id          UUID             PRIMARY KEY DEFAULT gen_random_uuid(),
    name        TEXT             NOT NULL,
    description TEXT,
    invite_code TEXT             NOT NULL UNIQUE,
    status      challenge_status NOT NULL DEFAULT 'draft',
    start_date  DATE             NOT NULL,
    end_date    DATE             NOT NULL,
    created_by  UUID             NOT NULL REFERENCES users(id),
    created_at  TIMESTAMPTZ      NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ      NOT NULL DEFAULT NOW()
);

CREATE TABLE user_challenges (
    id           UUID                 PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id      UUID                 NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    challenge_id UUID                 NOT NULL REFERENCES challenges(id) ON DELETE CASCADE,
    role         user_challenge_role  NOT NULL DEFAULT 'participant',
    status       user_challenge_status NOT NULL DEFAULT 'active',
    joined_at    TIMESTAMPTZ          NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, challenge_id)
);

CREATE INDEX idx_user_challenges_challenge_id ON user_challenges(challenge_id);

CREATE TABLE challenge_metrics (
    id           UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    challenge_id UUID        NOT NULL REFERENCES challenges(id) ON DELETE CASCADE,
    metric_id    UUID        NOT NULL REFERENCES metrics(id),
    metric_type  metric_type NOT NULL,
    target_value NUMERIC(10,2) NOT NULL,
    points       NUMERIC(10,2) NOT NULL DEFAULT 0,
    fine_amount  NUMERIC(10,2) NOT NULL DEFAULT 0,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_challenge_metrics_challenge_id ON challenge_metrics(challenge_id);

-- ============================================================
-- EVENTS
-- ============================================================

CREATE TABLE daily_submissions (
    id                UUID            PRIMARY KEY DEFAULT gen_random_uuid(),
    user_challenge_id UUID            NOT NULL REFERENCES user_challenges(id) ON DELETE CASCADE,
    date              DATE            NOT NULL,
    submission_type   submission_type NOT NULL DEFAULT 'submitted',
    submitted_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    UNIQUE(user_challenge_id, date)
);

CREATE INDEX idx_daily_submissions_user_challenge_id ON daily_submissions(user_challenge_id);
CREATE INDEX idx_daily_submissions_date ON daily_submissions(date);

CREATE TABLE submission_metric_values (
    id            UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    submission_id UUID          NOT NULL REFERENCES daily_submissions(id) ON DELETE CASCADE,
    metric_id     UUID          NOT NULL REFERENCES metrics(id),
    value         NUMERIC(10,2) NOT NULL,
    passed        BOOLEAN       NOT NULL,
    points_awarded NUMERIC(10,2) NOT NULL DEFAULT 0,
    created_at    TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_submission_metric_values_submission_id ON submission_metric_values(submission_id);

-- ============================================================
-- DERIVED STATE
-- ============================================================

CREATE TABLE challenge_scores (
    id                   UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    user_challenge_id    UUID          NOT NULL UNIQUE REFERENCES user_challenges(id) ON DELETE CASCADE,
    total_points         NUMERIC(10,2) NOT NULL DEFAULT 0,
    total_fines          NUMERIC(10,2) NOT NULL DEFAULT 0,
    last_submission_date DATE,
    updated_at           TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);
