package auth

import (
	"context"
	"errors"
	"log"

	"connectrpc.com/connect"

	authv1 "project/internal/gen/auth/v1"
	"project/internal/gen/auth/v1/authv1connect"
)

// Handler adapts Service to the generated Connect interface.
type Handler struct {
	authv1connect.UnimplementedAuthServiceHandler
	service *Service
}

// NewHandler constructs the Connect handler for AuthService.
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RequestLogin(
	ctx context.Context,
	req *connect.Request[authv1.RequestLoginRequest],
) (*connect.Response[authv1.RequestLoginResponse], error) {
	if err := h.service.RequestLogin(ctx, req.Msg.GetEmail()); err != nil {
		log.Printf("auth: request login failed: %v", err)
		return nil, connect.NewError(connect.CodeInternal, errors.New("auth: request failed"))
	}
	return connect.NewResponse(&authv1.RequestLoginResponse{}), nil
}

// SubmitLogin issues the session as an HttpOnly Set-Cookie header (see
// cookie.go) rather than in the response body — the token never crosses
// into JavaScript-readable territory (PRD 017).
func (h *Handler) SubmitLogin(
	ctx context.Context,
	req *connect.Request[authv1.SubmitLoginRequest],
) (*connect.Response[authv1.SubmitLoginResponse], error) {
	token, err := h.service.SubmitLogin(ctx, req.Msg.GetEmail(), req.Msg.GetCode())
	if errors.Is(err, errUnauthenticated) {
		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("invalid email or code"))
	}
	if err != nil {
		log.Printf("auth: submit login failed: %v", err)
		return nil, connect.NewError(connect.CodeInternal, errors.New("auth: login failed"))
	}

	resp := connect.NewResponse(&authv1.SubmitLoginResponse{})
	setSessionCookies(resp.Header(), token, accessTokenTTL)
	return resp, nil
}

// Logout clears the session cookie server-side. It is public (see the
// allowlist in cmd/server/main.go): it needs no identity check, and it must
// still succeed for a caller whose session already expired.
func (h *Handler) Logout(
	_ context.Context,
	_ *connect.Request[authv1.LogoutRequest],
) (*connect.Response[authv1.LogoutResponse], error) {
	resp := connect.NewResponse(&authv1.LogoutResponse{})
	clearSessionCookies(resp.Header())
	return resp, nil
}
