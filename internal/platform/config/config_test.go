package config

import (
	"fmt"
	"strings"
	"testing"
)

const (
	validJWTSecret   = "0123456789abcdef0123456789abcdef"
	validDatabaseURL = "postgres://user:s3cret-password@localhost:5432/app?sslmode=disable"
)

func setValidBaseEnv(t *testing.T) {
	t.Helper()
	t.Setenv("JWT_SECRET", validJWTSecret)
	t.Setenv("DATABASE_URL", validDatabaseURL)
	t.Setenv("PORT", "")
	t.Setenv("MAIL_URL", "")
	t.Setenv("IS_GUEST_REGISTRATION", "")
}

// TC-011-1: a non-numeric PORT is rejected.
func TestLoad_RejectsNonNumericPort_TC011_1(t *testing.T) {
	setValidBaseEnv(t)
	t.Setenv("PORT", "abc")

	_, err := Load()
	if err == nil {
		t.Fatal("expected an error for a non-numeric PORT")
	}
	if !strings.Contains(err.Error(), "PORT") {
		t.Fatalf("error should name PORT, got: %v", err)
	}
}

// TC-011-2: an out-of-range PORT is rejected.
func TestLoad_RejectsOutOfRangePort_TC011_2(t *testing.T) {
	setValidBaseEnv(t)
	t.Setenv("PORT", "70000")

	_, err := Load()
	if err == nil {
		t.Fatal("expected an error for an out-of-range PORT")
	}
	if !strings.Contains(err.Error(), "PORT") {
		t.Fatalf("error should name PORT, got: %v", err)
	}
}

// TC-011-3: a malformed DATABASE_URL is rejected without leaking its value.
func TestLoad_RejectsMalformedDatabaseURL_TC011_3(t *testing.T) {
	setValidBaseEnv(t)
	const malformed = "not-a-url"
	t.Setenv("DATABASE_URL", malformed)

	_, err := Load()
	if err == nil {
		t.Fatal("expected an error for a malformed DATABASE_URL")
	}
	if strings.Contains(err.Error(), malformed) {
		t.Fatalf("error must not echo the raw DATABASE_URL value, got: %v", err)
	}
}

// TC-011-4: SafeFields never includes the JWT secret or the raw database URL.
func TestConfig_SafeFields_NeverLeaksSecrets_TC011_4(t *testing.T) {
	setValidBaseEnv(t)
	t.Setenv("PORT", "9090")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	fields := cfg.SafeFields()
	serialized := fmt.Sprint(fields)

	if strings.Contains(serialized, validJWTSecret) {
		t.Fatal("SafeFields leaked the JWT secret")
	}
	if strings.Contains(serialized, validDatabaseURL) || strings.Contains(serialized, "s3cret-password") {
		t.Fatal("SafeFields leaked the raw DATABASE_URL")
	}
	if fields["port"] != "9090" {
		t.Fatalf("SafeFields[port] = %v, want 9090", fields["port"])
	}
}

// Regression: existing fail-closed behavior for JWT_SECRET/DATABASE_URL is
// unchanged by this PRD's new validation.
func TestLoad_StillRejectsShortJWTSecret(t *testing.T) {
	setValidBaseEnv(t)
	t.Setenv("JWT_SECRET", "too-short")

	if _, err := Load(); err == nil {
		t.Fatal("expected an error for a JWT_SECRET shorter than 32 bytes")
	}
}

func TestLoad_StillRejectsMissingDatabaseURL(t *testing.T) {
	setValidBaseEnv(t)
	t.Setenv("DATABASE_URL", "")

	if _, err := Load(); err == nil {
		t.Fatal("expected an error for a missing DATABASE_URL")
	}
}
