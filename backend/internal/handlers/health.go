package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/KaivalyaNaik/fitproof/pkg/respond"
)

func HealthHandler(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		dbStatus := "ok"
		httpStatus := http.StatusOK

		if err := pool.Ping(ctx); err != nil {
			dbStatus = "unreachable"
			httpStatus = http.StatusServiceUnavailable
		}

		respond.JSON(w, httpStatus, map[string]string{
			"status":   "ok",
			"database": dbStatus,
			"time":     time.Now().UTC().Format(time.RFC3339),
		})
	}
}
