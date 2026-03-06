package repositories

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"

	db "github.com/KaivalyaNaik/fitproof/internal/repositories/db"
)

var ErrDuplicateSubmission = errors.New("already submitted for today")

type SubmissionRepository struct {
	q *db.Queries
}

func NewSubmissionRepository(q *db.Queries) *SubmissionRepository {
	return &SubmissionRepository{q: q}
}

func (r *SubmissionRepository) CreateDailySubmission(ctx context.Context, params db.CreateDailySubmissionParams) (db.DailySubmission, error) {
	sub, err := r.q.CreateDailySubmission(ctx, params)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return db.DailySubmission{}, ErrDuplicateSubmission
		}
		return db.DailySubmission{}, err
	}
	return sub, nil
}

func (r *SubmissionRepository) CreateSubmissionMetricValue(ctx context.Context, params db.CreateSubmissionMetricValueParams) (db.SubmissionMetricValue, error) {
	return r.q.CreateSubmissionMetricValue(ctx, params)
}

func (r *SubmissionRepository) AddScorePoints(ctx context.Context, userChallengeID uuid.UUID, amount pgtype.Numeric, date pgtype.Date) (db.ChallengeScore, error) {
	return r.q.AddChallengeScorePoints(ctx, db.AddChallengeScorePointsParams{
		UserChallengeID: userChallengeID,
		Amount:          amount,
		Date:            date,
	})
}

func (r *SubmissionRepository) AddScoreFines(ctx context.Context, userChallengeID uuid.UUID, amount pgtype.Numeric) (db.ChallengeScore, error) {
	return r.q.AddChallengeScoreFines(ctx, db.AddChallengeScoreFinesParams{
		UserChallengeID: userChallengeID,
		Amount:          amount,
	})
}

func (r *SubmissionRepository) ListMembersWithoutSubmission(ctx context.Context, date pgtype.Date) ([]db.ListMembersWithoutSubmissionRow, error) {
	return r.q.ListMembersWithoutSubmission(ctx, date)
}

func (r *SubmissionRepository) ListUserSubmissions(ctx context.Context, userChallengeID uuid.UUID) ([]db.DailySubmission, error) {
	return r.q.ListUserSubmissions(ctx, userChallengeID)
}

func (r *SubmissionRepository) ListMetricValuesBySubmissions(ctx context.Context, ids []uuid.UUID) ([]db.ListMetricValuesBySubmissionsRow, error) {
	return r.q.ListMetricValuesBySubmissions(ctx, ids)
}

func (r *SubmissionRepository) GetSubmission(ctx context.Context, id uuid.UUID) (db.DailySubmission, error) {
	return r.q.GetSubmission(ctx, id)
}

func (r *SubmissionRepository) CreateSubmissionMedia(ctx context.Context, submissionID uuid.UUID, mediaKey string) (db.SubmissionMedia, error) {
	return r.q.CreateSubmissionMedia(ctx, submissionID, mediaKey)
}

func (r *SubmissionRepository) CountSubmissionMedia(ctx context.Context, submissionID uuid.UUID) (int64, error) {
	return r.q.CountSubmissionMedia(ctx, submissionID)
}

func (r *SubmissionRepository) ListSubmissionMediaBySubmissions(ctx context.Context, ids []uuid.UUID) ([]db.SubmissionMedia, error) {
	return r.q.ListSubmissionMediaBySubmissions(ctx, ids)
}

func (r *SubmissionRepository) ListMediaKeysByChallenge(ctx context.Context, challengeID uuid.UUID) ([]string, error) {
	return r.q.ListMediaKeysByChallenge(ctx, challengeID)
}

func (r *SubmissionRepository) ListSubmittedWithoutMedia(ctx context.Context, date pgtype.Date) ([]db.ListSubmittedWithoutMediaRow, error) {
	return r.q.ListSubmittedWithoutMedia(ctx, date)
}

func (r *SubmissionRepository) ListExpiredSubmissionMedia(ctx context.Context, cutoff time.Time) ([]db.ExpiredSubmissionMediaRow, error) {
	return r.q.ListExpiredSubmissionMedia(ctx, cutoff)
}

func (r *SubmissionRepository) DeleteExpiredSubmissionMedia(ctx context.Context, cutoff time.Time) error {
	return r.q.DeleteExpiredSubmissionMedia(ctx, cutoff)
}
