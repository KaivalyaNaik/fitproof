package handlers

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/KaivalyaNaik/fitproof/internal/middleware"
	"github.com/KaivalyaNaik/fitproof/internal/repositories"
	db "github.com/KaivalyaNaik/fitproof/internal/repositories/db"
	"github.com/KaivalyaNaik/fitproof/internal/services"
	"github.com/KaivalyaNaik/fitproof/pkg/respond"
)

type ChallengeHandler struct {
	svc        *services.ChallengeService
	metricRepo *repositories.MetricRepository
	logger     *slog.Logger
}

func NewChallengeHandler(svc *services.ChallengeService, metricRepo *repositories.MetricRepository, logger *slog.Logger) *ChallengeHandler {
	return &ChallengeHandler{svc: svc, metricRepo: metricRepo, logger: logger}
}

// ── request types ────────────────────────────────────────────────────────────

type createChallengeRequest struct {
	Name            string `json:"name"`
	Description     string `json:"description"`
	StartDate       string `json:"start_date"`
	EndDate         string `json:"end_date"`
	MediaRequired   bool   `json:"media_required"`
	MediaFineAmount string `json:"media_fine_amount"`
}

type joinChallengeRequest struct {
	InviteCode string `json:"invite_code"`
}

type metricRuleRequest struct {
	MetricID    string `json:"metric_id"`
	MetricType  string `json:"metric_type"`
	TargetValue string `json:"target_value"`
	Points      string `json:"points"`
	FineAmount  string `json:"fine_amount"`
}

// ── response types ───────────────────────────────────────────────────────────

type metricResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Unit        string `json:"unit"`
	Description string `json:"description,omitempty"`
}

type challengeMetricResponse struct {
	ID          string `json:"id"`
	MetricID    string `json:"metric_id"`
	MetricName  string `json:"metric_name"`
	MetricUnit  string `json:"metric_unit"`
	MetricType  string `json:"metric_type"`
	TargetValue string `json:"target_value"`
	Points      string `json:"points"`
	FineAmount  string `json:"fine_amount"`
}

type membershipResponse struct {
	ID       string `json:"id"`
	Role     string `json:"role"`
	JoinedAt string `json:"joined_at"`
}

type challengeResponse struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	Description     string `json:"description,omitempty"`
	InviteCode      string `json:"invite_code"`
	Status          string `json:"status"`
	StartDate       string `json:"start_date"`
	EndDate         string `json:"end_date"`
	CreatedAt       string `json:"created_at"`
	MediaRequired   bool   `json:"media_required"`
	MediaFineAmount string `json:"media_fine_amount"`
}

type challengeDetailResponse struct {
	challengeResponse
	Metrics    []challengeMetricResponse `json:"metrics"`
	Membership membershipResponse        `json:"membership"`
}

type challengeListItem struct {
	challengeResponse
	Membership membershipResponse `json:"membership"`
}

// ── helpers ──────────────────────────────────────────────────────────────────

func callerID(r *http.Request) (uuid.UUID, bool) {
	id, ok := r.Context().Value(middleware.UserIDKey).(uuid.UUID)
	return id, ok
}

func toChallengeResponse(c db.Challenge) challengeResponse {
	desc := ""
	if c.Description != nil {
		desc = *c.Description
	}
	return challengeResponse{
		ID:              c.ID.String(),
		Name:            c.Name,
		Description:     desc,
		InviteCode:      c.InviteCode,
		Status:          string(c.Status),
		StartDate:       c.StartDate.Time.Format("2006-01-02"),
		EndDate:         c.EndDate.Time.Format("2006-01-02"),
		CreatedAt:       c.CreatedAt.Time.Format(time.RFC3339),
		MediaRequired:   c.MediaRequired,
		MediaFineAmount: numericString(c.MediaFineAmount),
	}
}

func toChallengeMetricResponse(cm db.ListChallengeMetricsRow) challengeMetricResponse {
	return challengeMetricResponse{
		ID:          cm.ID.String(),
		MetricID:    cm.MetricID.String(),
		MetricName:  cm.MetricName,
		MetricUnit:  cm.MetricUnit,
		MetricType:  string(cm.MetricType),
		TargetValue: numericString(cm.TargetValue),
		Points:      numericString(cm.Points),
		FineAmount:  numericString(cm.FineAmount),
	}
}

func numericString(n pgtype.Numeric) string {
	s, _ := n.MarshalJSON()
	if len(s) >= 2 && s[0] == '"' {
		return string(s[1 : len(s)-1])
	}
	return string(s)
}

func mapChallengeError(h *ChallengeHandler, w http.ResponseWriter, err error, op string) {
	switch {
	case errors.Is(err, services.ErrChallengeNotFound):
		respond.Error(w, http.StatusNotFound, "challenge not found")
	case errors.Is(err, services.ErrAlreadyMember):
		respond.Error(w, http.StatusConflict, "already a member of this challenge")
	case errors.Is(err, services.ErrChallengeNotActive):
		respond.Error(w, http.StatusUnprocessableEntity, "challenge is not active")
	case errors.Is(err, services.ErrNotAuthorized):
		respond.Error(w, http.StatusForbidden, "not authorized")
	case errors.Is(err, services.ErrMetricNotFound):
		respond.Error(w, http.StatusUnprocessableEntity, "one or more metric IDs not found")
	case errors.Is(err, services.ErrHostCannotLeave):
		respond.Error(w, http.StatusForbidden, err.Error())
	default:
		h.logger.Error(op+" failed", slog.String("error", err.Error()))
		respond.Error(w, http.StatusInternalServerError, "internal server error")
	}
}

// ── handlers ─────────────────────────────────────────────────────────────────

func (h *ChallengeHandler) ListMetrics(w http.ResponseWriter, r *http.Request) {
	metrics, err := h.metricRepo.ListMetrics(r.Context())
	if err != nil {
		h.logger.Error("list metrics failed", slog.String("error", err.Error()))
		respond.Error(w, http.StatusInternalServerError, "internal server error")
		return
	}

	out := make([]metricResponse, len(metrics))
	for i, m := range metrics {
		desc := ""
		if m.Description != nil {
			desc = *m.Description
		}
		out[i] = metricResponse{
			ID:          m.ID.String(),
			Name:        m.Name,
			Unit:        m.Unit,
			Description: desc,
		}
	}
	respond.JSON(w, http.StatusOK, out)
}

func (h *ChallengeHandler) CreateChallenge(w http.ResponseWriter, r *http.Request) {
	userID, ok := callerID(r)
	if !ok {
		respond.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req createChallengeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Name == "" || req.StartDate == "" || req.EndDate == "" {
		respond.Error(w, http.StatusBadRequest, "name, start_date, and end_date are required")
		return
	}

	mediaFine := req.MediaFineAmount
	if mediaFine == "" {
		mediaFine = "0"
	}
	result, err := h.svc.CreateChallenge(r.Context(), userID, req.Name, req.Description, req.StartDate, req.EndDate, req.MediaRequired, mediaFine)
	if err != nil {
		mapChallengeError(h, w, err, "create challenge")
		return
	}

	respond.JSON(w, http.StatusCreated, challengeDetailResponse{
		challengeResponse: toChallengeResponse(result.Challenge),
		Metrics:           []challengeMetricResponse{},
		Membership: membershipResponse{
			ID:       result.Membership.ID.String(),
			Role:     string(result.Membership.Role),
			JoinedAt: result.Membership.JoinedAt.Time.Format(time.RFC3339),
		},
	})
}

func (h *ChallengeHandler) GetChallenge(w http.ResponseWriter, r *http.Request) {
	challengeID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid challenge id")
		return
	}

	_, ok := callerID(r)
	if !ok {
		respond.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	challenge, metrics, err := h.svc.GetChallenge(r.Context(), challengeID)
	if err != nil {
		mapChallengeError(h, w, err, "get challenge")
		return
	}

	metricResp := make([]challengeMetricResponse, len(metrics))
	for i, m := range metrics {
		metricResp[i] = toChallengeMetricResponse(m)
	}

	respond.JSON(w, http.StatusOK, struct {
		challengeResponse
		Metrics []challengeMetricResponse `json:"metrics"`
	}{
		challengeResponse: toChallengeResponse(challenge),
		Metrics:           metricResp,
	})
}

func (h *ChallengeHandler) ListUserChallenges(w http.ResponseWriter, r *http.Request) {
	userID, ok := callerID(r)
	if !ok {
		respond.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	rows, err := h.svc.ListUserChallenges(r.Context(), userID)
	if err != nil {
		h.logger.Error("list challenges failed", slog.String("error", err.Error()))
		respond.Error(w, http.StatusInternalServerError, "internal server error")
		return
	}

	out := make([]challengeListItem, len(rows))
	for i, row := range rows {
		desc := ""
		if row.Description != nil {
			desc = *row.Description
		}
		out[i] = challengeListItem{
			challengeResponse: challengeResponse{
				ID:              row.ID.String(),
				Name:            row.Name,
				Description:     desc,
				InviteCode:      row.InviteCode,
				Status:          string(row.Status),
				StartDate:       row.StartDate.Time.Format("2006-01-02"),
				EndDate:         row.EndDate.Time.Format("2006-01-02"),
				CreatedAt:       row.CreatedAt.Time.Format(time.RFC3339),
				MediaRequired:   row.MediaRequired,
				MediaFineAmount: numericString(row.MediaFineAmount),
			},
			Membership: membershipResponse{
				ID:       row.UcID.String(),
				Role:     string(row.UcRole),
				JoinedAt: row.UcJoinedAt.Time.Format(time.RFC3339),
			},
		}
	}
	respond.JSON(w, http.StatusOK, out)
}

func (h *ChallengeHandler) AddMetrics(w http.ResponseWriter, r *http.Request) {
	userID, ok := callerID(r)
	if !ok {
		respond.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	challengeID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid challenge id")
		return
	}

	var reqs []metricRuleRequest
	if err := json.NewDecoder(r.Body).Decode(&reqs); err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if len(reqs) == 0 {
		respond.Error(w, http.StatusBadRequest, "at least one metric rule is required")
		return
	}

	rules := make([]services.MetricRule, 0, len(reqs))
	for _, req := range reqs {
		metricID, err := uuid.Parse(req.MetricID)
		if err != nil {
			respond.Error(w, http.StatusBadRequest, "invalid metric_id: "+req.MetricID)
			return
		}
		mt := db.MetricType(req.MetricType)
		if mt != db.MetricTypeMin && mt != db.MetricTypeMax {
			respond.Error(w, http.StatusBadRequest, "metric_type must be 'min' or 'max'")
			return
		}
		rules = append(rules, services.MetricRule{
			MetricID:    metricID,
			MetricType:  mt,
			TargetValue: req.TargetValue,
			Points:      req.Points,
			FineAmount:  req.FineAmount,
		})
	}

	created, err := h.svc.AddMetrics(r.Context(), challengeID, userID, rules)
	if err != nil {
		mapChallengeError(h, w, err, "add metrics")
		return
	}

	type addedMetric struct {
		ID          string `json:"id"`
		MetricID    string `json:"metric_id"`
		MetricType  string `json:"metric_type"`
		TargetValue string `json:"target_value"`
		Points      string `json:"points"`
		FineAmount  string `json:"fine_amount"`
	}
	out := make([]addedMetric, len(created))
	for i, cm := range created {
		tv, _ := cm.TargetValue.MarshalJSON()
		pts, _ := cm.Points.MarshalJSON()
		fine, _ := cm.FineAmount.MarshalJSON()
		stripQuotes := func(b []byte) string {
			if len(b) >= 2 && b[0] == '"' {
				return string(b[1 : len(b)-1])
			}
			return string(b)
		}
		out[i] = addedMetric{
			ID:          cm.ID.String(),
			MetricID:    cm.MetricID.String(),
			MetricType:  string(cm.MetricType),
			TargetValue: stripQuotes(tv),
			Points:      stripQuotes(pts),
			FineAmount:  stripQuotes(fine),
		}
	}
	respond.JSON(w, http.StatusCreated, out)
}

func (h *ChallengeHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	challengeID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid challenge id")
		return
	}
	userID, ok := callerID(r)
	if !ok {
		respond.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	status := db.ChallengeStatus(req.Status)
	if status != db.ChallengeStatusCompleted && status != db.ChallengeStatusCancelled {
		respond.Error(w, http.StatusBadRequest, "status must be 'completed' or 'cancelled'")
		return
	}

	challenge, err := h.svc.CloseChallenge(r.Context(), challengeID, userID, status)
	if err != nil {
		mapChallengeError(h, w, err, "update challenge status")
		return
	}
	respond.JSON(w, http.StatusOK, toChallengeResponse(challenge))
}

func (h *ChallengeHandler) Leave(w http.ResponseWriter, r *http.Request) {
	challengeID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid challenge id")
		return
	}
	userID, ok := callerID(r)
	if !ok {
		respond.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	if err := h.svc.LeaveChallenge(r.Context(), challengeID, userID); err != nil {
		mapChallengeError(h, w, err, "leave challenge")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *ChallengeHandler) GetLeaderboard(w http.ResponseWriter, r *http.Request) {
	challengeID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid challenge id")
		return
	}
	_, ok := callerID(r)
	if !ok {
		respond.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	rows, err := h.svc.GetLeaderboard(r.Context(), challengeID)
	if err != nil {
		mapChallengeError(h, w, err, "get leaderboard")
		return
	}

	type leaderboardEntry struct {
		Rank               int64  `json:"rank"`
		UserID             string `json:"user_id"`
		DisplayName        string `json:"display_name"`
		TotalPoints        string `json:"total_points"`
		TotalFines         string `json:"total_fines"`
		LastSubmissionDate string `json:"last_submission_date,omitempty"`
	}
	out := make([]leaderboardEntry, len(rows))
	for i, row := range rows {
		lastDate := ""
		if row.LastSubmissionDate.Valid {
			lastDate = row.LastSubmissionDate.Time.Format("2006-01-02")
		}
		out[i] = leaderboardEntry{
			Rank:               row.Rank,
			UserID:             row.UserID.String(),
			DisplayName:        row.DisplayName,
			TotalPoints:        numericString(row.TotalPoints),
			TotalFines:         numericString(row.TotalFines),
			LastSubmissionDate: lastDate,
		}
	}
	respond.JSON(w, http.StatusOK, out)
}

func (h *ChallengeHandler) JoinChallenge(w http.ResponseWriter, r *http.Request) {
	userID, ok := callerID(r)
	if !ok {
		respond.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req joinChallengeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.InviteCode == "" {
		respond.Error(w, http.StatusBadRequest, "invite_code is required")
		return
	}

	result, err := h.svc.JoinChallenge(r.Context(), userID, req.InviteCode)
	if err != nil {
		mapChallengeError(h, w, err, "join challenge")
		return
	}

	respond.JSON(w, http.StatusCreated, challengeDetailResponse{
		challengeResponse: toChallengeResponse(result.Challenge),
		Metrics:           []challengeMetricResponse{},
		Membership: membershipResponse{
			ID:       result.Membership.ID.String(),
			Role:     string(result.Membership.Role),
			JoinedAt: result.Membership.JoinedAt.Time.Format(time.RFC3339),
		},
	})
}
