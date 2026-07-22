// Package observability defines a small, vendor-neutral logging interface
// and the request-correlation middleware built on top of it. Nothing here
// depends on a specific observability vendor's SDK; a real one (Sentry,
// Datadog, etc.) can implement Logger later without touching call sites.
package observability

import (
	"context"
	"io"
	"log/slog"
)

// Logger is the one logging interface the rest of this template depends on.
// Its shape mirrors log/slog's context-aware methods so the default
// implementation is a thin wrapper, but callers never import log/slog
// directly — that keeps a future vendor swap to this one file.
type Logger interface {
	Info(ctx context.Context, msg string, args ...any)
	Error(ctx context.Context, msg string, args ...any)
}

// slogLogger is the default Logger, writing structured JSON lines.
type slogLogger struct {
	logger *slog.Logger
}

// NewLogger returns the default Logger, writing JSON lines to w.
func NewLogger(w io.Writer) Logger {
	return &slogLogger{logger: slog.New(slog.NewJSONHandler(w, nil))}
}

func (l *slogLogger) Info(ctx context.Context, msg string, args ...any) {
	l.logger.InfoContext(ctx, msg, args...)
}

func (l *slogLogger) Error(ctx context.Context, msg string, args ...any) {
	l.logger.ErrorContext(ctx, msg, args...)
}
