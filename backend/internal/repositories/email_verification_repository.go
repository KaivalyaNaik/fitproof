package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	db "github.com/KaivalyaNaik/fitproof/internal/repositories/db"
)

type EmailVerificationRepository struct {
	q *db.Queries
}

func NewEmailVerificationRepository(q *db.Queries) *EmailVerificationRepository {
	return &EmailVerificationRepository{q: q}
}

func (r *EmailVerificationRepository) Create(ctx context.Context, userID uuid.UUID, codeHash string, expiresAt pgtype.Timestamptz) (db.EmailVerificationToken, error) {
	return r.q.CreateEmailVerificationToken(ctx, db.CreateEmailVerificationTokenParams{
		UserID:    userID,
		Code:      codeHash,
		ExpiresAt: expiresAt,
	})
}

func (r *EmailVerificationRepository) GetLatest(ctx context.Context, userID uuid.UUID) (db.EmailVerificationToken, error) {
	return r.q.GetLatestTokenForUser(ctx, userID)
}

func (r *EmailVerificationRepository) MarkUsed(ctx context.Context, id uuid.UUID) error {
	return r.q.MarkTokenUsed(ctx, id)
}

func (r *EmailVerificationRepository) DeleteAll(ctx context.Context, userID uuid.UUID) error {
	return r.q.DeleteUserTokens(ctx, userID)
}
