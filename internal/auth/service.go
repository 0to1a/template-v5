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
	repo                Repository
	delivery            LoginCodeSender
	jwtManager          *JWTManager
	now                 func() time.Time
	isGuestRegistration bool
	throttle            *loginThrottle
}

// NewService wires the auth vertical slice. The per-user TOTP secret is
// derived from jwtManager's own signing secret, so there is exactly one
// source of truth for it. Time flows through an injected clock (overridden
// in tests) so OTP behavior is deterministic without sleeping. When
// isGuestRegistration is true, RequestLogin auto-creates an active user for
// an email that doesn't have one yet, instead of silently doing nothing.
func NewService(repo Repository, delivery LoginCodeSender, jwtManager *JWTManager, isGuestRegistration bool) *Service {
	return &Service{
		repo:                repo,
		delivery:            delivery,
		jwtManager:          jwtManager,
		now:                 time.Now,
		isGuestRegistration: isGuestRegistration,
		throttle:            newLoginThrottle(),
	}
}

// RequestLogin delivers a login code to an active user. If none exists and
// guest registration is enabled, one is created first. Either way, the
// return value never reveals whether the account already existed.
func (s *Service) RequestLogin(ctx context.Context, email string) error {
	normalized := normalizeEmail(email)

	user, err := s.repo.GetActiveUserByEmail(ctx, normalized)
	if errors.Is(err, ErrUserNotFound) {
		if !s.isGuestRegistration {
			return nil
		}
		user, err = s.repo.CreateUser(ctx, normalized)
		if err != nil {
			return err
		}
	} else if err != nil {
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
// Unknown users and invalid codes both return errUnauthenticated. Repeated
// failures for one account are throttled (see PRD 015): once an account is
// locked out, further attempts are rejected without even checking the code,
// using the same generic error as an invalid code.
func (s *Service) SubmitLogin(ctx context.Context, email, code string) (string, error) {
	normalized := normalizeEmail(email)

	user, err := s.repo.GetActiveUserByEmail(ctx, normalized)
	if errors.Is(err, ErrUserNotFound) {
		return "", errUnauthenticated
	}
	if err != nil {
		return "", err
	}

	now := s.now()

	if s.throttle.locked(user.PublicUUID, now) {
		return "", errUnauthenticated
	}

	if !s.jwtManager.verifyLoginCode(user.PublicUUID, normalized, code, now) {
		s.throttle.recordFailure(user.PublicUUID, now)
		return "", errUnauthenticated
	}
	s.throttle.reset(user.PublicUUID)

	token, err := s.jwtManager.Issue(user.PublicUUID)
	if err != nil {
		return "", err
	}
	return token, nil
}
