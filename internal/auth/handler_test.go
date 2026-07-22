package auth

import (
	"context"
	"net/http"
	"testing"

	"connectrpc.com/connect"

	authv1 "project/internal/gen/auth/v1"
)

// cookiesFrom reads every Set-Cookie header off a response the same way a
// browser would.
func cookiesFrom(header http.Header) []*http.Cookie {
	return (&http.Response{Header: header}).Cookies()
}

func findCookie(cookies []*http.Cookie, name string) (*http.Cookie, bool) {
	for _, c := range cookies {
		if c.Name == name {
			return c, true
		}
	}
	return nil, false
}

// TC-017-1: a successful SubmitLogin's session cookie is HttpOnly and
// Secure, and the response body carries no token.
func TestHandlerSubmitLogin_TC017_1(t *testing.T) {
	service, _ := newTestService(t, fixedTime)
	handler := NewHandler(service)

	resp, err := handler.SubmitLogin(context.Background(), connect.NewRequest(&authv1.SubmitLoginRequest{
		Email: "admin@localhost",
		Code:  "123456",
	}))
	if err != nil {
		t.Fatal(err)
	}

	cookies := cookiesFrom(resp.Header())
	session, ok := findCookie(cookies, sessionCookieName)
	if !ok {
		t.Fatal("no session cookie set")
	}
	if !session.HttpOnly {
		t.Fatal("session cookie is not HttpOnly")
	}
	if !session.Secure {
		t.Fatal("session cookie is not Secure")
	}
	if session.SameSite != http.SameSiteStrictMode {
		t.Fatalf("session cookie SameSite = %v, want Strict", session.SameSite)
	}
}

// TC-017-2: no token value is present anywhere JavaScript can read it — the
// authed indicator cookie is not HttpOnly but carries no secret, and the
// response body has no access-token field at all (enforced by the proto:
// SubmitLoginResponse is empty).
func TestHandlerSubmitLogin_TC017_2(t *testing.T) {
	service, _ := newTestService(t, fixedTime)
	handler := NewHandler(service)

	resp, err := handler.SubmitLogin(context.Background(), connect.NewRequest(&authv1.SubmitLoginRequest{
		Email: "admin@localhost",
		Code:  "123456",
	}))
	if err != nil {
		t.Fatal(err)
	}

	cookies := cookiesFrom(resp.Header())
	authed, ok := findCookie(cookies, authedCookieName)
	if !ok {
		t.Fatal("no authed indicator cookie set")
	}
	if authed.HttpOnly {
		t.Fatal("authed indicator cookie must be readable by JavaScript, not HttpOnly")
	}
	if authed.Value != "1" {
		t.Fatalf("authed indicator cookie leaks a value beyond a plain flag: %q", authed.Value)
	}
}

// TC-017-3: Logout clears both cookies (immediate expiry, empty value).
func TestHandlerLogout_TC017_3(t *testing.T) {
	handler := NewHandler(nil)

	resp, err := handler.Logout(context.Background(), connect.NewRequest(&authv1.LogoutRequest{}))
	if err != nil {
		t.Fatal(err)
	}

	cookies := cookiesFrom(resp.Header())
	session, ok := findCookie(cookies, sessionCookieName)
	if !ok {
		t.Fatal("logout did not clear the session cookie")
	}
	if session.Value != "" || session.MaxAge >= 0 {
		t.Fatalf("session cookie not expired: value=%q maxAge=%d", session.Value, session.MaxAge)
	}

	authed, ok := findCookie(cookies, authedCookieName)
	if !ok {
		t.Fatal("logout did not clear the authed indicator cookie")
	}
	if authed.Value != "" || authed.MaxAge >= 0 {
		t.Fatalf("authed cookie not expired: value=%q maxAge=%d", authed.Value, authed.MaxAge)
	}
}
