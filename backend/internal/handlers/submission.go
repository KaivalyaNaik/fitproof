package handlers

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/KaivalyaNaik/fitproof/internal/repositories"
	"github.com/KaivalyaNaik/fitproof/internal/services"
	"github.com/KaivalyaNaik/fitproof/pkg/respond"
)

const maxMediaSize = 50 << 20 // 50 MB

type SubmissionHandler struct {
	svc    *services.SubmissionService
	logger *slog.Logger
}

func NewSubmissionHandler(svc *services.SubmissionService, logger *slog.Logger) *SubmissionHandler {
	return &SubmissionHandler{svc: svc, logger: logger}
}

type submitMetricRequest struct {
	MetricID string `json:"metric_id"`
	Value    string `json:"value"`
}

type submitRequest struct {
	Metrics []submitMetricRequest `json:"metrics"`
}

func (h *SubmissionHandler) ListSubmissions(w http.ResponseWriter, r *http.Request) {
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

	items, err := h.svc.ListUserSubmissions(r.Context(), userID, challengeID)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrNotMember):
			respond.Error(w, http.StatusForbidden, "not a member of this challenge")
		default:
			h.logger.Error("list submissions failed", slog.String("error", err.Error()))
			respond.Error(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}
	respond.JSON(w, http.StatusOK, items)
}

func (h *SubmissionHandler) Submit(w http.ResponseWriter, r *http.Request) {
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

	var req submitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if len(req.Metrics) == 0 {
		respond.Error(w, http.StatusBadRequest, "at least one metric is required")
		return
	}

	mvs := make([]services.MetricValue, 0, len(req.Metrics))
	for _, m := range req.Metrics {
		metricID, err := uuid.Parse(m.MetricID)
		if err != nil {
			respond.Error(w, http.StatusBadRequest, "invalid metric_id: "+m.MetricID)
			return
		}
		if m.Value == "" {
			respond.Error(w, http.StatusBadRequest, "value is required for each metric")
			return
		}
		mvs = append(mvs, services.MetricValue{MetricID: metricID, Value: m.Value})
	}

	result, err := h.svc.Submit(r.Context(), userID, challengeID, mvs)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrNotMember):
			respond.Error(w, http.StatusForbidden, "not a member of this challenge")
		case errors.Is(err, services.ErrUnknownMetric):
			respond.Error(w, http.StatusUnprocessableEntity, "metric not part of this challenge")
		case errors.Is(err, repositories.ErrDuplicateSubmission):
			respond.Error(w, http.StatusConflict, "already submitted for today")
		default:
			h.logger.Error("submit failed", slog.String("error", err.Error()))
			respond.Error(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	type metricResult struct {
		MetricID      string `json:"metric_id"`
		Value         string `json:"value"`
		Passed        bool   `json:"passed"`
		PointsAwarded string `json:"points_awarded"`
	}
	metrics := make([]metricResult, len(result.Metrics))
	for i, m := range result.Metrics {
		metrics[i] = metricResult{
			MetricID:      m.MetricID.String(),
			Value:         m.Value,
			Passed:        m.Passed,
			PointsAwarded: m.PointsAwarded,
		}
	}

	respond.JSON(w, http.StatusCreated, struct {
		ID                string         `json:"id"`
		Date              string         `json:"date"`
		Metrics           []metricResult `json:"metrics"`
		TotalPointsEarned string         `json:"total_points_earned"`
	}{
		ID:                result.Submission.ID.String(),
		Date:              result.Submission.Date.Time.Format(time.DateOnly),
		Metrics:           metrics,
		TotalPointsEarned: result.TotalPointsEarned,
	})
}

func (h *SubmissionHandler) UploadMedia(w http.ResponseWriter, r *http.Request) {
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

	subID, err := uuid.Parse(chi.URLParam(r, "subID"))
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid submission id")
		return
	}

	if err := r.ParseMultipartForm(maxMediaSize + (1 << 20)); err != nil {
		respond.Error(w, http.StatusBadRequest, "request too large or malformed")
		return
	}

	file, header, err := r.FormFile("media")
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "media file is required")
		return
	}
	defer file.Close()

	contentType := header.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "image/") && !strings.HasPrefix(contentType, "video/") {
		respond.Error(w, http.StatusBadRequest, "only image or video files are allowed")
		return
	}

	if header.Size > maxMediaSize {
		respond.Error(w, http.StatusBadRequest, "file too large (max 50 MB)")
		return
	}

	fileKey, err := h.svc.UploadMedia(r.Context(), userID, challengeID, subID, header.Filename, contentType, file)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrNotMember):
			respond.Error(w, http.StatusForbidden, "not a member of this challenge")
		case errors.Is(err, services.ErrMediaNotConfigured):
			respond.Error(w, http.StatusServiceUnavailable, "media storage not configured")
		case errors.Is(err, services.ErrMediaLimitReached):
			respond.Error(w, http.StatusUnprocessableEntity, "maximum 4 media files per submission")
		default:
			h.logger.Error("upload media failed", slog.String("error", err.Error()))
			respond.Error(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	respond.JSON(w, http.StatusOK, map[string]string{"media_key": fileKey})
}
