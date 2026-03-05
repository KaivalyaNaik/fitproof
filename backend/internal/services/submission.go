package services

import (
	"context"
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
	ErrNotMember     = errors.New("not a member of this challenge")
	ErrUnknownMetric = errors.New("metric not part of this challenge")
)

type MetricValueDetail struct {
	MetricID      uuid.UUID `json:"metric_id"`
	MetricName    string    `json:"metric_name"`
	Value         string    `json:"value"`
	Passed        bool      `json:"passed"`
	PointsAwarded string    `json:"points_awarded"`
}

type SubmissionHistoryItem struct {
	ID                uuid.UUID           `json:"id"`
	Date              string              `json:"date"`
	SubmissionType    string              `json:"submission_type"`
	SubmittedAt       string              `json:"submitted_at"`
	Metrics           []MetricValueDetail `json:"metrics"`
	TotalPointsEarned string              `json:"total_points_earned"`
}

type MetricValue struct {
	MetricID uuid.UUID
	Value    string
}

type MetricEvalResult struct {
	MetricID      uuid.UUID
	Value         string
	Passed        bool
	PointsAwarded string
}

type SubmissionResult struct {
	Submission        db.DailySubmission
	Metrics           []MetricEvalResult
	TotalPointsEarned string
}

type SubmissionService struct {
	pool     *pgxpool.Pool
	queries  *db.Queries
	chalRepo *repositories.ChallengeRepository
	subRepo  *repositories.SubmissionRepository
}

func NewSubmissionService(
	pool *pgxpool.Pool,
	queries *db.Queries,
	chalRepo *repositories.ChallengeRepository,
	subRepo *repositories.SubmissionRepository,
) *SubmissionService {
	return &SubmissionService{
		pool:     pool,
		queries:  queries,
		chalRepo: chalRepo,
		subRepo:  subRepo,
	}
}

func numericF64(n pgtype.Numeric) float64 {
	f, _ := n.Float64Value()
	return f.Float64
}

func (s *SubmissionService) Submit(ctx context.Context, userID, challengeID uuid.UUID, metrics []MetricValue) (SubmissionResult, error) {
	uc, err := s.chalRepo.GetUserChallenge(ctx, userID, challengeID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return SubmissionResult{}, ErrNotMember
		}
		return SubmissionResult{}, err
	}

	rules, err := s.chalRepo.ListChallengeMetrics(ctx, challengeID)
	if err != nil {
		return SubmissionResult{}, err
	}
	ruleMap := make(map[uuid.UUID]db.ListChallengeMetricsRow, len(rules))
	for _, r := range rules {
		ruleMap[r.MetricID] = r
	}

	type parsedMetric struct {
		mv    MetricValue
		rule  db.ListChallengeMetricsRow
		value pgtype.Numeric
	}
	parsed := make([]parsedMetric, 0, len(metrics))
	for _, mv := range metrics {
		rule, ok := ruleMap[mv.MetricID]
		if !ok {
			return SubmissionResult{}, ErrUnknownMetric
		}
		v, err := parseNumeric(mv.Value)
		if err != nil {
			return SubmissionResult{}, fmt.Errorf("invalid value %q for metric %s: %w", mv.Value, mv.MetricID, err)
		}
		parsed = append(parsed, parsedMetric{mv, rule, v})
	}

	today := pgtype.Date{Time: time.Now().UTC().Truncate(24 * time.Hour), Valid: true}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return SubmissionResult{}, err
	}
	defer tx.Rollback(ctx)

	txSubRepo := repositories.NewSubmissionRepository(s.queries.WithTx(tx))

	sub, err := txSubRepo.CreateDailySubmission(ctx, db.CreateDailySubmissionParams{
		UserChallengeID: uc.ID,
		Date:            today,
		SubmissionType:  db.SubmissionTypeSubmitted,
	})
	if err != nil {
		return SubmissionResult{}, err
	}

	evalResults := make([]MetricEvalResult, 0, len(parsed))
	totalPoints := 0.0

	for _, p := range parsed {
		targetF := numericF64(p.rule.TargetValue)
		valueF := numericF64(p.value)

		passed := false
		switch p.rule.MetricType {
		case db.MetricTypeMin:
			passed = valueF >= targetF
		case db.MetricTypeMax:
			passed = valueF <= targetF
		}

		pointsF := 0.0
		if passed {
			pointsF = numericF64(p.rule.Points)
		}
		totalPoints += pointsF

		pointsNum, _ := parseNumeric(fmt.Sprintf("%.2f", pointsF))

		if _, err := txSubRepo.CreateSubmissionMetricValue(ctx, db.CreateSubmissionMetricValueParams{
			SubmissionID:  sub.ID,
			MetricID:      p.mv.MetricID,
			Value:         p.value,
			Passed:        passed,
			PointsAwarded: pointsNum,
		}); err != nil {
			return SubmissionResult{}, err
		}

		evalResults = append(evalResults, MetricEvalResult{
			MetricID:      p.mv.MetricID,
			Value:         p.mv.Value,
			Passed:        passed,
			PointsAwarded: fmt.Sprintf("%.2f", pointsF),
		})
	}

	totalPointsNum, _ := parseNumeric(fmt.Sprintf("%.2f", totalPoints))
	if _, err = txSubRepo.AddScorePoints(ctx, uc.ID, totalPointsNum, today); err != nil {
		return SubmissionResult{}, err
	}

	if err = tx.Commit(ctx); err != nil {
		return SubmissionResult{}, err
	}

	return SubmissionResult{
		Submission:        sub,
		Metrics:           evalResults,
		TotalPointsEarned: fmt.Sprintf("%.2f", totalPoints),
	}, nil
}

func (s *SubmissionService) ListUserSubmissions(ctx context.Context, userID, challengeID uuid.UUID) ([]SubmissionHistoryItem, error) {
	uc, err := s.chalRepo.GetUserChallenge(ctx, userID, challengeID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotMember
		}
		return nil, err
	}

	subs, err := s.subRepo.ListUserSubmissions(ctx, uc.ID)
	if err != nil {
		return nil, err
	}
	if len(subs) == 0 {
		return []SubmissionHistoryItem{}, nil
	}

	subIDs := make([]uuid.UUID, len(subs))
	for i, s := range subs {
		subIDs[i] = s.ID
	}

	mvRows, err := s.subRepo.ListMetricValuesBySubmissions(ctx, subIDs)
	if err != nil {
		return nil, err
	}

	mvMap := make(map[uuid.UUID][]MetricValueDetail, len(subs))
	ptMap := make(map[uuid.UUID]float64, len(subs))
	for _, mv := range mvRows {
		mvMap[mv.SubmissionID] = append(mvMap[mv.SubmissionID], MetricValueDetail{
			MetricID:      mv.MetricID,
			MetricName:    mv.MetricName,
			Value:         fmt.Sprintf("%.2f", numericF64(mv.Value)),
			Passed:        mv.Passed,
			PointsAwarded: fmt.Sprintf("%.2f", numericF64(mv.PointsAwarded)),
		})
		ptMap[mv.SubmissionID] += numericF64(mv.PointsAwarded)
	}

	result := make([]SubmissionHistoryItem, len(subs))
	for i, sub := range subs {
		metrics := mvMap[sub.ID]
		if metrics == nil {
			metrics = []MetricValueDetail{}
		}
		result[i] = SubmissionHistoryItem{
			ID:                sub.ID,
			Date:              sub.Date.Time.Format("2006-01-02"),
			SubmissionType:    string(sub.SubmissionType),
			SubmittedAt:       sub.SubmittedAt.Time.Format(time.RFC3339),
			Metrics:           metrics,
			TotalPointsEarned: fmt.Sprintf("%.2f", ptMap[sub.ID]),
		}
	}
	return result, nil
}

func (s *SubmissionService) ProcessMissedSubmissions(ctx context.Context, dateStr string) error {
	date, err := parseDate(dateStr)
	if err != nil {
		return err
	}

	rows, err := s.subRepo.ListMembersWithoutSubmission(ctx, date)
	if err != nil {
		return err
	}

	for _, row := range rows {
		tx, err := s.pool.Begin(ctx)
		if err != nil {
			return err
		}

		txSubRepo := repositories.NewSubmissionRepository(s.queries.WithTx(tx))

		_, err = txSubRepo.CreateDailySubmission(ctx, db.CreateDailySubmissionParams{
			UserChallengeID: row.UserChallengeID,
			Date:            date,
			SubmissionType:  db.SubmissionTypeMissed,
		})
		if err != nil {
			tx.Rollback(ctx)
			if errors.Is(err, repositories.ErrDuplicateSubmission) {
				continue
			}
			return err
		}

		fineNum, ok := row.TotalFine.(pgtype.Numeric)
		if ok && numericF64(fineNum) > 0 {
			if _, err = txSubRepo.AddScoreFines(ctx, row.UserChallengeID, fineNum); err != nil {
				tx.Rollback(ctx)
				return err
			}
		}

		if err = tx.Commit(ctx); err != nil {
			return err
		}
	}

	return nil
}
