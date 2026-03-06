ALTER TABLE challenges DROP COLUMN IF EXISTS media_fine_amount;
ALTER TABLE challenges DROP COLUMN IF EXISTS media_required;
DROP TABLE IF EXISTS submission_media;
