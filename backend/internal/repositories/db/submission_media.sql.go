package db

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

const createSubmissionMedia = `
INSERT INTO submission_media (submission_id, media_key)
VALUES ($1, $2)
RETURNING id, submission_id, media_key, created_at
`

func (q *Queries) CreateSubmissionMedia(ctx context.Context, submissionID uuid.UUID, mediaKey string) (SubmissionMedia, error) {
	row := q.db.QueryRow(ctx, createSubmissionMedia, submissionID, mediaKey)
	var i SubmissionMedia
	err := row.Scan(&i.ID, &i.SubmissionID, &i.MediaKey, &i.CreatedAt)
	return i, err
}

const countSubmissionMedia = `SELECT COUNT(*) FROM submission_media WHERE submission_id = $1`

func (q *Queries) CountSubmissionMedia(ctx context.Context, submissionID uuid.UUID) (int64, error) {
	var n int64
	err := q.db.QueryRow(ctx, countSubmissionMedia, submissionID).Scan(&n)
	return n, err
}

const listSubmissionMediaBySubmissions = `
SELECT id, submission_id, media_key, created_at
FROM submission_media
WHERE submission_id = ANY($1::uuid[])
ORDER BY created_at
`

func (q *Queries) ListSubmissionMediaBySubmissions(ctx context.Context, ids []uuid.UUID) ([]SubmissionMedia, error) {
	rows, err := q.db.Query(ctx, listSubmissionMediaBySubmissions, ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []SubmissionMedia
	for rows.Next() {
		var i SubmissionMedia
		if err := rows.Scan(&i.ID, &i.SubmissionID, &i.MediaKey, &i.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	return items, rows.Err()
}

const listMediaKeysByChallenge = `
SELECT sm.media_key
FROM submission_media sm
JOIN daily_submissions ds ON ds.id = sm.submission_id
JOIN user_challenges uc   ON uc.id = ds.user_challenge_id
WHERE uc.challenge_id = $1
`

func (q *Queries) ListMediaKeysByChallenge(ctx context.Context, challengeID uuid.UUID) ([]string, error) {
	rows, err := q.db.Query(ctx, listMediaKeysByChallenge, challengeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var keys []string
	for rows.Next() {
		var k string
		if err := rows.Scan(&k); err != nil {
			return nil, err
		}
		keys = append(keys, k)
	}
	return keys, rows.Err()
}

type ListSubmittedWithoutMediaRow struct {
	SubmissionID    uuid.UUID
	UserChallengeID uuid.UUID
	MediaFineAmount pgtype.Numeric
}

const listSubmittedWithoutMedia = `
SELECT ds.id, ds.user_challenge_id, c.media_fine_amount
FROM daily_submissions ds
JOIN user_challenges uc ON uc.id = ds.user_challenge_id
JOIN challenges c        ON c.id  = uc.challenge_id
WHERE ds.date           = $1
  AND ds.submission_type = 'submitted'
  AND c.status           = 'active'
  AND c.media_required   = true
  AND NOT EXISTS (
    SELECT 1 FROM submission_media sm WHERE sm.submission_id = ds.id
  )
`

func (q *Queries) ListSubmittedWithoutMedia(ctx context.Context, date pgtype.Date) ([]ListSubmittedWithoutMediaRow, error) {
	rows, err := q.db.Query(ctx, listSubmittedWithoutMedia, date)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []ListSubmittedWithoutMediaRow
	for rows.Next() {
		var i ListSubmittedWithoutMediaRow
		if err := rows.Scan(&i.SubmissionID, &i.UserChallengeID, &i.MediaFineAmount); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	return items, rows.Err()
}
