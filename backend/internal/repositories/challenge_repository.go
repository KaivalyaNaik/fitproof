package repositories

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"

	db "github.com/KaivalyaNaik/fitproof/internal/repositories/db"
)

var (
	ErrInviteCodeConflict = errors.New("invite code conflict")
	ErrAlreadyMember      = errors.New("already a member of this challenge")
)

type ChallengeRepository struct {
	q *db.Queries
}

func NewChallengeRepository(q *db.Queries) *ChallengeRepository {
	return &ChallengeRepository{q: q}
}

func (r *ChallengeRepository) CreateChallenge(ctx context.Context, params db.CreateChallengeParams) (db.Challenge, error) {
	c, err := r.q.CreateChallenge(ctx, params)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return db.Challenge{}, ErrInviteCodeConflict
		}
		return db.Challenge{}, err
	}
	return c, nil
}

func (r *ChallengeRepository) GetChallengeByID(ctx context.Context, id uuid.UUID) (db.Challenge, error) {
	return r.q.GetChallengeByID(ctx, id)
}

func (r *ChallengeRepository) GetChallengeByInviteCode(ctx context.Context, code string) (db.Challenge, error) {
	return r.q.GetChallengeByInviteCode(ctx, code)
}

func (r *ChallengeRepository) ListUserChallenges(ctx context.Context, userID uuid.UUID) ([]db.ListUserChallengesRow, error) {
	return r.q.ListUserChallenges(ctx, userID)
}

func (r *ChallengeRepository) CreateChallengeMetric(ctx context.Context, params db.CreateChallengeMetricParams) (db.ChallengeMetric, error) {
	return r.q.CreateChallengeMetric(ctx, params)
}

func (r *ChallengeRepository) ListChallengeMetrics(ctx context.Context, challengeID uuid.UUID) ([]db.ListChallengeMetricsRow, error) {
	return r.q.ListChallengeMetrics(ctx, challengeID)
}

func (r *ChallengeRepository) CreateUserChallenge(ctx context.Context, params db.CreateUserChallengeParams) (db.UserChallenge, error) {
	uc, err := r.q.CreateUserChallenge(ctx, params)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return db.UserChallenge{}, ErrAlreadyMember
		}
		return db.UserChallenge{}, err
	}
	return uc, nil
}

func (r *ChallengeRepository) GetUserChallenge(ctx context.Context, userID, challengeID uuid.UUID) (db.UserChallenge, error) {
	return r.q.GetUserChallenge(ctx, db.GetUserChallengeParams{
		UserID:      userID,
		ChallengeID: challengeID,
	})
}

func (r *ChallengeRepository) CreateChallengeScore(ctx context.Context, userChallengeID uuid.UUID) (db.ChallengeScore, error) {
	return r.q.CreateChallengeScore(ctx, userChallengeID)
}

func (r *ChallengeRepository) GetLeaderboard(ctx context.Context, challengeID uuid.UUID) ([]db.GetChallengeLeaderboardRow, error) {
	return r.q.GetChallengeLeaderboard(ctx, challengeID)
}

func (r *ChallengeRepository) UpdateChallengeStatus(ctx context.Context, id uuid.UUID, status db.ChallengeStatus) (db.Challenge, error) {
	return r.q.UpdateChallengeStatus(ctx, db.UpdateChallengeStatusParams{
		ID:     id,
		Status: status,
	})
}

func (r *ChallengeRepository) LeaveChallenge(ctx context.Context, userID, challengeID uuid.UUID) (db.UserChallenge, error) {
	return r.q.LeaveChallenge(ctx, db.LeaveChallengeParams{
		UserID:      userID,
		ChallengeID: challengeID,
	})
}
