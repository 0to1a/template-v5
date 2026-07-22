package server

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"
)

// TC-010-8: NewSPAHandler requires index.html to exist in distFS.
func TestNewSPAHandler_MissingIndexHTML(t *testing.T) {
	fsys := fstest.MapFS{}

	if _, err := NewSPAHandler(fsys); err == nil {
		t.Fatal("expected an error when index.html is missing from distFS")
	}
}

// TC-010-8: table-driven coverage of NewSPAHandler's routing branches —
// existing file, SPA fallback for HTML navigation, and 404 otherwise.
func TestSPAHandler(t *testing.T) {
	fsys := fstest.MapFS{
		"index.html": &fstest.MapFile{Data: []byte("<html>index</html>")},
		"app.js":     &fstest.MapFile{Data: []byte("console.log('app')")},
	}
	handler, err := NewSPAHandler(fsys)
	if err != nil {
		t.Fatalf("NewSPAHandler: %v", err)
	}

	tests := []struct {
		name       string
		method     string
		path       string
		accept     string
		wantStatus int
		wantBody   string
	}{
		{
			name:       "an existing file is served as-is",
			method:     http.MethodGet,
			path:       "/app.js",
			accept:     "*/*",
			wantStatus: http.StatusOK,
			wantBody:   "console.log('app')",
		},
		{
			name:       "root path falls back to index.html",
			method:     http.MethodGet,
			path:       "/",
			accept:     "text/html",
			wantStatus: http.StatusOK,
			wantBody:   "<html>index</html>",
		},
		{
			name:       "unknown path with HTML navigation falls back to index.html",
			method:     http.MethodGet,
			path:       "/dashboard",
			accept:     "text/html,*/*",
			wantStatus: http.StatusOK,
			wantBody:   "<html>index</html>",
		},
		{
			name:       "HEAD on an unknown path with HTML navigation gets 200 with no body",
			method:     http.MethodHead,
			path:       "/dashboard",
			accept:     "text/html",
			wantStatus: http.StatusOK,
			wantBody:   "",
		},
		{
			name:       "unknown path without an HTML Accept 404s",
			method:     http.MethodGet,
			path:       "/dashboard",
			accept:     "application/json",
			wantStatus: http.StatusNotFound,
			wantBody:   "404 page not found\n",
		},
		{
			name:       "a non-GET/HEAD method 404s even with an HTML Accept",
			method:     http.MethodPost,
			path:       "/dashboard",
			accept:     "text/html",
			wantStatus: http.StatusNotFound,
			wantBody:   "404 page not found\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			if tt.accept != "" {
				req.Header.Set("Accept", tt.accept)
			}
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", rec.Code, tt.wantStatus)
			}
			if got := rec.Body.String(); got != tt.wantBody {
				t.Fatalf("body = %q, want %q", got, tt.wantBody)
			}
		})
	}
}
