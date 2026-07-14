package auth

import "context"

// OTPDelivery sends a login code to a user. Implementations must never log
// the code.
type OTPDelivery interface {
	SendLoginCode(ctx context.Context, email, code string) error
}

// NoopOTPDelivery discards the code. It exists so the server can run before
// a real delivery provider (email/SMS) is wired up; building that provider
// framework is out of scope for this template.
type NoopOTPDelivery struct{}

func (NoopOTPDelivery) SendLoginCode(context.Context, string, string) error {
	return nil
}
