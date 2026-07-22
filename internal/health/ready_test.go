package health

import (
	"context"
	"errors"
	"net/http/httptest"
	"testing"
	"time"
)

type fakePinger struct {
	err error
}

func (p fakePinger) Ping(context.Context) error { return p.err }

// TC-012-1: liveness is unaffected by anything database-related.
func TestHealth_LivenessIgnoresDatabase_TC012_1(t *testing.T) {
	req := httptest.NewRequest("GET", "/health", nil)
	rec := httptest.NewRecorder()

	Handler().ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
}

// TC-012-2: readiness succeeds when the pinger succeeds.
func TestReadyHandler_Reachable_TC012_2(t *testing.T) {
	req := httptest.NewRequest("GET", "/health/ready", nil)
	rec := httptest.NewRecorder()

	ReadyHandler(fakePinger{err: nil}, time.Second).ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
	if got := rec.Body.String(); got != `{"status":"ok"}` {
		t.Fatalf("unexpected body: %s", got)
	}
}

// TC-012-3: readiness fails, with a distinct body, when the pinger fails.
func TestReadyHandler_Unreachable_TC012_3(t *testing.T) {
	req := httptest.NewRequest("GET", "/health/ready", nil)
	rec := httptest.NewRecorder()

	ReadyHandler(fakePinger{err: errors.New("connection refused")}, time.Second).ServeHTTP(rec, req)

	if rec.Code/100 == 2 {
		t.Fatalf("expected a non-2xx status, got %d", rec.Code)
	}
	if got := rec.Body.String(); got != `{"status":"unavailable"}` {
		t.Fatalf("unexpected body: %s", got)
	}
}
