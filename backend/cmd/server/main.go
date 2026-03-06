package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"github.com/KaivalyaNaik/fitproof/internal/config"
	"github.com/KaivalyaNaik/fitproof/internal/db"
	"github.com/KaivalyaNaik/fitproof/internal/handlers"
	"github.com/KaivalyaNaik/fitproof/internal/middleware"
	"github.com/KaivalyaNaik/fitproof/internal/repositories"
	repodb "github.com/KaivalyaNaik/fitproof/internal/repositories/db"
	"github.com/KaivalyaNaik/fitproof/internal/services"
	"github.com/KaivalyaNaik/fitproof/pkg/drive"
	"github.com/KaivalyaNaik/fitproof/pkg/email"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	cfg := config.Load()
	logger.Info("config loaded", slog.String("env", cfg.AppEnv))

	if err := db.RunMigrations(cfg.DatabaseURL, cfg.MigrationsPath); err != nil {
		logger.Error("migrations failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
	logger.Info("migrations applied")

	ctx := context.Background()
	pool, err := db.NewPool(ctx, cfg)
	if err != nil {
		logger.Error("database connection failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer pool.Close()
	logger.Info("database connected")

	queries := repodb.New(pool)
	userRepo := repositories.NewUserRepository(queries)
	tokenRepo := repositories.NewTokenRepository(queries)
	emailVerifRepo := repositories.NewEmailVerificationRepository(queries)

	emailSender := email.NewSender(cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPUsername, cfg.SMTPPassword, cfg.SMTPFrom)

	googleEnabled := cfg.GoogleClientID != "" && cfg.GoogleClientSecret != ""
	var oauthConfig *oauth2.Config
	if googleEnabled {
		oauthConfig = &oauth2.Config{
			ClientID:     cfg.GoogleClientID,
			ClientSecret: cfg.GoogleClientSecret,
			RedirectURL:  cfg.GoogleCallbackURL,
			Scopes:       []string{"openid", "email", "profile"},
			Endpoint:     google.Endpoint,
		}
	}

	authSvc := services.NewAuthService(
		userRepo, tokenRepo, emailVerifRepo, emailSender, oauthConfig,
		cfg.JWTSecret, cfg.JWTAccessTokenTTL, cfg.JWTRefreshTokenTTL, cfg.EmailVerificationTTL,
	)
	authH := handlers.NewAuthHandler(authSvc, logger, cfg.AppEnv == "production", cfg.JWTAccessTokenTTL, cfg.JWTRefreshTokenTTL, cfg.FrontendURL, googleEnabled)

	// Google Drive media storage (optional — disabled if credentials not set)
	var driveSvc *drive.Service
	if cfg.GoogleDriveCredentials != "" && cfg.GoogleDriveFolderID != "" {
		var driveErr error
		driveSvc, driveErr = drive.New(ctx, []byte(cfg.GoogleDriveCredentials), cfg.GoogleDriveFolderID)
		if driveErr != nil {
			logger.Error("drive init failed", slog.String("error", driveErr.Error()))
			os.Exit(1)
		}
		logger.Info("google drive media storage enabled")
	}

	metricRepo := repositories.NewMetricRepository(queries)
	chalRepo := repositories.NewChallengeRepository(queries)
	subRepo := repositories.NewSubmissionRepository(queries)
	chalSvc := services.NewChallengeService(pool, queries, metricRepo, chalRepo, subRepo, driveSvc)
	chalH := handlers.NewChallengeHandler(chalSvc, metricRepo, logger)
	userH := handlers.NewUserHandler(userRepo, logger)

	subSvc := services.NewSubmissionService(pool, queries, chalRepo, subRepo, driveSvc)
	subH := handlers.NewSubmissionHandler(subSvc, logger)

	r := chi.NewRouter()
	r.Use(middleware.Recovery(logger))
	r.Use(middleware.RequestLogger(logger))
	r.Use(chimiddleware.RealIP)

	r.Get("/health", handlers.HealthHandler(pool))
	r.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("."))
	})

	r.Route("/auth", func(r chi.Router) {
		r.Post("/register", authH.Register)
		r.Post("/login", authH.Login)
		r.Post("/refresh", authH.Refresh)
		r.Post("/logout", authH.Logout)

		// Google OAuth (browser-initiated — no JWT middleware)
		r.Get("/google", authH.GoogleLogin)
		r.Get("/google/callback", authH.GoogleCallback)

		// Email verification (requires access token)
		r.Group(func(r chi.Router) {
			r.Use(middleware.Authenticate(cfg.JWTSecret))
			r.Post("/verify/send", authH.SendVerificationCode)
			r.Post("/verify", authH.VerifyEmail)
		})
	})

	r.Group(func(r chi.Router) {
		r.Use(middleware.Authenticate(cfg.JWTSecret))

		r.Get("/me", userH.GetMe)
		r.Get("/me/stats", userH.GetStats)
		r.Get("/metrics", chalH.ListMetrics)

		r.Route("/challenges", func(r chi.Router) {
			r.Post("/", chalH.CreateChallenge)
			r.Get("/", chalH.ListUserChallenges)
			r.Post("/join", chalH.JoinChallenge)
			r.Get("/{id}", chalH.GetChallenge)
			r.Post("/{id}/metrics", chalH.AddMetrics)
			r.Post("/{id}/submissions", subH.Submit)
			r.Get("/{id}/submissions", subH.ListSubmissions)
			r.Post("/{id}/submissions/{subID}/media", subH.UploadMedia)
			r.Get("/{id}/leaderboard", chalH.GetLeaderboard)
			r.Patch("/{id}/status", chalH.UpdateStatus)
			r.Post("/{id}/leave", chalH.Leave)
		})
	})

	srv := &http.Server{
		Addr:         ":" + cfg.AppPort,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	cronCtx, cronCancel := context.WithCancel(context.Background())
	defer cronCancel()
	go runMissedSubmissionCron(cronCtx, subSvc, logger)

	serverErr := make(chan error, 1)
	go func() {
		logger.Info("server starting", slog.String("addr", srv.Addr))
		serverErr <- srv.ListenAndServe()
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serverErr:
		logger.Error("server error", slog.String("error", err.Error()))
	case sig := <-quit:
		logger.Info("shutdown signal received", slog.String("signal", sig.String()))
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("graceful shutdown failed", slog.String("error", err.Error()))
		os.Exit(1)
	}

	logger.Info("server stopped")
}

func runMissedSubmissionCron(ctx context.Context, svc *services.SubmissionService, logger *slog.Logger) {
	for {
		now := time.Now().UTC()
		next := time.Date(now.Year(), now.Month(), now.Day(), 0, 5, 0, 0, time.UTC)
		if !next.After(now) {
			next = next.Add(24 * time.Hour)
		}
		logger.Info("cron: next run scheduled", slog.Time("at", next))
		select {
		case <-ctx.Done():
			return
		case <-time.After(next.Sub(now)):
		}
		date := time.Now().UTC().AddDate(0, 0, -1).Format("2006-01-02")
		logger.Info("cron: processing missed submissions", slog.String("date", date))
		if err := svc.ProcessMissedSubmissions(ctx, date); err != nil {
			logger.Error("cron: missed submissions failed", slog.String("error", err.Error()))
		}
	}
}
