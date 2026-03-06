package repositories

import (
	"context"
	"errors"

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

func (r *SubmissionRepository) SetMediaKey(ctx context.Context, id uuid.UUID, fileID string) error {
	return r.q.SetSubmissionMediaKey(ctx, id, &fileID)
}

func (r *SubmissionRepository) ListMediaKeysByChallenge(ctx context.Context, challengeID uuid.UUID) ([]string, error) {
	keys, err := r.q.ListMediaKeysByChallenge(ctx, challengeID)
	if err != nil {
		return nil, err
	}
	result := make([]string, 0, len(keys))
	for _, k := range keys {
		if k != nil {
			result = append(result, *k)
		}
	}
	return result, nil
}
