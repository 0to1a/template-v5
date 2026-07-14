package auth

import (
	"bytes"
	"context"
	_ "embed"
	"html/template"
	"log"
)

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

//go:embed otp.html
var otpTemplateSource string

var otpTemplate = template.Must(template.New("otp").Parse(otpTemplateSource))

// EmailSender sends a single rendered HTML email. Implemented by
// internal/mail.SMTPSender; kept as an interface here so this package does
// not depend on SMTP specifics.
type EmailSender interface {
	Send(ctx context.Context, to, subject, html string) error
}

// EmailLoginCodeSender sends login codes as HTML email, rendered from the
// embedded otp.html template.
type EmailLoginCodeSender struct {
	sender EmailSender
}

// NewEmailLoginCodeSender wraps an EmailSender to deliver login codes.
func NewEmailLoginCodeSender(sender EmailSender) *EmailLoginCodeSender {
	return &EmailLoginCodeSender{sender: sender}
}

// SendLoginCode renders otp.html with code and sends it. Failures are logged
// here — never the code itself — because the caller (Service.RequestLogin)
// intentionally discards this error to avoid an account-existence oracle.
func (e *EmailLoginCodeSender) SendLoginCode(ctx context.Context, email, code string) error {
	var body bytes.Buffer
	if err := otpTemplate.Execute(&body, struct{ Code string }{Code: code}); err != nil {
		log.Printf("auth: rendering login code email failed: %v", err)
		return err
	}

	if err := e.sender.Send(ctx, email, "Your login code", body.String()); err != nil {
		log.Printf("auth: sending login code email failed: %v", err)
		return err
	}
	return nil
}
