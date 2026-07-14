package auth

import "context"

// LoginCodeSender sends a login code to a user. Implementations must never
// log the code.
type LoginCodeSender interface {
	SendLoginCode(ctx context.Context, email, code string) error
}

// NoopLoginCodeSender discards the code. It exists so the server can run
// before a real delivery provider (email/SMS) is wired up; building that
// provider framework is out of scope for this template.
type NoopLoginCodeSender struct{}

func (NoopLoginCodeSender) SendLoginCode(context.Context, string, string) error {
	return nil
}
