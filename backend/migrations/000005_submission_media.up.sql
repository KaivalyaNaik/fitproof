CREATE TABLE IF NOT EXISTS submission_media (
  id            UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  submission_id UUID        NOT NULL REFERENCES daily_submissions(id) ON DELETE CASCADE,
  media_key     TEXT        NOT NULL,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS submission_media_submission_id_idx ON submission_media(submission_id);

ALTER TABLE challenges ADD COLUMN IF NOT EXISTS media_required   BOOLEAN        NOT NULL DEFAULT false;
ALTER TABLE challenges ADD COLUMN IF NOT EXISTS media_fine_amount NUMERIC(10,2)  NOT NULL DEFAULT 0;
