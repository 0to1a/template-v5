package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/binary"
	"fmt"
	"strings"
	"time"
)

const (
	totpDigits = 6
	totpPeriod = 300 * time.Second

	// localAdminEmail is the one development account that accepts the fixed
	// code below, and only when developmentMode is true (see isLocalAdmin).
	// The match is exact — other @localhost addresses still go through TOTP
	// — so the static-credential surface stays as small as possible.
	localAdminEmail = "admin@localhost"
	localAdminCode  = "123456"

	totpSecretMessagePrefix = "template-v5:login-totp:v1:"
)

// normalizeEmail applies the single normalization rule used across lookups
// and storage: trim whitespace, lowercase.
func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

// isLocalAdmin reports whether a normalized email is the seeded development
// account that uses the fixed login code — and only when developmentMode is
// true. Outside development, admin@localhost has no special case and goes
// through TOTP like every other account (see PRD 014).
func isLocalAdmin(normalizedEmail string, developmentMode bool) bool {
	return developmentMode && normalizedEmail == localAdminEmail
}

// deriveTOTPSecret produces a per-user TOTP secret from the server's JWT
// signing secret and the user's public identity. No per-user secret is ever
// stored: rotating JWT_SECRET invalidates every current OTP.
func deriveTOTPSecret(jwtSecret []byte, publicUUID string) []byte {
	mac := hmac.New(sha256.New, jwtSecret)
	mac.Write([]byte(totpSecretMessagePrefix + publicUUID))
	return mac.Sum(nil)
}

// hotp implements RFC 4226 HMAC-based one-time passwords with SHA-256.
func hotp(secret []byte, counter uint64) string {
	var counterBytes [8]byte
	binary.BigEndian.PutUint64(counterBytes[:], counter)

	mac := hmac.New(sha256.New, secret)
	mac.Write(counterBytes[:])
	sum := mac.Sum(nil)

	offset := sum[len(sum)-1] & 0x0F
	binCode := (uint32(sum[offset]&0x7F) << 24) |
		(uint32(sum[offset+1]) << 16) |
		(uint32(sum[offset+2]) << 8) |
		uint32(sum[offset+3])

	mod := uint32(1)
	for range totpDigits {
		mod *= 10
	}
	return fmt.Sprintf("%0*d", totpDigits, binCode%mod)
}

// currentTOTP returns the code for the step containing now, per RFC 6238.
// Skew is fixed at 0: only the current step is ever accepted, so a code is
// valid for its whole 5-minute step and can be replayed within it. This is a
// documented limitation, not a bug; closing it needs a replay-tracking store
// this template does not implement.
func currentTOTP(secret []byte, now time.Time) string {
	counter := uint64(now.Unix()) / uint64(totpPeriod.Seconds())
	return hotp(secret, counter)
}

// generateCode returns the login code for this user at time now: the fixed
// code for the seeded local admin when developmentMode is true, otherwise
// the current TOTP step. Time is a parameter (not time.Now) so tests are
// deterministic without sleeping.
func generateCode(jwtSecret []byte, publicUUID, normalizedEmail string, now time.Time, developmentMode bool) string {
	if isLocalAdmin(normalizedEmail, developmentMode) {
		return localAdminCode
	}
	secret := deriveTOTPSecret(jwtSecret, publicUUID)
	return currentTOTP(secret, now)
}

// verifyCode checks code against the expected value for this user at time
// now, in constant time.
func verifyCode(jwtSecret []byte, publicUUID, normalizedEmail, code string, now time.Time, developmentMode bool) bool {
	expected := generateCode(jwtSecret, publicUUID, normalizedEmail, now, developmentMode)
	return subtle.ConstantTimeCompare([]byte(expected), []byte(code)) == 1
}
