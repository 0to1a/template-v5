package observability

import (
	"context"
	"crypto/rand"
	"encoding/hex"
)

// HeaderRequestID is the header a caller may set to propagate its own
// correlation ID, and the header every response echoes it back on.
const HeaderRequestID = "X-Request-Id"

type correlationIDKey struct{}

// NewCorrelationID generates a fresh, unguessable correlation ID. It is an
// opaque trace label, never a security token, so a non-cryptographic
// encoding (hex of random bytes) is enough.
func NewCorrelationID() string {
	var b [16]byte
	_, _ = rand.Read(b[:])
	return hex.EncodeToString(b[:])
}

// WithCorrelationID attaches id to ctx.
func WithCorrelationID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, correlationIDKey{}, id)
}

// CorrelationIDFromContext retrieves the ID attached by WithCorrelationID.
func CorrelationIDFromContext(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(correlationIDKey{}).(string)
	return id, ok
}
