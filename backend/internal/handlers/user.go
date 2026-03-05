package handlers

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/KaivalyaNaik/fitproof/internal/middleware"
	"github.com/KaivalyaNaik/fitproof/internal/repositories"
	"github.com/KaivalyaNaik/fitproof/pkg/respond"
	"github.com/google/uuid"
)

type UserHandler struct {
	userRepo *repositories.UserRepository
	logger   *slog.Logger
}

func NewUserHandler(repo *repositories.UserRepository, logger *slog.Logger) *UserHandler {
	return &UserHandler{userRepo: repo, logger: logger}
}

func (h *UserHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(uuid.UUID)
	if !ok {
		respond.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	stats, err := h.userRepo.GetUserStats(r.Context(), userID)
	if err != nil {
		h.logger.Error("get stats failed", slog.String("error", err.Error()))
		respond.Error(w, http.StatusInternalServerError, "internal server error")
		return
	}

	respond.JSON(w, http.StatusOK, struct {
		ChallengesJoined  int64  `json:"challenges_joined"`
		TotalPoints       string `json:"total_points"`
		TotalFines        string `json:"total_fines"`
		TotalSubmissions  int64  `json:"total_submissions"`
		MissedSubmissions int64  `json:"missed_submissions"`
	}{
		ChallengesJoined:  stats.ChallengesJoined,
		TotalPoints:       numericString(stats.TotalPoints),
		TotalFines:        numericString(stats.TotalFines),
		TotalSubmissions:  stats.TotalSubmissions,
		MissedSubmissions: stats.MissedSubmissions,
	})
}

func (h *UserHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(uuid.UUID)
	if !ok {
		respond.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	user, err := h.userRepo.GetUserByID(r.Context(), userID)
	if err != nil {
		h.logger.Error("get me failed", slog.String("error", err.Error()))
		respond.Error(w, http.StatusInternalServerError, "internal server error")
		return
	}

	respond.JSON(w, http.StatusOK, userResponse{
		ID:          user.ID.String(),
		Email:       user.Email,
		DisplayName: user.DisplayName,
		CreatedAt:   user.CreatedAt.Time.Format(time.RFC3339),
	})
}
