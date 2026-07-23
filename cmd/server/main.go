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
	"project/internal/mail"
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

	// A managed Postgres instance (Railway included) can still be starting
	// up right after a cold start. Wait for it instead of letting the first
	// migration query fail startup outright.
	if err := database.WaitReady(ctx, pool); err != nil {
		return err
	}

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

	// MAIL_URL is optional. Unset means login codes are discarded, same as
	// before this feature existed; a malformed non-empty value aborts
	// startup rather than failing silently on the first login request.
	mailSender, err := mail.NewSMTPSenderFromURL(cfg.MailURL)
	if err != nil {
		return err
	}
	var loginCodeSender auth.LoginCodeSender = auth.NoopLoginCodeSender{}
	if mailSender != nil {
		loginCodeSender = auth.NewEmailLoginCodeSender(mailSender)
	}

	authHandler := auth.NewHandler(auth.NewService(
		auth.NewRepository(queries),
		loginCodeSender,
		jwtManager,
		cfg.IsGuestRegistration,
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
