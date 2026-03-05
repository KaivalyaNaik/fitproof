package repositories

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"

	db "github.com/KaivalyaNaik/fitproof/internal/repositories/db"
)

var ErrEmailAlreadyExists = errors.New("email already exists")

type UserRepository struct {
	q *db.Queries
}

func NewUserRepository(q *db.Queries) *UserRepository {
	return &UserRepository{q: q}
}

func (r *UserRepository) CreateUser(ctx context.Context, params db.CreateUserParams) (db.User, error) {
	user, err := r.q.CreateUser(ctx, params)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return db.User{}, ErrEmailAlreadyExists
		}
		return db.User{}, err
	}
	return user, nil
}

func (r *UserRepository) GetUserByEmail(ctx context.Context, email string) (db.User, error) {
	return r.q.GetUserByEmail(ctx, email)
}

func (r *UserRepository) GetUserByID(ctx context.Context, id uuid.UUID) (db.User, error) {
	return r.q.GetUserByID(ctx, id)
}

func (r *UserRepository) GetUserStats(ctx context.Context, userID uuid.UUID) (db.GetUserStatsRow, error) {
	return r.q.GetUserStats(ctx, userID)
}

func (r *UserRepository) CreateUserOAuth(ctx context.Context, params db.CreateUserOAuthParams) (db.User, error) {
	user, err := r.q.CreateUserOAuth(ctx, params)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return db.User{}, ErrEmailAlreadyExists
		}
		return db.User{}, err
	}
	return user, nil
}

func (r *UserRepository) GetUserByGoogleID(ctx context.Context, googleID string) (db.User, error) {
	return r.q.GetUserByGoogleID(ctx, &googleID)
}

func (r *UserRepository) UpdateUserGoogleID(ctx context.Context, id uuid.UUID, googleID string) (db.User, error) {
	return r.q.UpdateUserGoogleID(ctx, db.UpdateUserGoogleIDParams{
		ID:       id,
		GoogleID: &googleID,
	})
}

func (r *UserRepository) SetEmailVerified(ctx context.Context, id uuid.UUID) error {
	return r.q.SetEmailVerified(ctx, id)
}
