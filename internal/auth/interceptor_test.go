package auth

import (
	"context"
	"errors"
	"testing"

	"connectrpc.com/connect"

	authv1 "project/internal/gen/auth/v1"
)

// Invariant test for secure-by-default auth: any procedure that is not in
// the public allowlist must be rejected without a valid bearer token. The
// request below carries no procedure spec, so it can never match an
// allowlist entry — exactly the situation of a newly added, unlisted
// procedure.
func TestInterceptor_DefaultDeny(t *testing.T) {
	m := newTestJWTManager(t, fixedTime)
	interceptor := NewInterceptor(m, map[string]bool{})

	next := connect.UnaryFunc(func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
		t.Fatal("handler reached without authentication")
		return nil, nil
	})

	req := connect.NewRequest(&authv1.RequestLoginRequest{})
	_, err := interceptor.WrapUnary(next)(context.Background(), req)

	var connectErr *connect.Error
	if !errors.As(err, &connectErr) || connectErr.Code() != connect.CodeUnauthenticated {
		t.Fatalf("expected CodeUnauthenticated, got %v", err)
	}
}

func TestInterceptor_ValidTokenAttachesPrincipal(t *testing.T) {
	m := newTestJWTManager(t, fixedTime)
	interceptor := NewInterceptor(m, map[string]bool{})

	token, err := m.Issue(testUUID)
	if err != nil {
		t.Fatal(err)
	}

	var got Principal
	next := connect.UnaryFunc(func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
		principal, err := RequirePrincipal(ctx)
		if err != nil {
			return nil, err
		}
		got = principal
		return connect.NewResponse(&authv1.RequestLoginResponse{}), nil
	})

	req := connect.NewRequest(&authv1.RequestLoginRequest{})
	req.Header().Set("Cookie", sessionCookieName+"="+token)

	if _, err := interceptor.WrapUnary(next)(context.Background(), req); err != nil {
		t.Fatal(err)
	}
	if got.PublicUUID != testUUID {
		t.Fatalf("principal = %s, want %s", got.PublicUUID, testUUID)
	}
}

func TestInterceptor_MalformedTokenRejected(t *testing.T) {
	m := newTestJWTManager(t, fixedTime)
	interceptor := NewInterceptor(m, map[string]bool{})

	next := connect.UnaryFunc(func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
		t.Fatal("handler reached with a malformed token")
		return nil, nil
	})

	req := connect.NewRequest(&authv1.RequestLoginRequest{})
	req.Header().Set("Cookie", sessionCookieName+"=not-a-jwt")

	if _, err := interceptor.WrapUnary(next)(context.Background(), req); err == nil {
		t.Fatal("malformed token was accepted")
	}
}

// TC-017-2 (interceptor side): no Authorization header is honored anymore —
// the session lives only in the cookie.
func TestInterceptor_AuthorizationHeaderIgnored(t *testing.T) {
	m := newTestJWTManager(t, fixedTime)
	interceptor := NewInterceptor(m, map[string]bool{})

	token, err := m.Issue(testUUID)
	if err != nil {
		t.Fatal(err)
	}

	next := connect.UnaryFunc(func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
		t.Fatal("handler reached without a session cookie")
		return nil, nil
	})

	req := connect.NewRequest(&authv1.RequestLoginRequest{})
	req.Header().Set("Authorization", "Bearer "+token)

	if _, err := interceptor.WrapUnary(next)(context.Background(), req); err == nil {
		t.Fatal("a bare Authorization header was accepted without a session cookie")
	}
}
