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
	return connect.NewResponse(&authv1.SubmitLoginResponse{AccessToken: token}), nil
}
