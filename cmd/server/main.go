// Command server is the composition root: load configuration, construct
// dependencies, apply migrations, register handlers, and start listening.
// Business logic belongs in the domain packages under internal/, not here.
package main

import (
	"context"
	"io/fs"
	"log"
	"net/http"

	"connectrpc.com/connect"

	dbfs "project/db"
	"project/internal/auth"
	"project/internal/gen/auth/v1/authv1connect"
	"project/internal/gen/db"
	"project/internal/health"
	"project/internal/platform/config"
	"project/internal/platform/database"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	ctx := context.Background()

	pool, err := database.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()

	// Migrations are embedded and applied at startup, up-only. A migration
	// failure aborts startup loudly instead of running against an unknown
	// schema.
	migrations, err := fs.Sub(dbfs.Migrations, "migrations")
	if err != nil {
		return err
	}
	if err := database.Migrate(ctx, pool, migrations); err != nil {
		return err
	}

	queries := db.New(pool)

	jwtManager, err := auth.NewJWTManager(cfg.JWTSecret)
	if err != nil {
		return err
	}

	authHandler := auth.NewHandler(auth.NewService(
		auth.NewRepository(queries),
		auth.NoopLoginCodeSender{},
		jwtManager,
	))

	// Every Connect procedure is protected by default. Only the procedures
	// listed here bypass authentication — forgetting to list a new one locks
	// it, never exposes it.
	publicProcedures := map[string]bool{
		authv1connect.AuthServiceRequestLoginProcedure: true,
		authv1connect.AuthServiceSubmitLoginProcedure:  true,
	}
	withAuth := connect.WithInterceptors(auth.NewInterceptor(jwtManager, publicProcedures))

	mux := http.NewServeMux()
	mux.Handle("GET /health", health.Handler())
	registerAuth(mux, authHandler, withAuth)

	if err := registerFrontend(mux); err != nil {
		return err
	}

	addr := ":" + cfg.Port
	log.Printf("listening on %s", addr)
	return http.ListenAndServe(addr, mux)
}
