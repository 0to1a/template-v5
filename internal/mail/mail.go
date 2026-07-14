// Package mail sends HTML email over SMTP, configured by a single URL.
package mail

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"net/smtp"
	"net/url"
)

// Config holds the SMTP connection details parsed from a MAIL_URL value.
type Config struct {
	Host     string
	Port     string
	Username string
	Password string
}

// ParseURL parses a MAIL_URL value of the form
// smtp://username:password@host:port. Username and password are optional;
// scheme, host, and port are required.
func ParseURL(raw string) (Config, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return Config{}, fmt.Errorf("mail: parsing MAIL_URL: %w", err)
	}
	if u.Scheme != "smtp" {
		return Config{}, fmt.Errorf("mail: MAIL_URL scheme must be smtp, got %q", u.Scheme)
	}
	host := u.Hostname()
	if host == "" {
		return Config{}, fmt.Errorf("mail: MAIL_URL is missing a host")
	}
	port := u.Port()
	if port == "" {
		return Config{}, fmt.Errorf("mail: MAIL_URL is missing a port")
	}

	cfg := Config{Host: host, Port: port}
	if u.User != nil {
		cfg.Username = u.User.Username()
		cfg.Password, _ = u.User.Password()
	}
	return cfg, nil
}

// SMTPSender sends HTML email through a single configured SMTP server.
type SMTPSender struct {
	cfg Config
}

// NewSMTPSender builds a sender from an already-parsed configuration.
func NewSMTPSender(cfg Config) *SMTPSender {
	return &SMTPSender{cfg: cfg}
}

// NewSMTPSenderFromURL parses raw as a MAIL_URL and builds a sender from it.
// An empty raw value means "no SMTP delivery configured": it returns a nil
// sender and a nil error, distinct from a malformed non-empty value, which
// returns an error.
func NewSMTPSenderFromURL(raw string) (*SMTPSender, error) {
	if raw == "" {
		return nil, nil
	}
	cfg, err := ParseURL(raw)
	if err != nil {
		return nil, err
	}
	return NewSMTPSender(cfg), nil
}

// Send delivers a single HTML email. The connection uses PLAIN auth when the
// configured username is non-empty; PLAIN auth is only attempted over a
// connection to localhost or under TLS, per net/smtp's own safeguard.
func (s *SMTPSender) Send(_ context.Context, to, subject, html string) error {
	addr := net.JoinHostPort(s.cfg.Host, s.cfg.Port)

	var auth smtp.Auth
	if s.cfg.Username != "" {
		auth = smtp.PlainAuth("", s.cfg.Username, s.cfg.Password, s.cfg.Host)
	}

	msg := buildMessage(s.cfg.Username, to, subject, html)
	return smtp.SendMail(addr, auth, s.cfg.Username, []string{to}, msg)
}

// buildMessage renders an RFC 5322 message with an HTML body. Lines are
// CRLF-terminated as SMTP requires.
func buildMessage(from, to, subject, html string) []byte {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "From: %s\r\n", from)
	fmt.Fprintf(&buf, "To: %s\r\n", to)
	fmt.Fprintf(&buf, "Subject: %s\r\n", subject)
	buf.WriteString("MIME-Version: 1.0\r\n")
	buf.WriteString("Content-Type: text/html; charset=\"UTF-8\"\r\n")
	buf.WriteString("\r\n")
	buf.WriteString(html)
	return buf.Bytes()
}
