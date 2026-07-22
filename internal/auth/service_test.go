package auth

import (
	"context"
	"errors"
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

// CreateUser mimics the database's unique-email constraint so a fake
// service under test can't silently create a duplicate account.
func (f *fakeRepo) CreateUser(_ context.Context, normalizedEmail string) (User, error) {
	if _, ok := f.users[normalizedEmail]; ok {
		return User{}, errors.New("fakeRepo: user already exists")
	}
	user := User{PublicUUID: "guest-" + normalizedEmail, Email: normalizedEmail}
	f.users[normalizedEmail] = user
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
	return newTestServiceWithGuestRegistration(t, now, false)
}

func newTestServiceWithGuestRegistration(t *testing.T, now time.Time, isGuestRegistration bool) (*Service, *recordingDelivery) {
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

	service := NewService(repo, delivery, jwtManager, isGuestRegistration)
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

// TC-005-1: with guest registration enabled, requesting a login for an
// unrecognized email creates an active user and delivers a code to it.
func TestRequestLogin_TC005_1(t *testing.T) {
	service, delivery := newTestServiceWithGuestRegistration(t, fixedTime, true)
	ctx := context.Background()

	if err := service.RequestLogin(ctx, "new@example.com"); err != nil {
		t.Fatalf("RequestLogin: %v", err)
	}

	if _, err := service.repo.GetActiveUserByEmail(ctx, "new@example.com"); err != nil {
		t.Fatalf("expected a user to now exist for new@example.com: %v", err)
	}
	if len(delivery.sent) != 1 {
		t.Fatalf("expected exactly one code delivered, got %d", len(delivery.sent))
	}
}

// TC-005-2: with guest registration disabled (the default), requesting a
// login for an unrecognized email creates no account and sends no code.
func TestRequestLogin_TC005_2(t *testing.T) {
	service, delivery := newTestService(t, fixedTime)
	ctx := context.Background()

	if err := service.RequestLogin(ctx, "new@example.com"); err != nil {
		t.Fatalf("RequestLogin: %v", err)
	}

	if _, err := service.repo.GetActiveUserByEmail(ctx, "new@example.com"); !errors.Is(err, ErrUserNotFound) {
		t.Fatalf("expected no user to be created, got err = %v", err)
	}
	if len(delivery.sent) != 0 {
		t.Fatalf("expected no code delivered, got %d", len(delivery.sent))
	}
}

// TC-005-3: with guest registration enabled, requesting a login for an
// email that already has an active user does not create a duplicate.
func TestRequestLogin_TC005_3(t *testing.T) {
	service, delivery := newTestServiceWithGuestRegistration(t, fixedTime, true)
	ctx := context.Background()

	if err := service.RequestLogin(ctx, "admin@localhost"); err != nil {
		t.Fatalf("RequestLogin: %v", err)
	}

	user, err := service.repo.GetActiveUserByEmail(ctx, "admin@localhost")
	if err != nil {
		t.Fatalf("existing user vanished: %v", err)
	}
	if user.PublicUUID != adminUUID {
		t.Fatalf("existing user was replaced: PublicUUID = %s, want %s", user.PublicUUID, adminUUID)
	}
	if len(delivery.sent) != 1 {
		t.Fatalf("expected exactly one code delivered, got %d", len(delivery.sent))
	}
}

// TC-005-4: a user auto-registered by RequestLogin can complete SubmitLogin
// with the delivered code and receive a valid JWT.
func TestSubmitLogin_TC005_4(t *testing.T) {
	service, delivery := newTestServiceWithGuestRegistration(t, fixedTime, true)
	ctx := context.Background()

	if err := service.RequestLogin(ctx, "new@example.com"); err != nil {
		t.Fatalf("RequestLogin: %v", err)
	}
	if len(delivery.sent) != 1 {
		t.Fatalf("expected exactly one code delivered, got %d", len(delivery.sent))
	}

	token, err := service.SubmitLogin(ctx, "new@example.com", delivery.sent[0])
	if err != nil {
		t.Fatalf("SubmitLogin: %v", err)
	}
	if _, err := service.jwtManager.Parse(token); err != nil {
		t.Fatalf("issued token failed validation: %v", err)
	}
}

// TC-015-1: once loginFailureThreshold wrong attempts land for one account,
// even the correct code is rejected — the account is locked out.
func TestSubmitLogin_TC015_1(t *testing.T) {
	service, _ := newTestService(t, fixedTime)
	ctx := context.Background()

	for i := 0; i < loginFailureThreshold; i++ {
		if _, err := service.SubmitLogin(ctx, "admin@localhost", "000000"); err != errUnauthenticated {
			t.Fatalf("failed attempt %d: err = %v, want errUnauthenticated", i, err)
		}
	}

	if _, err := service.SubmitLogin(ctx, "admin@localhost", "123456"); err != errUnauthenticated {
		t.Fatalf("locked-out account with correct code: err = %v, want errUnauthenticated", err)
	}
}

// TC-015-2: a successful login resets the failure counter, so the account
// is not throttled by failures that happened before it.
func TestSubmitLogin_TC015_2(t *testing.T) {
	service, _ := newTestService(t, fixedTime)
	ctx := context.Background()

	for i := 0; i < loginFailureThreshold-1; i++ {
		if _, err := service.SubmitLogin(ctx, "admin@localhost", "000000"); err != errUnauthenticated {
			t.Fatalf("failed attempt %d: err = %v, want errUnauthenticated", i, err)
		}
	}
	if _, err := service.SubmitLogin(ctx, "admin@localhost", "123456"); err != nil {
		t.Fatalf("correct code below threshold: %v", err)
	}

	// If the counter had not reset, these failures plus the ones above would
	// exceed the threshold and the next correct code would be rejected.
	for i := 0; i < loginFailureThreshold-1; i++ {
		if _, err := service.SubmitLogin(ctx, "admin@localhost", "000000"); err != errUnauthenticated {
			t.Fatalf("post-reset failed attempt %d: err = %v, want errUnauthenticated", i, err)
		}
	}
	if _, err := service.SubmitLogin(ctx, "admin@localhost", "123456"); err != nil {
		t.Fatalf("account was throttled despite the earlier reset: %v", err)
	}
}

// TC-015-3: throttling one account never blocks another account's login.
func TestSubmitLogin_TC015_3(t *testing.T) {
	service, delivery := newTestService(t, fixedTime)
	ctx := context.Background()

	for i := 0; i < loginFailureThreshold; i++ {
		if _, err := service.SubmitLogin(ctx, "admin@localhost", "000000"); err != errUnauthenticated {
			t.Fatalf("failed attempt %d: err = %v, want errUnauthenticated", i, err)
		}
	}
	if _, err := service.SubmitLogin(ctx, "admin@localhost", "123456"); err != errUnauthenticated {
		t.Fatal("admin@localhost should be locked out")
	}

	if err := service.RequestLogin(ctx, "user@example.com"); err != nil {
		t.Fatal(err)
	}
	if _, err := service.SubmitLogin(ctx, "user@example.com", delivery.sent[0]); err != nil {
		t.Fatalf("unrelated account was throttled by admin@localhost's lockout: %v", err)
	}
}
