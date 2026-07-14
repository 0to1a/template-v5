// Package config loads process configuration from the environment.
package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

// Config holds all runtime configuration for the server.
type Config struct {
	Port        string
	DatabaseURL string
	JWTSecret   string
	MailURL     string
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

	// Optional: when unset, login codes are not emailed (see
	// auth.NoopLoginCodeSender). Format parsed and validated in internal/mail.
	mailURL := os.Getenv("MAIL_URL")

	return &Config{
		Port:        port,
		DatabaseURL: databaseURL,
		JWTSecret:   jwtSecret,
		MailURL:     mailURL,
	}, nil
}
