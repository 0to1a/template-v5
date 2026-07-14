package main

import (
	"net/http"

	"connectrpc.com/connect"

	"project/internal/auth"
	"project/internal/gen/auth/v1/authv1connect"
)

// registerAuth mounts the AuthService Connect handler. One file per domain
// keeps the composition root uniform: a new domain adds its own
// register_<domain>.go and one call in main.go — never a second registry.
func registerAuth(mux *http.ServeMux, handler *auth.Handler, opts ...connect.HandlerOption) {
	mux.Handle(authv1connect.NewAuthServiceHandler(handler, opts...))
}
