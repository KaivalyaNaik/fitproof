package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv  string
	AppPort string

	DatabaseURL    string
	DBMaxConns     int32
	DBMinConns     int32
	MigrationsPath string

	JWTSecret          string
	JWTAccessTokenTTL  time.Duration
	JWTRefreshTokenTTL time.Duration

	// Google OAuth
	GoogleClientID     string
	GoogleClientSecret string
	GoogleCallbackURL  string

	// Frontend
	FrontendURL string

	// Email (SMTP) — empty SMTP_HOST disables sending
	SMTPHost     string
	SMTPPort     int
	SMTPUsername string
	SMTPPassword string
	SMTPFrom     string

	// Email verification OTP TTL
	EmailVerificationTTL time.Duration

	// Google Drive media storage
	GoogleDriveCredentials string // raw service account JSON
	GoogleDriveFolderID    string
}

func Load() *Config {
	// Ignore error — no .env file is expected in production
	_ = godotenv.Load()

	return &Config{
		AppEnv:  getEnv("APP_ENV", "development"),
		AppPort: getEnv("APP_PORT", "8080"),

		DatabaseURL:    mustGetEnv("DATABASE_URL"),
		DBMaxConns:     int32(getEnvInt("DB_MAX_CONNECTIONS", 10)),
		DBMinConns:     int32(getEnvInt("DB_MIN_CONNECTIONS", 2)),
		MigrationsPath: getEnv("MIGRATIONS_PATH", "./migrations"),

		JWTSecret:          mustGetEnv("JWT_SECRET"),
		JWTAccessTokenTTL:  getEnvDuration("JWT_ACCESS_TOKEN_TTL", 15*time.Minute),
		JWTRefreshTokenTTL: getEnvDuration("JWT_REFRESH_TOKEN_TTL", 168*time.Hour),

		GoogleClientID:     getEnv("GOOGLE_CLIENT_ID", ""),
		GoogleClientSecret: getEnv("GOOGLE_CLIENT_SECRET", ""),
		GoogleCallbackURL:  getEnv("GOOGLE_CALLBACK_URL", "http://localhost:3000/api/auth/google/callback"),

		FrontendURL: getEnv("FRONTEND_URL", "http://localhost:3000"),

		SMTPHost:     getEnv("SMTP_HOST", ""),
		SMTPPort:     getEnvInt("SMTP_PORT", 587),
		SMTPUsername: getEnv("SMTP_USERNAME", ""),
		SMTPPassword: getEnv("SMTP_PASSWORD", ""),
		SMTPFrom:     getEnv("SMTP_FROM", "FitProof <noreply@fitproof.app>"),

		EmailVerificationTTL: getEnvDuration("EMAIL_VERIFICATION_TTL", 15*time.Minute),

		GoogleDriveCredentials: getEnv("GOOGLE_DRIVE_CREDENTIALS", ""),
		GoogleDriveFolderID:    getEnv("GOOGLE_DRIVE_FOLDER_ID", ""),
	}
}

func mustGetEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		panic(fmt.Sprintf("required environment variable %q is not set", key))
	}
	return v
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return fallback
}
