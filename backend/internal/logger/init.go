package logger

import (
	"context"
	"log/slog"
	"os"

	"github.com/KaivalyaNaik/fitproof/internal/config"
)

type Closer func(context.Context) error

// New returns a slog.Logger and a closer. When cfg.LokiURL is blank Loki is
// disabled and the closer is a no-op — same disabled-when-blank pattern as
// pkg/email's NewSender and the R2 nil-check in cmd/server/main.go.
func New(cfg config.Config) (*slog.Logger, Closer) {
	stdout := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})

	if cfg.LokiURL == "" {
		return slog.New(stdout), func(context.Context) error { return nil }
	}

	loki := NewLokiHandler(LokiConfig{
		URL:   cfg.LokiURL,
		User:  cfg.LokiUser,
		Token: cfg.LokiToken,
		Labels: map[string]string{
			"service": "fitproof-api",
			"env":     cfg.AppEnv,
			"version": shortSHA(firstNonEmpty(os.Getenv("RENDER_GIT_COMMIT"), "dev")),
		},
		BatchInterval: cfg.LokiBatchInterval,
		BatchSize:     cfg.LokiBatchSize,
		MaxBufferSize: 10000,
	})

	return slog.New(NewFanout(stdout, loki)), loki.Close
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}

func shortSHA(s string) string {
	if len(s) > 7 {
		return s[:7]
	}
	return s
}
