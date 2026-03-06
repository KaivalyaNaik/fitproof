package services

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/KaivalyaNaik/fitproof/internal/repositories"
	db "github.com/KaivalyaNaik/fitproof/internal/repositories/db"
)

var (
	ErrChallengeNotFound  = errors.New("challenge not found")
	ErrAlreadyMember      = errors.New("already a member of this challenge")
	ErrChallengeNotActive = errors.New("challenge is not active")
	ErrNotAuthorized      = errors.New("not authorized")
	ErrMetricNotFound     = errors.New("metric not found")
	ErrHostCannotLeave    = errors.New("host cannot leave — close the challenge instead")
)

type MetricRule struct {
	MetricID    uuid.UUID
	MetricType  db.MetricType
	TargetValue string
	Points      string
	FineAmount  string
}

type ChallengeResult struct {
	Challenge  db.Challenge
	Membership db.UserChallenge
}

// MediaDeleter allows the challenge service to clean up Drive files on close.
type MediaDeleter interface {
	Delete(ctx context.Context, fileID string) error
}

type ChallengeService struct {
	pool       *pgxpool.Pool
	queries    *db.Queries
	metricRepo *repositories.MetricRepository
	chalRepo   *repositories.ChallengeRepository
	subRepo    *repositories.SubmissionRepository
	media      MediaDeleter // nil when Drive is not configured
}

func NewChallengeService(
	pool *pgxpool.Pool,
	queries *db.Queries,
	metricRepo *repositories.MetricRepository,
	chalRepo *repositories.ChallengeRepository,
	subRepo *repositories.SubmissionRepository,
	media MediaDeleter,
) *ChallengeService {
	return &ChallengeService{
		pool:       pool,
		queries:    queries,
		metricRepo: metricRepo,
		chalRepo:   chalRepo,
		subRepo:    subRepo,
		media:      media,
	}
}

const inviteCodeChars = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func generateInviteCode() (string, error) {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	for i := range b {
		b[i] = inviteCodeChars[int(b[i])%len(inviteCodeChars)]
	}
	return string(b), nil
}

func parseNumeric(s string) (pgtype.Numeric, error) {
	var n pgtype.Numeric
	return n, n.Scan(s)
}

func parseDate(s string) (pgtype.Date, error) {
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return pgtype.Date{}, fmt.Errorf("invalid date %q: expected YYYY-MM-DD", s)
	}
	return pgtype.Date{Time: t, Valid: true}, nil
}

func (s *ChallengeService) CreateChallenge(ctx context.Context, userID uuid.UUID, name, description, startDate, endDate string, mediaRequired bool, mediaFineAmount string) (ChallengeResult, error) {
	start, err := parseDate(startDate)
	if err != nil {
		return ChallengeResult{}, err
	}
	end, err := parseDate(endDate)
	if err != nil {
		return ChallengeResult{}, err
	}

	var descPtr *string
	if description != "" {
		descPtr = &description
	}

	mediaFine, err := parseNumeric(mediaFineAmount)
	if err != nil {
		return ChallengeResult{}, fmt.Errorf("invalid media_fine_amount %q: %w", mediaFineAmount, err)
	}

	for attempt := 0; attempt < 3; attempt++ {
		code, err := generateInviteCode()
		if err != nil {
			return ChallengeResult{}, err
		}

		tx, err := s.pool.Begin(ctx)
		if err != nil {
			return ChallengeResult{}, err
		}

		txRepo := repositories.NewChallengeRepository(s.queries.WithTx(tx))

		challenge, err := txRepo.CreateChallenge(ctx, db.CreateChallengeParams{
			Name:            name,
			Description:     descPtr,
			InviteCode:      code,
			Status:          db.ChallengeStatusActive,
			StartDate:       start,
			EndDate:         end,
			CreatedBy:       userID,
			MediaRequired:   mediaRequired,
			MediaFineAmount: mediaFine,
		})
		if err != nil {
			_ = tx.Rollback(ctx)
			if errors.Is(err, repositories.ErrInviteCodeConflict) {
				continue
			}
			return ChallengeResult{}, err
		}

		membership, err := txRepo.CreateUserChallenge(ctx, db.CreateUserChallengeParams{
			UserID:      userID,
			ChallengeID: challenge.ID,
			Role:        db.UserChallengeRoleHost,
		})
		if err != nil {
			_ = tx.Rollback(ctx)
			return ChallengeResult{}, err
		}

		if _, err = txRepo.CreateChallengeScore(ctx, membership.ID); err != nil {
			_ = tx.Rollback(ctx)
			return ChallengeResult{}, err
		}

		if err = tx.Commit(ctx); err != nil {
			return ChallengeResult{}, err
		}

		return ChallengeResult{Challenge: challenge, Membership: membership}, nil
	}

	return ChallengeResult{}, errors.New("failed to generate unique invite code")
}

func (s *ChallengeService) GetChallenge(ctx context.Context, challengeID uuid.UUID) (db.Challenge, []db.ListChallengeMetricsRow, error) {
	challenge, err := s.chalRepo.GetChallengeByID(ctx, challengeID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.Challenge{}, nil, ErrChallengeNotFound
		}
		return db.Challenge{}, nil, err
	}

	metrics, err := s.chalRepo.ListChallengeMetrics(ctx, challengeID)
	if err != nil {
		return db.Challenge{}, nil, err
	}

	return challenge, metrics, nil
}

func (s *ChallengeService) ListUserChallenges(ctx context.Context, userID uuid.UUID) ([]db.ListUserChallengesRow, error) {
	return s.chalRepo.ListUserChallenges(ctx, userID)
}

func (s *ChallengeService) AddMetrics(ctx context.Context, challengeID, userID uuid.UUID, rules []MetricRule) ([]db.ChallengeMetric, error) {
	membership, err := s.chalRepo.GetUserChallenge(ctx, userID, challengeID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotAuthorized
		}
		return nil, err
	}
	if membership.Role != db.UserChallengeRoleHost && membership.Role != db.UserChallengeRoleCohost {
		return nil, ErrNotAuthorized
	}

	if _, err := s.chalRepo.GetChallengeByID(ctx, challengeID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrChallengeNotFound
		}
		return nil, err
	}

	type parsedRule struct {
		rule        MetricRule
		targetValue pgtype.Numeric
		points      pgtype.Numeric
		fineAmount  pgtype.Numeric
	}
	parsed := make([]parsedRule, 0, len(rules))
	for _, r := range rules {
		if _, err := s.metricRepo.GetMetricByID(ctx, r.MetricID); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, ErrMetricNotFound
			}
			return nil, err
		}
		tv, err := parseNumeric(r.TargetValue)
		if err != nil {
			return nil, fmt.Errorf("invalid target_value %q: %w", r.TargetValue, err)
		}
		pts, err := parseNumeric(r.Points)
		if err != nil {
			return nil, fmt.Errorf("invalid points %q: %w", r.Points, err)
		}
		fine, err := parseNumeric(r.FineAmount)
		if err != nil {
			return nil, fmt.Errorf("invalid fine_amount %q: %w", r.FineAmount, err)
		}
		parsed = append(parsed, parsedRule{r, tv, pts, fine})
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	txRepo := repositories.NewChallengeRepository(s.queries.WithTx(tx))
	result := make([]db.ChallengeMetric, 0, len(parsed))
	for _, p := range parsed {
		cm, err := txRepo.CreateChallengeMetric(ctx, db.CreateChallengeMetricParams{
			ChallengeID: challengeID,
			MetricID:    p.rule.MetricID,
			MetricType:  p.rule.MetricType,
			TargetValue: p.targetValue,
			Points:      p.points,
			FineAmount:  p.fineAmount,
		})
		if err != nil {
			return nil, err
		}
		result = append(result, cm)
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}

	return result, nil
}

func (s *ChallengeService) CloseChallenge(ctx context.Context, challengeID, userID uuid.UUID, status db.ChallengeStatus) (db.Challenge, error) {
	uc, err := s.chalRepo.GetUserChallenge(ctx, userID, challengeID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.Challenge{}, ErrChallengeNotFound
		}
		return db.Challenge{}, err
	}
	if uc.Role != db.UserChallengeRoleHost && uc.Role != db.UserChallengeRoleCohost {
		return db.Challenge{}, ErrNotAuthorized
	}

	challenge, err := s.chalRepo.UpdateChallengeStatus(ctx, challengeID, status)
	if err != nil {
		return db.Challenge{}, err
	}

	// Asynchronously delete media files from R2 — don't fail the close if storage is unavailable.
	if s.media != nil && s.subRepo != nil {
		go func() {
			bgCtx := context.Background()
			keys, err := s.subRepo.ListMediaKeysByChallenge(bgCtx, challengeID)
			if err != nil {
				return
			}
			for _, key := range keys {
				_ = s.media.Delete(bgCtx, key)
			}
		}()
	}

	return challenge, nil
}

func (s *ChallengeService) LeaveChallenge(ctx context.Context, challengeID, userID uuid.UUID) error {
	uc, err := s.chalRepo.GetUserChallenge(ctx, userID, challengeID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrChallengeNotFound
		}
		return err
	}
	if uc.Role == db.UserChallengeRoleHost {
		return ErrHostCannotLeave
	}
	_, err = s.chalRepo.LeaveChallenge(ctx, userID, challengeID)
	return err
}

func (s *ChallengeService) GetLeaderboard(ctx context.Context, challengeID uuid.UUID) ([]db.GetChallengeLeaderboardRow, error) {
	if _, err := s.chalRepo.GetChallengeByID(ctx, challengeID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrChallengeNotFound
		}
		return nil, err
	}
	return s.chalRepo.GetLeaderboard(ctx, challengeID)
}

func (s *ChallengeService) JoinChallenge(ctx context.Context, userID uuid.UUID, inviteCode string) (ChallengeResult, error) {
	challenge, err := s.chalRepo.GetChallengeByInviteCode(ctx, inviteCode)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ChallengeResult{}, ErrChallengeNotFound
		}
		return ChallengeResult{}, err
	}

	if challenge.Status != db.ChallengeStatusActive {
		return ChallengeResult{}, ErrChallengeNotActive
	}

	_, err = s.chalRepo.GetUserChallenge(ctx, userID, challenge.ID)
	if err == nil {
		return ChallengeResult{}, ErrAlreadyMember
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return ChallengeResult{}, err
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return ChallengeResult{}, err
	}
	defer tx.Rollback(ctx)

	txRepo := repositories.NewChallengeRepository(s.queries.WithTx(tx))

	membership, err := txRepo.CreateUserChallenge(ctx, db.CreateUserChallengeParams{
		UserID:      userID,
		ChallengeID: challenge.ID,
		Role:        db.UserChallengeRoleParticipant,
	})
	if err != nil {
		if errors.Is(err, repositories.ErrAlreadyMember) {
			return ChallengeResult{}, ErrAlreadyMember
		}
		return ChallengeResult{}, err
	}

	if _, err = txRepo.CreateChallengeScore(ctx, membership.ID); err != nil {
		return ChallengeResult{}, err
	}

	if err = tx.Commit(ctx); err != nil {
		return ChallengeResult{}, err
	}

	return ChallengeResult{Challenge: challenge, Membership: membership}, nil
}
