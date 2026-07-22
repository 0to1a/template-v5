package observability

import (
	"net/http"
	"time"
)

// RequestLogging wraps next so every request carries a correlation ID (its
// own X-Request-Id header if it set one, otherwise a freshly generated ID)
// and produces exactly one structured log line. Only method, path, status,
// duration, and the correlation ID are logged — never headers or bodies —
// so bearer tokens and login codes can never leak into logs through here.
func RequestLogging(logger Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			id := r.Header.Get(HeaderRequestID)
			if id == "" {
				id = NewCorrelationID()
			}
			w.Header().Set(HeaderRequestID, id)
			ctx := WithCorrelationID(r.Context(), id)

			rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
			start := time.Now()
			next.ServeHTTP(rec, r.WithContext(ctx))
			duration := time.Since(start)

			logger.Info(ctx, "http_request",
				"request_id", id,
				"method", r.Method,
				"path", r.URL.Path,
				"status", rec.status,
				"duration_ms", duration.Milliseconds(),
			)
		})
	}
}

// statusRecorder captures the status code a handler writes, defaulting to
// 200 for handlers that never call WriteHeader explicitly.
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}
