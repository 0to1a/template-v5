// Package database owns the PostgreSQL connection pool.
package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Connect constructs a pooled connection to PostgreSQL. This does not dial
// the database: pgxpool connects lazily on first use, so the health RPC and
// server startup never depend on PostgreSQL being reachable. Callers are
// responsible for closing the returned pool.
func Connect(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("database: creating pool: %w", err)
	}
	return pool, nil
}
