// Package health serves the plain-HTTP liveness endpoint. It lives outside
// Connect on purpose: GET /health must answer without authentication and
// without touching the database, so load balancers and CI can
// probe the process itself and nothing else.
package health

import "net/http"

// Handler responds to GET /health. It takes no dependencies — adding any
// (database, config) would silently change what "healthy" means.
func Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})
}
