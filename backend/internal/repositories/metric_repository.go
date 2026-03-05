package repositories

import (
	"context"

	"github.com/google/uuid"

	db "github.com/KaivalyaNaik/fitproof/internal/repositories/db"
)

type MetricRepository struct {
	q *db.Queries
}

func NewMetricRepository(q *db.Queries) *MetricRepository {
	return &MetricRepository{q: q}
}

func (r *MetricRepository) ListMetrics(ctx context.Context) ([]db.Metric, error) {
	return r.q.ListMetrics(ctx)
}

func (r *MetricRepository) GetMetricByID(ctx context.Context, id uuid.UUID) (db.Metric, error) {
	return r.q.GetMetricByID(ctx, id)
}
