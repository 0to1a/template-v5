package config

import "testing"

// unsetEnv marks key as unset (empty) for the duration of the test,
// restored automatically by t.Setenv. An empty JWT_SECRET/DATABASE_URL is
// indistinguishable from an unset one to Load's checks, so this keeps
// every subtest hermetic regardless of the ambient environment.
func unsetEnv(t *testing.T, key string) {
	t.Helper()
	t.Setenv(key, "")
}

const validJWTSecret = "01234567890123456789012345678901" // 33 bytes, >= the required 32

// TC-010-8: table-driven coverage of Load's validation and defaulting.
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
			databaseURL: "postgres://user:pass@localhost:5432/db",
			wantErr:     true,
		},
		{
			name:        "JWT_SECRET shorter than 32 bytes fails",
			jwtSecret:   "too-short",
			databaseURL: "postgres://user:pass@localhost:5432/db",
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
			databaseURL: "postgres://user:pass@localhost:5432/db",
			wantErr:     false,
			wantPort:    "8080",
			wantMailURL: "",
			wantIsGuest: false,
		},
		{
			name:                "every field set is reflected as-is",
			jwtSecret:           validJWTSecret,
			databaseURL:         "postgres://user:pass@localhost:5432/db",
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
			databaseURL:         "postgres://user:pass@localhost:5432/db",
			isGuestRegistration: "true",
			wantErr:             false,
			wantPort:            "8080",
			wantIsGuest:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, key := range []string{"PORT", "JWT_SECRET", "DATABASE_URL", "MAIL_URL", "IS_GUEST_REGISTRATION"} {
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
