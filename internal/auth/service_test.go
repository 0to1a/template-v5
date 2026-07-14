package auth

import (
	"context"
	"testing"
	"time"
)

// fakeRepo serves a fixed set of users keyed by normalized email.
type fakeRepo struct {
	users map[string]User
}

func (f *fakeRepo) GetActiveUserByEmail(_ context.Context, normalizedEmail string) (User, error) {
	user, ok := f.users[normalizedEmail]
	if !ok {
		return User{}, ErrUserNotFound
	}
	return user, nil
}

// recordingDelivery captures the codes it is asked to send.
type recordingDelivery struct {
	sent []string
}

func (d *recordingDelivery) SendLoginCode(_ context.Context, _, code string) error {
	d.sent = append(d.sent, code)
	return nil
}

const adminUUID = "00000000-0000-0000-0000-000000000001"

func newTestService(t *testing.T, now time.Time) (*Service, *recordingDelivery) {
	t.Helper()
	jwtManager, err := NewJWTManager(jwtTestSecret)
	if err != nil {
		t.Fatal(err)
	}
	jwtManager.now = func() time.Time { return now }

	repo := &fakeRepo{users: map[string]User{
		"admin@localhost":  {PublicUUID: adminUUID, Email: "admin@localhost"},
		"user@example.com": {PublicUUID: testUUID, Email: "user@example.com"},
	}}
	delivery := &recordingDelivery{}

	service := NewService(repo, delivery, jwtManager)
	service.now = func() time.Time { return now }
	return service, delivery
}

// TC-001-2: RequestLogin returns the same generic result for registered and
// unregistered emails, so callers cannot enumerate accounts.
func TestRequestLogin_TC001_2(t *testing.T) {
	service, delivery := newTestService(t, fixedTime)
	ctx := context.Background()

	if err := service.RequestLogin(ctx, "user@example.com"); err != nil {
		t.Fatalf("registered email: %v", err)
	}
	if err := service.RequestLogin(ctx, "nobody@example.com"); err != nil {
		t.Fatalf("unregistered email must get the same generic result, got: %v", err)
	}

	// The observable response is identical; only delivery differs, and
	// delivery happens out-of-band where the caller can't see it.
	if len(delivery.sent) != 1 {
		t.Fatalf("expected exactly one code delivered, got %d", len(delivery.sent))
	}
}

// TC-001-3: the seeded local admin logs in with the fixed development code
// and receives a valid HS256 JWT whose subject is its public identity.
func TestSubmitLogin_TC001_3(t *testing.T) {
	service, _ := newTestService(t, fixedTime)

	token, err := service.SubmitLogin(context.Background(), "admin@localhost", "123456")
	if err != nil {
		t.Fatal(err)
	}

	principal, err := service.jwtManager.Parse(token)
	if err != nil {
		t.Fatalf("issued token failed validation: %v", err)
	}
	if principal.PublicUUID != adminUUID {
		t.Fatalf("subject = %s, want %s", principal.PublicUUID, adminUUID)
	}
}

// TC-001-4: a wrong code returns the generic unauthenticated error and no
// token; an unknown user gets the exact same error.
func TestSubmitLogin_TC001_4(t *testing.T) {
	service, _ := newTestService(t, fixedTime)
	ctx := context.Background()

	token, err := service.SubmitLogin(ctx, "admin@localhost", "654321")
	if err != errUnauthenticated {
		t.Fatalf("wrong code: err = %v, want errUnauthenticated", err)
	}
	if token != "" {
		t.Fatal("wrong code still produced a token")
	}

	_, unknownErr := service.SubmitLogin(ctx, "nobody@example.com", "123456")
	if unknownErr != errUnauthenticated {
		t.Fatalf("unknown user: err = %v, want the same generic error", unknownErr)
	}
}

// A regular (non-admin) user logs in with the current TOTP step, and the
// previous step's code is rejected — end to end through the service.
func TestSubmitLogin_TOTPUser(t *testing.T) {
	service, delivery := newTestService(t, fixedTime)
	ctx := context.Background()

	if err := service.RequestLogin(ctx, "user@example.com"); err != nil {
		t.Fatal(err)
	}
	if len(delivery.sent) != 1 {
		t.Fatalf("expected one delivered code, got %d", len(delivery.sent))
	}

	if _, err := service.SubmitLogin(ctx, "user@example.com", delivery.sent[0]); err != nil {
		t.Fatalf("current TOTP rejected: %v", err)
	}

	stale := generateCode([]byte(jwtTestSecret), testUUID, "user@example.com", fixedTime.Add(-totpPeriod))
	if _, err := service.SubmitLogin(ctx, "user@example.com", stale); err != errUnauthenticated {
		t.Fatalf("previous-step TOTP: err = %v, want errUnauthenticated", err)
	}
}
