package main

import (
	"net/http"
	"testing"
)

// TC-009-1: the production *http.Server has every connection timeout set.
func TestNewHTTPServer_TimeoutsAreNonZero_TC009_1(t *testing.T) {
	srv := newHTTPServer(":0", http.NotFoundHandler())

	if srv.ReadHeaderTimeout <= 0 {
		t.Error("ReadHeaderTimeout must be greater than zero")
	}
	if srv.ReadTimeout <= 0 {
		t.Error("ReadTimeout must be greater than zero")
	}
	if srv.WriteTimeout <= 0 {
		t.Error("WriteTimeout must be greater than zero")
	}
	if srv.IdleTimeout <= 0 {
		t.Error("IdleTimeout must be greater than zero")
	}
}
