// Package server hosts HTTP/SPA concerns shared across the composition root.
package server

import (
	"io/fs"
	"net/http"
	"path"
	"strings"
)

// NewSPAHandler returns an http.Handler that serves a built single-page
// application out of distFS. Existing files are served as-is. A GET/HEAD
// request that is not an existing file and that looks like HTML navigation
// falls back to index.html. Every other miss returns 404, so unmatched RPC
// requests are never rewritten into HTML.
//
// Register this handler after all RPC handlers: it only ever receives
// requests that no more specific route claimed.
func NewSPAHandler(distFS fs.FS) (http.Handler, error) {
	index, err := fs.ReadFile(distFS, "index.html")
	if err != nil {
		return nil, err
	}

	fileServer := http.FileServer(http.FS(distFS))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			http.NotFound(w, r)
			return
		}

		cleaned := strings.TrimPrefix(path.Clean("/"+r.URL.Path), "/")
		if cleaned == "" {
			cleaned = "index.html"
		}

		if info, err := fs.Stat(distFS, cleaned); err == nil && !info.IsDir() {
			fileServer.ServeHTTP(w, r)
			return
		}

		if isHTMLNavigation(r) {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(http.StatusOK)
			if r.Method == http.MethodGet {
				_, _ = w.Write(index)
			}
			return
		}

		http.NotFound(w, r)
	}), nil
}

func isHTMLNavigation(r *http.Request) bool {
	return strings.Contains(r.Header.Get("Accept"), "text/html")
}
