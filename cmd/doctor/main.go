// Command doctor is the thin, impure shell behind `make doctor`: it gathers
// local environment state (tool versions, .env presence, a PostgreSQL ping)
// and hands each fact to internal/platform/doctor for classification. It
// never installs a tool, writes .env, or mutates the database — it only
// reads local state and, for the PostgreSQL check, opens a connection to
// verify reachability.
package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"

	"project/internal/platform/doctor"
)

// Pinned versions this project requires, mirroring the Makefile's
// GO_REQUIRED and BUN_PINNED so the two never drift apart silently.
const (
	goRequired = "1.26"
	bunPinned  = "1.3.14"
)

func main() {
	if !run() {
		os.Exit(1)
	}
}

func run() bool {
	_, envStatErr := os.Stat(".env")
	envExists := envStatErr == nil

	// Best-effort: populate process env from .env like config.Load does, so
	// the checks below see what the server would see. A missing .env is
	// already reported by the EnvFile check; a malformed one still lets
	// remaining checks run against whatever the process environment has.
	_ = godotenv.Load()

	checks := []doctor.Check{
		doctor.GoVersion(goVersionOutput(), goRequired),
		doctor.BunVersion(bunVersionOutput(), bunPinned),
		doctor.EnvFile(envExists),
		doctor.DatabaseURLConfigured(os.Getenv("DATABASE_URL")),
		doctor.JWTSecret(os.Getenv("JWT_SECRET")),
		doctor.MailURLConfigured(os.Getenv("MAIL_URL")),
	}

	if databaseURL := os.Getenv("DATABASE_URL"); databaseURL != "" {
		checks = append(checks, doctor.PostgresReachable(pingPostgres(databaseURL)))
	}

	allOK := true
	for _, c := range checks {
		symbol := "✓" // ✓
		if !c.OK() {
			symbol = "✗" // ✗
			allOK = false
		}
		fmt.Printf("%s %s: %s\n", symbol, c.Name, c.Detail)
		if c.Remediation != "" {
			fmt.Printf("    → %s\n", c.Remediation)
		}
	}
	return allOK
}

func goVersionOutput() string {
	out, err := exec.Command("go", "env", "GOVERSION").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func bunVersionOutput() string {
	out, err := exec.Command("bun", "--version").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// pingPostgres opens a pooled connection and pings it — a read-only
// reachability check that creates nothing. A short timeout keeps doctor
// responsive when nothing is listening at all.
func pingPostgres(databaseURL string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()

	return pool.Ping(ctx)
}
