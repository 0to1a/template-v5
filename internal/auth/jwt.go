package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// accessTokenTTL is how long an issued bearer token remains valid.
const accessTokenTTL = 24 * time.Hour

var errInvalidToken = errors.New("auth: invalid token")

// Principal is the authenticated identity carried through request context.
type Principal struct {
	PublicUUID string
}

type principalContextKey struct{}

// WithPrincipal returns a context carrying the authenticated principal. This
// is the only sanctioned way to attach a principal to a context.
func WithPrincipal(ctx context.Context, principal Principal) context.Context {
	return context.WithValue(ctx, principalContextKey{}, principal)
}

// PrincipalFromContext retrieves the principal set by WithPrincipal.
func PrincipalFromContext(ctx context.Context) (Principal, bool) {
	principal, ok := ctx.Value(principalContextKey{}).(Principal)
	return principal, ok
}

// JWTManager issues and validates HS256 bearer tokens. Construct with
// NewJWTManager; the zero value is not usable. Time flows through the
// injected clock so expiry behavior is testable without sleeping.
type JWTManager struct {
	secret []byte
	now    func() time.Time
}

// NewJWTManager requires a secret of at least 32 bytes.
func NewJWTManager(secret string) (*JWTManager, error) {
	if len(secret) < 32 {
		return nil, fmt.Errorf("auth: JWT secret must be at least 32 bytes")
	}
	return &JWTManager{secret: []byte(secret), now: time.Now}, nil
}

// generateLoginCode derives a login code with the same private key material
// used for JWT signing without exposing that material outside JWTManager.
func (m *JWTManager) generateLoginCode(publicUUID, normalizedEmail string, now time.Time) string {
	return generateCode(m.secret, publicUUID, normalizedEmail, now)
}

// verifyLoginCode verifies a login code without exposing the signing secret.
func (m *JWTManager) verifyLoginCode(publicUUID, normalizedEmail, code string, now time.Time) bool {
	return verifyCode(m.secret, publicUUID, normalizedEmail, code, now)
}

// Issue creates a bearer token whose subject is the user's public identity.
func (m *JWTManager) Issue(publicUUID string) (string, error) {
	now := m.now()
	claims := jwt.RegisteredClaims{
		Subject:   publicUUID,
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(accessTokenTTL)),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signed, err := token.SignedString(m.secret)
	if err != nil {
		return "", fmt.Errorf("auth: signing token: %w", err)
	}
	return signed, nil
}

// Parse validates the signature, algorithm (HS256 only; "none" and any other
// algorithm are rejected), and nbf, and requires both exp and a non-empty
// subject to be present.
func (m *JWTManager) Parse(tokenString string) (Principal, error) {
	var claims jwt.RegisteredClaims

	token, err := jwt.ParseWithClaims(tokenString, &claims, func(*jwt.Token) (any, error) {
		return m.secret, nil
	},
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Name}),
		jwt.WithExpirationRequired(),
		jwt.WithTimeFunc(func() time.Time { return m.now() }),
	)
	if err != nil || !token.Valid {
		return Principal{}, errInvalidToken
	}

	if claims.Subject == "" {
		return Principal{}, errInvalidToken
	}

	return Principal{PublicUUID: claims.Subject}, nil
}
