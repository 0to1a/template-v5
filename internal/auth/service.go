package auth

import (
	"context"
	"errors"
	"time"
)

// errUnauthenticated is the single generic error returned for unknown users
// and invalid codes alike, so the two are indistinguishable to a caller.
var errUnauthenticated = errors.New("auth: invalid credentials")

// Service implements the login vertical slice: request login, submit login,
// issue JWTs.
type Service struct {
	repo       Repository
	delivery   LoginCodeSender
	jwtManager *JWTManager
	now        func() time.Time
}

// NewService wires the auth vertical slice. The per-user TOTP secret is
// derived from jwtManager's own signing secret, so there is exactly one
// source of truth for it. Time flows through an injected clock (overridden
// in tests) so OTP behavior is deterministic without sleeping.
func NewService(repo Repository, delivery LoginCodeSender, jwtManager *JWTManager) *Service {
	return &Service{
		repo:       repo,
		delivery:   delivery,
		jwtManager: jwtManager,
		now:        time.Now,
	}
}

// RequestLogin delivers a login code to an active user, if one exists. It
// never creates a user and never reveals, via its return value, whether the
// account exists.
func (s *Service) RequestLogin(ctx context.Context, email string) error {
	normalized := normalizeEmail(email)

	user, err := s.repo.GetActiveUserByEmail(ctx, normalized)
	if errors.Is(err, ErrUserNotFound) {
		return nil
	}
	if err != nil {
		return err
	}

	code := s.jwtManager.generateLoginCode(user.PublicUUID, normalized, s.now())
	// Delivery failure is not surfaced: an operational failure here must not
	// become an account-existence oracle. A real delivery provider should
	// log failures on its own side (never the code itself).
	_ = s.delivery.SendLoginCode(ctx, user.Email, code)
	return nil
}

// SubmitLogin validates the code and, on success, returns a signed JWT.
// Unknown users and invalid codes both return errUnauthenticated.
func (s *Service) SubmitLogin(ctx context.Context, email, code string) (string, error) {
	normalized := normalizeEmail(email)

	user, err := s.repo.GetActiveUserByEmail(ctx, normalized)
	if errors.Is(err, ErrUserNotFound) {
		return "", errUnauthenticated
	}
	if err != nil {
		return "", err
	}

	if !s.jwtManager.verifyLoginCode(user.PublicUUID, normalized, code, s.now()) {
		return "", errUnauthenticated
	}

	token, err := s.jwtManager.Issue(user.PublicUUID)
	if err != nil {
		return "", err
	}
	return token, nil
}
