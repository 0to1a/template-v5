package config

import (
	"fmt"
	"strings"
	"testing"
)

// unsetEnv marks key as unset (empty) for the duration of the test,
// restored automatically by t.Setenv. An empty JWT_SECRET/DATABASE_URL is
// indistinguishable from an unset one to Load's checks, so this keeps
// every subtest hermetic regardless of the ambient environment.
func unsetEnv(t *testing.T, key string) {
	t.Helper()
	t.Setenv(key, "")
}

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
	t.Setenv("APP_ENV", "")
}

// TC-019-8: table-driven coverage of Load's validation and defaulting.
func TestLoad(t *testing.T) {
	tests := []struct {
		name                string
		jwtSecret           string
		databaseURL         string
		port                string
		mailURL             string
		isGuestRegistration string
		wantErr             bool
		wantPort            string
		wantMailURL         string
		wantIsGuest         bool
	}{
		{
			name:        "missing JWT_SECRET fails",
			jwtSecret:   "",
			databaseURL: validDatabaseURL,
			wantErr:     true,
		},
		{
			name:        "JWT_SECRET shorter than 32 bytes fails",
			jwtSecret:   "too-short",
			databaseURL: validDatabaseURL,
			wantErr:     true,
		},
		{
			name:        "missing DATABASE_URL fails",
			jwtSecret:   validJWTSecret,
			databaseURL: "",
			wantErr:     true,
		},
		{
			name:        "a minimal valid config gets PORT/MAIL_URL/IS_GUEST_REGISTRATION defaults",
			jwtSecret:   validJWTSecret,
			databaseURL: validDatabaseURL,
			wantErr:     false,
			wantPort:    "8080",
			wantMailURL: "",
			wantIsGuest: false,
		},
		{
			name:                "every field set is reflected as-is",
			jwtSecret:           validJWTSecret,
			databaseURL:         validDatabaseURL,
			port:                "9090",
			mailURL:             "smtp://user:pass@localhost:1025",
			isGuestRegistration: "1",
			wantErr:             false,
			wantPort:            "9090",
			wantMailURL:         "smtp://user:pass@localhost:1025",
			wantIsGuest:         true,
		},
		{
			name:                "IS_GUEST_REGISTRATION other than the literal \"1\" is false",
			jwtSecret:           validJWTSecret,
			databaseURL:         validDatabaseURL,
			isGuestRegistration: "true",
			wantErr:             false,
			wantPort:            "8080",
			wantIsGuest:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, key := range []string{"PORT", "JWT_SECRET", "DATABASE_URL", "MAIL_URL", "IS_GUEST_REGISTRATION", "APP_ENV"} {
				unsetEnv(t, key)
			}
			if tt.jwtSecret != "" {
				t.Setenv("JWT_SECRET", tt.jwtSecret)
			}
			if tt.databaseURL != "" {
				t.Setenv("DATABASE_URL", tt.databaseURL)
			}
			if tt.port != "" {
				t.Setenv("PORT", tt.port)
			}
			if tt.mailURL != "" {
				t.Setenv("MAIL_URL", tt.mailURL)
			}
			if tt.isGuestRegistration != "" {
				t.Setenv("IS_GUEST_REGISTRATION", tt.isGuestRegistration)
			}

			cfg, err := Load()

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected an error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("Load: %v", err)
			}
			if cfg.Port != tt.wantPort {
				t.Errorf("Port = %q, want %q", cfg.Port, tt.wantPort)
			}
			if cfg.MailURL != tt.wantMailURL {
				t.Errorf("MailURL = %q, want %q", cfg.MailURL, tt.wantMailURL)
			}
			if cfg.IsGuestRegistration != tt.wantIsGuest {
				t.Errorf("IsGuestRegistration = %v, want %v", cfg.IsGuestRegistration, tt.wantIsGuest)
			}
			if cfg.JWTSecret != tt.jwtSecret {
				t.Errorf("JWTSecret = %q, want %q", cfg.JWTSecret, tt.jwtSecret)
			}
			if cfg.DatabaseURL != tt.databaseURL {
				t.Errorf("DatabaseURL = %q, want %q", cfg.DatabaseURL, tt.databaseURL)
			}
		})
	}
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

// TC-014-3: an unrecognized APP_ENV value fails startup, naming the variable.
func TestLoad_RejectsUnrecognizedAppEnv_TC014_3(t *testing.T) {
	setValidBaseEnv(t)
	t.Setenv("APP_ENV", "staging")

	_, err := Load()
	if err == nil {
		t.Fatal("expected an error for an unrecognized APP_ENV")
	}
	if !strings.Contains(err.Error(), "APP_ENV") {
		t.Fatalf("error should name APP_ENV, got: %v", err)
	}
}

// An unset APP_ENV defaults to production: fail closed rather than silently
// running with development-only behavior (see PRD 014).
func TestLoad_UnsetAppEnvDefaultsToProduction(t *testing.T) {
	setValidBaseEnv(t)
	t.Setenv("APP_ENV", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.AppEnv != AppEnvProduction {
		t.Fatalf("AppEnv = %q, want %q", cfg.AppEnv, AppEnvProduction)
	}
}

// APP_ENV=development is accepted and recorded verbatim.
func TestLoad_AcceptsExplicitDevelopmentAppEnv(t *testing.T) {
	setValidBaseEnv(t)
	t.Setenv("APP_ENV", "development")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.AppEnv != AppEnvDevelopment {
		t.Fatalf("AppEnv = %q, want %q", cfg.AppEnv, AppEnvDevelopment)
	}
}
