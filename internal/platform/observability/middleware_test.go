package observability

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// recordingLogger captures every Info/Error call instead of writing anywhere,
// so tests can assert on exactly what would have been logged.
type recordingLogger struct {
	infoCalls []call
}

type call struct {
	msg  string
	args []any
}

func (l *recordingLogger) Info(_ context.Context, msg string, args ...any) {
	l.infoCalls = append(l.infoCalls, call{msg: msg, args: args})
}

func (l *recordingLogger) Error(_ context.Context, msg string, args ...any) {
	l.infoCalls = append(l.infoCalls, call{msg: msg, args: args})
}

func argValue(c call, key string) (any, bool) {
	for i := 0; i+1 < len(c.args); i += 2 {
		if k, ok := c.args[i].(string); ok && k == key {
			return c.args[i+1], true
		}
	}
	return nil, false
}

// TC-010-1: no incoming X-Request-Id means one is generated for the response.
func TestRequestLogging_GeneratesCorrelationID_TC010_1(t *testing.T) {
	logger := &recordingLogger{}
	handler := RequestLogging(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/widgets", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	id := rec.Header().Get(HeaderRequestID)
	if id == "" {
		t.Fatal("expected a generated X-Request-Id header, got none")
	}
}

// TC-010-2: an incoming X-Request-Id is echoed back unchanged.
func TestRequestLogging_ReusesIncomingCorrelationID_TC010_2(t *testing.T) {
	logger := &recordingLogger{}
	handler := RequestLogging(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/widgets", nil)
	req.Header.Set(HeaderRequestID, "existing-id")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if got := rec.Header().Get(HeaderRequestID); got != "existing-id" {
		t.Fatalf("X-Request-Id = %q, want %q", got, "existing-id")
	}
}

// TC-010-3: exactly one structured log line per request, carrying method,
// path, status, duration, and the correlation ID, and never the value of a
// sensitive header like Authorization.
func TestRequestLogging_LogsOnceWithoutHeaderValues_TC010_3(t *testing.T) {
	logger := &recordingLogger{}
	const secretToken = "Bearer super-secret-token"
	handler := RequestLogging(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	}))

	req := httptest.NewRequest(http.MethodPost, "/widgets", nil)
	req.Header.Set("Authorization", secretToken)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if len(logger.infoCalls) != 1 {
		t.Fatalf("expected exactly 1 log call, got %d", len(logger.infoCalls))
	}
	entry := logger.infoCalls[0]

	id := rec.Header().Get(HeaderRequestID)
	if got, ok := argValue(entry, "request_id"); !ok || got != id {
		t.Fatalf("request_id = %v, want %v", got, id)
	}
	if got, ok := argValue(entry, "method"); !ok || got != http.MethodPost {
		t.Fatalf("method = %v, want %v", got, http.MethodPost)
	}
	if got, ok := argValue(entry, "path"); !ok || got != "/widgets" {
		t.Fatalf("path = %v, want /widgets", got)
	}
	if got, ok := argValue(entry, "status"); !ok || got != http.StatusCreated {
		t.Fatalf("status = %v, want %v", got, http.StatusCreated)
	}
	if _, ok := argValue(entry, "duration_ms"); !ok {
		t.Fatal("expected a duration_ms field")
	}

	for _, arg := range entry.args {
		if s, ok := arg.(string); ok && strings.Contains(s, "super-secret-token") {
			t.Fatalf("logged argument leaked the Authorization header value: %v", arg)
		}
	}
}
