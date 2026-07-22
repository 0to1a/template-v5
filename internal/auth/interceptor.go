package auth

import (
	"context"
	"errors"

	"connectrpc.com/connect"
)

var errMissingOrInvalidToken = errors.New("unauthenticated")

// NewInterceptor builds the global authentication interceptor. Procedures in
// publicProcedures (generated Connect procedure paths, e.g.
// "/auth.v1.AuthService/RequestLogin") bypass authentication; every other
// procedure requires a valid session cookie (see PRD 017 and cookie.go),
// whose principal is attached to the request context via WithPrincipal.
func NewInterceptor(jwtManager *JWTManager, publicProcedures map[string]bool) connect.Interceptor {
	return connect.UnaryInterceptorFunc(func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			if publicProcedures[req.Spec().Procedure] {
				return next(ctx, req)
			}

			token, ok := sessionTokenFromCookie(req.Header())
			if !ok {
				return nil, connect.NewError(connect.CodeUnauthenticated, errMissingOrInvalidToken)
			}

			principal, err := jwtManager.Parse(token)
			if err != nil {
				return nil, connect.NewError(connect.CodeUnauthenticated, errMissingOrInvalidToken)
			}

			return next(WithPrincipal(ctx, principal), req)
		}
	})
}

// RequirePrincipal extracts the authenticated principal from ctx, returning
// a ready-to-use Unauthenticated Connect error when the interceptor didn't
// attach one. Every handler that requires authentication should call this
// instead of PrincipalFromContext directly, so the unauthenticated response
// stays identical everywhere.
func RequirePrincipal(ctx context.Context) (Principal, error) {
	principal, ok := PrincipalFromContext(ctx)
	if !ok {
		return Principal{}, connect.NewError(connect.CodeUnauthenticated, errMissingOrInvalidToken)
	}
	return principal, nil
}
