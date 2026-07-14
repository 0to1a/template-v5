package database

import (
	"context"
	"fmt"
	"io/fs"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	goosedb "github.com/pressly/goose/v3/database"
)

// Migrate applies all pending migrations from migrations (a filesystem whose
// root contains the .sql files) at server startup. Up only: this never rolls
// back, resets, or drops anything. A failure here must abort startup — the
// schema the binary was built against is the schema it runs against.
func Migrate(ctx context.Context, pool *pgxpool.Pool, migrations fs.FS) error {
	sqldb := stdlib.OpenDBFromPool(pool)
	defer sqldb.Close()

	provider, err := goose.NewProvider(goosedb.DialectPostgres, sqldb, migrations)
	if err != nil {
		return fmt.Errorf("database: creating migration provider: %w", err)
	}

	if _, err := provider.Up(ctx); err != nil {
		return fmt.Errorf("database: applying migrations: %w", err)
	}
	return nil
}
