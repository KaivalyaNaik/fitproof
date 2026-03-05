package handlers

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/KaivalyaNaik/fitproof/internal/middleware"
	"github.com/KaivalyaNaik/fitproof/internal/services"
	"github.com/KaivalyaNaik/fitproof/pkg/respond"
)

type AuthHandler struct {
	authService   *services.AuthService
	logger        *slog.Logger
	isProd        bool
	accessTTL     time.Duration
	refreshTTL    time.Duration
	frontendURL   string
	googleEnabled bool
}

func NewAuthHandler(svc *services.AuthService, logger *slog.Logger, isProd bool, accessTTL, refreshTTL time.Duration, frontendURL string, googleEnabled bool) *AuthHandler {
	return &AuthHandler{
		authService:   svc,
		logger:        logger,
		isProd:        isProd,
		accessTTL:     accessTTL,
		refreshTTL:    refreshTTL,
		frontendURL:   frontendURL,
		googleEnabled: googleEnabled,
	}
}

// ── request / response types ──────────────────────────────────────────────────

type registerRequest struct {
	Email       string `json:"email"`
	Password    string `json:"password"`
	DisplayName string `json:"display_name"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type verifyEmailRequest struct {
	Code string `json:"code"`
}

type userResponse struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	DisplayName   string `json:"display_name"`
	EmailVerified bool   `json:"email_verified"`
	CreatedAt     string `json:"created_at"`
}

// ── cookie helpers ────────────────────────────────────────────────────────────

func (h *AuthHandler) setAccessCookie(w http.ResponseWriter, token string, ttl time.Duration) {
	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    token,
		Path:     "/",
		MaxAge:   int(ttl.Seconds()),
		HttpOnly: true,
		Secure:   h.isProd,
		SameSite: http.SameSiteLaxMode,
	})
}

func (h *AuthHandler) setRefreshCookie(w http.ResponseWriter, token string, ttl time.Duration) {
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    token,
		Path:     "/auth",
		MaxAge:   int(ttl.Seconds()),
		HttpOnly: true,
		Secure:   h.isProd,
		SameSite: http.SameSiteLaxMode,
	})
}

func (h *AuthHandler) clearAuthCookies(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{Name: "access_token", Value: "", Path: "/", MaxAge: -1, HttpOnly: true, Secure: h.isProd, SameSite: http.SameSiteLaxMode})
	http.SetCookie(w, &http.Cookie{Name: "refresh_token", Value: "", Path: "/auth", MaxAge: -1, HttpOnly: true, Secure: h.isProd, SameSite: http.SameSiteLaxMode})
}

func deviceID(r *http.Request) string {
	if id := r.Header.Get("X-Device-ID"); id != "" {
		return id
	}
	return uuid.New().String()
}

// ── password auth ─────────────────────────────────────────────────────────────

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Email == "" || req.Password == "" || req.DisplayName == "" {
		respond.Error(w, http.StatusBadRequest, "email, password, and display_name are required")
		return
	}

	result, err := h.authService.Register(r.Context(), req.Email, req.Password, req.DisplayName,
		deviceID(r), r.RemoteAddr, r.Header.Get("User-Agent"))
	if err != nil {
		if errors.Is(err, services.ErrEmailAlreadyExists) {
			respond.Error(w, http.StatusConflict, "email already in use")
			return
		}
		h.logger.Error("register failed", slog.String("error", err.Error()))
		respond.Error(w, http.StatusInternalServerError, "internal server error")
		return
	}

	h.setAccessCookie(w, result.AccessToken, h.accessTTL)
	h.setRefreshCookie(w, result.RefreshToken, h.refreshTTL)
	respond.JSON(w, http.StatusCreated, userResponse{
		ID:            result.User.ID.String(),
		Email:         result.User.Email,
		DisplayName:   result.User.DisplayName,
		EmailVerified: result.User.EmailVerified,
		CreatedAt:     result.User.CreatedAt.Time.Format(time.RFC3339),
	})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Email == "" || req.Password == "" {
		respond.Error(w, http.StatusBadRequest, "email and password are required")
		return
	}

	result, err := h.authService.Login(r.Context(), req.Email, req.Password,
		deviceID(r), r.RemoteAddr, r.Header.Get("User-Agent"))
	if err != nil {
		if errors.Is(err, services.ErrInvalidCredentials) {
			respond.Error(w, http.StatusUnauthorized, "invalid credentials")
			return
		}
		h.logger.Error("login failed", slog.String("error", err.Error()))
		respond.Error(w, http.StatusInternalServerError, "internal server error")
		return
	}

	h.setAccessCookie(w, result.AccessToken, h.accessTTL)
	h.setRefreshCookie(w, result.RefreshToken, h.refreshTTL)
	respond.JSON(w, http.StatusOK, userResponse{
		ID:            result.User.ID.String(),
		Email:         result.User.Email,
		DisplayName:   result.User.DisplayName,
		EmailVerified: result.User.EmailVerified,
		CreatedAt:     result.User.CreatedAt.Time.Format(time.RFC3339),
	})
}

func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		respond.Error(w, http.StatusUnauthorized, "missing refresh token")
		return
	}

	result, err := h.authService.Refresh(r.Context(), cookie.Value,
		deviceID(r), r.RemoteAddr, r.Header.Get("User-Agent"))
	if err != nil {
		if errors.Is(err, services.ErrInvalidRefreshToken) {
			h.clearAuthCookies(w)
			respond.Error(w, http.StatusUnauthorized, "invalid or expired refresh token")
			return
		}
		h.logger.Error("refresh failed", slog.String("error", err.Error()))
		respond.Error(w, http.StatusInternalServerError, "internal server error")
		return
	}

	h.setAccessCookie(w, result.AccessToken, h.accessTTL)
	h.setRefreshCookie(w, result.RefreshToken, h.refreshTTL)
	w.WriteHeader(http.StatusNoContent)
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("refresh_token")
	if err == nil {
		_ = h.authService.Logout(r.Context(), cookie.Value)
	}
	h.clearAuthCookies(w)
	w.WriteHeader(http.StatusNoContent)
}

// ── Google OAuth ──────────────────────────────────────────────────────────────

func (h *AuthHandler) GoogleLogin(w http.ResponseWriter, r *http.Request) {
	if !h.googleEnabled {
		respond.Error(w, http.StatusNotImplemented, "Google OAuth is not configured")
		return
	}
	url, err := h.authService.GoogleLoginURL()
	if err != nil {
		h.logger.Error("google login url failed", slog.String("error", err.Error()))
		respond.Error(w, http.StatusInternalServerError, "internal server error")
		return
	}
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func (h *AuthHandler) GoogleCallback(w http.ResponseWriter, r *http.Request) {
	if !h.googleEnabled {
		http.Redirect(w, r, h.frontendURL+"/login?error=oauth_failed", http.StatusTemporaryRedirect)
		return
	}
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")

	if code == "" || state == "" {
		http.Redirect(w, r, h.frontendURL+"/login?error=oauth_failed", http.StatusTemporaryRedirect)
		return
	}

	result, err := h.authService.HandleGoogleCallback(r.Context(), code, state,
		deviceID(r), r.RemoteAddr, r.Header.Get("User-Agent"))
	if err != nil {
		h.logger.Error("google callback failed", slog.String("error", err.Error()))
		http.Redirect(w, r, h.frontendURL+"/login?error=oauth_failed", http.StatusTemporaryRedirect)
		return
	}

	h.setAccessCookie(w, result.AccessToken, h.accessTTL)
	h.setRefreshCookie(w, result.RefreshToken, h.refreshTTL)
	http.Redirect(w, r, h.frontendURL+"/dashboard", http.StatusTemporaryRedirect)
}

// ── email verification ────────────────────────────────────────────────────────

func (h *AuthHandler) SendVerificationCode(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(uuid.UUID)
	if !ok {
		respond.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	err := h.authService.SendVerificationCode(r.Context(), userID)
	if err != nil {
		if errors.Is(err, services.ErrEmailAlreadyVerified) {
			respond.Error(w, http.StatusConflict, "email already verified")
			return
		}
		h.logger.Error("send verification code failed", slog.String("error", err.Error()))
		respond.Error(w, http.StatusInternalServerError, "internal server error")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *AuthHandler) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(uuid.UUID)
	if !ok {
		respond.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req verifyEmailRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Code == "" {
		respond.Error(w, http.StatusBadRequest, "code is required")
		return
	}

	if err := h.authService.VerifyEmail(r.Context(), userID, req.Code); err != nil {
		if errors.Is(err, services.ErrInvalidVerifyCode) {
			respond.Error(w, http.StatusUnprocessableEntity, "invalid or expired code")
			return
		}
		if errors.Is(err, services.ErrEmailAlreadyVerified) {
			respond.Error(w, http.StatusConflict, "email already verified")
			return
		}
		h.logger.Error("verify email failed", slog.String("error", err.Error()))
		respond.Error(w, http.StatusInternalServerError, "internal server error")
		return
	}

	user, err := h.authService.GetUserByID(r.Context(), userID)
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "internal server error")
		return
	}
	respond.JSON(w, http.StatusOK, userResponse{
		ID:            user.ID.String(),
		Email:         user.Email,
		DisplayName:   user.DisplayName,
		EmailVerified: user.EmailVerified,
		CreatedAt:     user.CreatedAt.Time.Format(time.RFC3339),
	})
}
