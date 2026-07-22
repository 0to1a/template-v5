package main

import (
	"net/http"
	"time"
)

const (
	// readHeaderTimeout bounds how long a client may take to send request
	// headers, closing slow-loris-style connections.
	readHeaderTimeout = 5 * time.Second
	// readTimeout bounds the whole request (headers + body).
	readTimeout = 15 * time.Second
	// writeTimeout bounds how long a handler may take to write a response.
	writeTimeout = 30 * time.Second
	// idleTimeout bounds how long a keep-alive connection may sit idle
	// between requests.
	idleTimeout = 60 * time.Second
	// shutdownTimeout bounds how long SIGTERM/SIGINT wait for in-flight
	// requests to finish before the listener is force-closed.
	shutdownTimeout = 20 * time.Second
)

// newHTTPServer builds the production *http.Server. Every timeout is set
// explicitly: the zero-value http.Server has none, which lets a slow or
// malicious client hold a connection open indefinitely.
func newHTTPServer(addr string, handler http.Handler) *http.Server {
	return &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: readHeaderTimeout,
		ReadTimeout:       readTimeout,
		WriteTimeout:      writeTimeout,
		IdleTimeout:       idleTimeout,
	}
}
