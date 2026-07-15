package auth

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	"project/internal/gen/db"
)

// ErrUserNotFound indicates no active user matched the lookup.
var ErrUserNotFound = errors.New("auth: user not found")

// User is the subset of user data the auth domain needs, decoupled from the
// generated sqlc row type.
type User struct {
	PublicUUID string
	Email      string
}

// Repository is the persistence boundary for authentication, so tests can
// substitute a fake without a real database.
type Repository interface {
	GetActiveUserByEmail(ctx context.Context, normalizedEmail string) (User, error)
	CreateUser(ctx context.Context, normalizedEmail string) (User, error)
}

type repository struct {
	queries *db.Queries
}

// NewRepository builds the default, database-backed Repository.
func NewRepository(queries *db.Queries) Repository {
	return &repository{queries: queries}
}

func (r *repository) GetActiveUserByEmail(ctx context.Context, normalizedEmail string) (User, error) {
	row, err := r.queries.GetActiveUserByEmail(ctx, normalizedEmail)
	if errors.Is(err, pgx.ErrNoRows) {
		return User{}, ErrUserNotFound
	}
	if err != nil {
		return User{}, err
	}
	return User{
		PublicUUID: row.PublicUuid.String(),
		Email:      row.Email,
	}, nil
}

func (r *repository) CreateUser(ctx context.Context, normalizedEmail string) (User, error) {
	row, err := r.queries.CreateUser(ctx, normalizedEmail)
	if err != nil {
		return User{}, err
	}
	return User{
		PublicUUID: row.PublicUuid.String(),
		Email:      row.Email,
	}, nil
}
