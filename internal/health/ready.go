package health

import (
	"context"
	"net/http"
	"time"
)

// Pinger is the one capability the readiness handler needs from a database
// pool: prove it can be reached. *pgxpool.Pool already satisfies this, so
// main.go passes the pool directly without a wrapper.
type Pinger interface {
	Ping(ctx context.Context) error
}

// ReadyHandler responds to GET /health/ready: 200 when pinger is reachable
// within timeout, 503 otherwise. Unlike Handler (liveness), this endpoint
// depends on the database on purpose — it answers "can this process serve
// traffic", not "is this process alive".
func ReadyHandler(pinger Pinger, timeout time.Duration) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), timeout)
		defer cancel()

		w.Header().Set("Content-Type", "application/json")
		if err := pinger.Ping(ctx); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte(`{"status":"unavailable"}`))
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})
}
