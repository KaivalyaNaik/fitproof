package repositories

import (
	"context"

	"github.com/google/uuid"

	db "github.com/KaivalyaNaik/fitproof/internal/repositories/db"
)

type TokenRepository struct {
	q *db.Queries
}

func NewTokenRepository(q *db.Queries) *TokenRepository {
	return &TokenRepository{q: q}
}

func (r *TokenRepository) CreateRefreshToken(ctx context.Context, params db.CreateRefreshTokenParams) (db.RefreshToken, error) {
	return r.q.CreateRefreshToken(ctx, params)
}

func (r *TokenRepository) GetRefreshTokenByHash(ctx context.Context, hash string) (db.RefreshToken, error) {
	return r.q.GetRefreshTokenByHash(ctx, hash)
}

func (r *TokenRepository) RevokeRefreshToken(ctx context.Context, id uuid.UUID) error {
	return r.q.RevokeRefreshToken(ctx, id)
}

func (r *TokenRepository) RevokeDeviceTokens(ctx context.Context, params db.RevokeDeviceTokensParams) error {
	return r.q.RevokeDeviceTokens(ctx, params)
}
