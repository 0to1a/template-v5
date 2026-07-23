// Package database owns the PostgreSQL connection pool.
package database

import (
	"context"
	"fmt"
	"log"
	"time"

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

// waitReadyAttempts and waitReadyDelay bound how long WaitReady tolerates a
// database that is still starting up (e.g. Railway's Postgres reporting
// SQLSTATE 57P03 right after a cold start) before giving up.
const (
	waitReadyAttempts = 10
	waitReadyDelay    = 15 * time.Second
)

// WaitReady blocks until PostgreSQL accepts connections, retrying up to
// waitReadyAttempts times with a waitReadyDelay pause in between. Managed
// Postgres instances (Railway included) can take a few seconds to come back
// up after sleeping, and without this a cold start aborts the whole server
// on the first ping instead of waiting it out.
func WaitReady(ctx context.Context, pool *pgxpool.Pool) error {
	return waitReady(ctx, waitReadyAttempts, waitReadyDelay, pool.Ping)
}

// waitReady holds the retry loop itself, decoupled from *pgxpool.Pool so it
// can be exercised in tests without a live database.
func waitReady(ctx context.Context, attempts int, delay time.Duration, ping func(context.Context) error) error {
	var err error
	for attempt := 1; attempt <= attempts; attempt++ {
		if err = ping(ctx); err == nil {
			return nil
		}
		if attempt == attempts {
			break
		}
		log.Printf("database: not ready yet (attempt %d/%d): %v", attempt, attempts, err)
		select {
		case <-time.After(delay):
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return fmt.Errorf("database: waiting for ready: %w", err)
}
