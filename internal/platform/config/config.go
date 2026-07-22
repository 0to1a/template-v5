// Package config loads process configuration from the environment.
package config

import (
	"fmt"
	"net/url"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config holds all runtime configuration for the server.
type Config struct {
	Port                string
	DatabaseURL         string
	JWTSecret           string
	MailURL             string
	IsGuestRegistration bool
}

// SafeFields returns the subset of Config that is safe to log at startup:
// no secret ever appears here. It exists so nothing has to (and nothing
// should) print the Config struct itself, which carries JWTSecret and a
// DatabaseURL that may embed a password.
func (c *Config) SafeFields() map[string]any {
	return map[string]any{
		"port":                  c.Port,
		"mail_delivery_enabled": c.MailURL != "",
		"is_guest_registration": c.IsGuestRegistration,
	}
}

// Load reads configuration from the environment, optionally populated by a
// local .env file. A missing .env file is not an error, and .env values never
// override variables already set in the process environment.
func Load() (*Config, error) {
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("config: loading .env: %w", err)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	if err := validatePort(port); err != nil {
		return nil, err
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if len(jwtSecret) < 32 {
		return nil, fmt.Errorf("config: JWT_SECRET must be set to at least 32 bytes")
	}

	// Required because the server applies embedded migrations at startup;
	// there is no mode that runs without a database.
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return nil, fmt.Errorf("config: DATABASE_URL must be set")
	}
	if err := validateDatabaseURL(databaseURL); err != nil {
		return nil, err
	}

	// Optional: when unset, login codes are not emailed (see
	// auth.NoopLoginCodeSender). Format parsed and validated in internal/mail.
	mailURL := os.Getenv("MAIL_URL")

	// Optional: when unset or anything other than "1", RequestLogin never
	// creates an account (today's behavior, see internal/auth.Service).
	isGuestRegistration := os.Getenv("IS_GUEST_REGISTRATION") == "1"

	return &Config{
		Port:                port,
		DatabaseURL:         databaseURL,
		JWTSecret:           jwtSecret,
		MailURL:             mailURL,
		IsGuestRegistration: isGuestRegistration,
	}, nil
}

// validatePort rejects anything that is not a valid TCP port number. PORT is
// not a secret, so its own value is safe to include in the error.
func validatePort(port string) error {
	n, err := strconv.Atoi(port)
	if err != nil || n < 1 || n > 65535 {
		return fmt.Errorf("config: PORT must be a valid TCP port number (1-65535), got %q", port)
	}
	return nil
}

// validateDatabaseURL rejects a DATABASE_URL that isn't a well-formed
// postgres URL. The error deliberately never includes raw, which may embed
// a password.
func validateDatabaseURL(raw string) error {
	u, err := url.Parse(raw)
	if err != nil || (u.Scheme != "postgres" && u.Scheme != "postgresql") || u.Hostname() == "" {
		return fmt.Errorf("config: DATABASE_URL must be a valid postgres:// or postgresql:// URL")
	}
	return nil
}
