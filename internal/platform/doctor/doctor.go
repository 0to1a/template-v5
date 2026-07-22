// Package doctor implements the read-only checks behind `make doctor`. Every
// function here is pure: it classifies state the caller already gathered
// (a version string, a ping error) into a Check and never touches the
// filesystem, network, or a tool installation itself. Gathering that state
// is cmd/doctor's job, kept thin on purpose so the diagnosis logic is
// testable without a real Go/Bun install or a live PostgreSQL server.
package doctor

import (
	"fmt"
	"strings"

	"project/internal/mail"
)

// Status is the outcome of a single check.
type Status string

const (
	StatusOK   Status = "ok"
	StatusFail Status = "fail"
)

// Check is one diagnosed aspect of the local environment.
type Check struct {
	Name        string
	Status      Status
	Detail      string
	Remediation string // empty when Status is StatusOK
}

// OK reports whether the check passed.
func (c Check) OK() bool { return c.Status == StatusOK }

// GoVersion classifies the Go toolchain in use. reported is the output of
// `go env GOVERSION` (e.g. "go1.26.5"), or "" if the go command was not
// found. required is the project's pinned minor version (e.g. "1.26"),
// matched the same way Makefile's _check-tools does: the exact minor version
// or any patch release of it.
func GoVersion(reported, required string) Check {
	const name = "Go toolchain"
	if reported == "" {
		return Check{
			Name:        name,
			Status:      StatusFail,
			Detail:      "go was not found in PATH",
			Remediation: fmt.Sprintf("install Go %s.x from https://go.dev/dl/", required),
		}
	}

	prefix := "go" + required
	if reported == prefix || strings.HasPrefix(reported, prefix+".") {
		return Check{Name: name, Status: StatusOK, Detail: reported}
	}
	return Check{
		Name:        name,
		Status:      StatusFail,
		Detail:      fmt.Sprintf("found %s, this project requires %s.x", reported, prefix),
		Remediation: fmt.Sprintf("install Go %s.x from https://go.dev/dl/ (currently %s)", required, reported),
	}
}

// BunVersion classifies the Bun install in use. reported is the output of
// `bun --version` (e.g. "1.3.14"), or "" if bun was not found. pinned is the
// exact version this project requires (Makefile's BUN_PINNED); unlike Go,
// bun must match exactly.
func BunVersion(reported, pinned string) Check {
	const name = "Bun"
	installCmd := fmt.Sprintf(`curl -fsSL https://bun.sh/install | bash -s "bun-v%s"`, pinned)

	if reported == "" {
		return Check{
			Name:        name,
			Status:      StatusFail,
			Detail:      "bun was not found in PATH",
			Remediation: fmt.Sprintf("install bun %s: %s", pinned, installCmd),
		}
	}
	if reported == pinned {
		return Check{Name: name, Status: StatusOK, Detail: reported}
	}
	return Check{
		Name:        name,
		Status:      StatusFail,
		Detail:      fmt.Sprintf("found %s, this project pins %s", reported, pinned),
		Remediation: fmt.Sprintf("install bun %s: %s", pinned, installCmd),
	}
}

// EnvFile classifies whether a local .env file is present.
func EnvFile(exists bool) Check {
	const name = ".env file"
	if exists {
		return Check{Name: name, Status: StatusOK, Detail: "present"}
	}
	return Check{
		Name:        name,
		Status:      StatusFail,
		Detail:      "not found",
		Remediation: "copy the template: cp .env.example .env, then set DATABASE_URL and JWT_SECRET",
	}
}

// JWTSecret classifies the configured JWT_SECRET, mirroring the minimum
// length config.Load enforces at server startup.
func JWTSecret(secret string) Check {
	const name = "JWT_SECRET"
	const minBytes = 32

	if len(secret) >= minBytes {
		return Check{Name: name, Status: StatusOK, Detail: fmt.Sprintf("%d bytes", len(secret))}
	}
	detail := fmt.Sprintf("%d bytes, need at least %d", len(secret), minBytes)
	if secret == "" {
		detail = "not set"
	}
	return Check{
		Name:        name,
		Status:      StatusFail,
		Detail:      detail,
		Remediation: "set JWT_SECRET in .env to at least 32 random bytes, e.g. output of: openssl rand -hex 32",
	}
}

// DatabaseURLConfigured classifies whether DATABASE_URL is set, mirroring
// the requirement config.Load enforces at server startup.
func DatabaseURLConfigured(databaseURL string) Check {
	const name = "DATABASE_URL"
	if databaseURL != "" {
		return Check{Name: name, Status: StatusOK, Detail: "set"}
	}
	return Check{
		Name:        name,
		Status:      StatusFail,
		Detail:      "not set",
		Remediation: "set DATABASE_URL in .env, e.g. postgres://user:password@localhost:5432/template_v5?sslmode=disable",
	}
}

// MailURLConfigured classifies the optional MAIL_URL. An empty value is
// valid (login codes are simply not emailed, see internal/mail); a non-empty
// value must parse, reusing the same parser the server uses at startup so
// doctor never disagrees with what the server will do.
func MailURLConfigured(mailURL string) Check {
	const name = "MAIL_URL"
	if mailURL == "" {
		return Check{Name: name, Status: StatusOK, Detail: "not set (login codes will not be emailed)"}
	}
	if _, err := mail.ParseURL(mailURL); err != nil {
		return Check{
			Name:        name,
			Status:      StatusFail,
			Detail:      err.Error(),
			Remediation: "set MAIL_URL to smtp://user:pass@host:port, or leave it unset to disable email delivery",
		}
	}
	return Check{Name: name, Status: StatusOK, Detail: "set"}
}

// PostgresReachable classifies the result of a connection attempt the caller
// already made against DATABASE_URL. Doctor never issues writes: a nil
// pingErr means the configured server, database, and credentials all check
// out; any other error is reported verbatim so the developer can tell a
// down server apart from a credentials or database-name mismatch.
func PostgresReachable(pingErr error) Check {
	const name = "PostgreSQL"
	if pingErr == nil {
		return Check{Name: name, Status: StatusOK, Detail: "connected"}
	}
	return Check{
		Name:        name,
		Status:      StatusFail,
		Detail:      pingErr.Error(),
		Remediation: "verify PostgreSQL is running and DATABASE_URL's host/port/user/password/database match an existing role and database (this check only connects; it never creates one)",
	}
}
