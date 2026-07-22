package auth

import (
	"net/http"
	"time"
)

const (
	// sessionCookieName carries the signed JWT. HttpOnly: JavaScript in the
	// browser can never read it, closing the risk PRD 017 exists for.
	sessionCookieName = "template_v5_session"

	// authedCookieName is a companion, non-HttpOnly indicator with no
	// secret value, so web/src/lib/auth.ts can answer "am I logged in"
	// without ever holding the session token itself.
	authedCookieName = "template_v5_authed"
)

// setSessionCookies attaches the session and authed-indicator cookies to a
// response after a successful login. Both are Secure and
// SameSite=Strict — per the owner's approval on ALV-14, SameSite alone is
// this template's CSRF defense, with no separate double-submit token.
func setSessionCookies(header http.Header, token string, ttl time.Duration) {
	maxAge := int(ttl.Seconds())
	header.Add("Set-Cookie", sessionCookie(sessionCookieName, token, true, maxAge).String())
	header.Add("Set-Cookie", sessionCookie(authedCookieName, "1", false, maxAge).String())
}

// clearSessionCookies expires both cookies immediately, used by Logout.
func clearSessionCookies(header http.Header) {
	header.Add("Set-Cookie", sessionCookie(sessionCookieName, "", true, -1).String())
	header.Add("Set-Cookie", sessionCookie(authedCookieName, "", false, -1).String())
}

func sessionCookie(name, value string, httpOnly bool, maxAge int) *http.Cookie {
	return &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		MaxAge:   maxAge,
		HttpOnly: httpOnly,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	}
}

// sessionTokenFromCookie extracts the session token from an incoming
// request's Cookie header, if present.
func sessionTokenFromCookie(header http.Header) (string, bool) {
	req := http.Request{Header: header}
	cookie, err := req.Cookie(sessionCookieName)
	if err != nil {
		return "", false
	}
	return cookie.Value, true
}
