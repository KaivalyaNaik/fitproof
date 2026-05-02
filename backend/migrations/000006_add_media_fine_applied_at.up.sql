ALTER TABLE daily_submissions
  ADD COLUMN IF NOT EXISTS media_fine_applied_at TIMESTAMPTZ;
