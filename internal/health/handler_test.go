package health

import (
	"net/http/httptest"
	"testing"
)

// TC-001-1: /health answers without authentication and without any database
// dependency — the handler is constructed with no dependencies at all.
func TestHealth_TC001_1(t *testing.T) {
	req := httptest.NewRequest("GET", "/health", nil)
	rec := httptest.NewRecorder()

	Handler().ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
	if got := rec.Body.String(); got != `{"status":"ok"}` {
		t.Fatalf("unexpected body: %s", got)
	}
}
