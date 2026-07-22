package auth

import (
	"testing"
	"time"
)

var testSecret = []byte("test-secret-0123456789-0123456789-abc")

const testUUID = "00000000-0000-0000-0000-000000000002"

// fixedTime is an arbitrary instant; all OTP tests derive from it instead of
// time.Now so they are deterministic and never sleep.
var fixedTime = time.Date(2026, 7, 14, 12, 0, 0, 0, time.UTC)

func TestVerifyCode_CurrentStepAccepted(t *testing.T) {
	code := generateCode(testSecret, testUUID, "user@example.com", fixedTime, true)
	if !verifyCode(testSecret, testUUID, "user@example.com", code, fixedTime, true) {
		t.Fatal("current-step TOTP was rejected")
	}
	// Still the same 300-second step 299 seconds later.
	later := fixedTime.Truncate(totpPeriod).Add(totpPeriod - time.Second)
	if !verifyCode(testSecret, testUUID, "user@example.com", code, later, true) {
		t.Fatal("code rejected within its own step")
	}
}

func TestVerifyCode_PreviousStepRejected(t *testing.T) {
	previous := generateCode(testSecret, testUUID, "user@example.com", fixedTime.Add(-totpPeriod), true)
	current := generateCode(testSecret, testUUID, "user@example.com", fixedTime, true)
	if previous == current {
		t.Fatal("test setup: adjacent steps produced the same code")
	}
	// Skew is 0: the previous step's code must not verify now.
	if verifyCode(testSecret, testUUID, "user@example.com", previous, fixedTime, true) {
		t.Fatal("previous-step TOTP was accepted; skew must be 0")
	}
}

func TestVerifyCode_InvalidCodeRejected(t *testing.T) {
	if verifyCode(testSecret, testUUID, "user@example.com", "000000", fixedTime, true) &&
		verifyCode(testSecret, testUUID, "user@example.com", "999999", fixedTime, true) {
		t.Fatal("arbitrary codes were accepted")
	}
	if verifyCode(testSecret, testUUID, "user@example.com", "", fixedTime, true) {
		t.Fatal("empty code was accepted")
	}
}

func TestLocalAdminCode_ExactMatchOnly(t *testing.T) {
	// The fixed development code belongs to admin@localhost exactly, and
	// only in development mode.
	if got := generateCode(testSecret, testUUID, "admin@localhost", fixedTime, true); got != localAdminCode {
		t.Fatalf("admin@localhost code = %s, want %s", got, localAdminCode)
	}
	// Any other @localhost address must go through TOTP, not the fixed code.
	if verifyCode(testSecret, testUUID, "other@localhost", localAdminCode, fixedTime, true) {
		t.Fatal("fixed code accepted for a non-admin @localhost account")
	}
}

// Supports TC-014-1: outside development mode, admin@localhost gets no
// special case — the fixed code never verifies, and its generated code is
// the same TOTP derivation as any other account. The end-to-end behavior
// through Service.SubmitLogin is asserted by TestSubmitLogin_TC014_1 in
// service_test.go.
func TestGenerateCode_ProductionModeDisablesLocalAdmin(t *testing.T) {
	prodCode := generateCode(testSecret, testUUID, "admin@localhost", fixedTime, false)
	if prodCode == localAdminCode {
		t.Fatal("fixed code was generated for admin@localhost outside development mode")
	}
	if !verifyCode(testSecret, testUUID, "admin@localhost", prodCode, fixedTime, false) {
		t.Fatal("admin@localhost's own current TOTP code was rejected outside development mode")
	}
	if verifyCode(testSecret, testUUID, "admin@localhost", localAdminCode, fixedTime, false) {
		t.Fatal("fixed code verified for admin@localhost outside development mode")
	}
}

// Supports TC-014-2: development mode preserves today's behavior for the
// fixed code.
func TestGenerateCode_DevelopmentModePreservesLocalAdmin(t *testing.T) {
	if got := generateCode(testSecret, testUUID, "admin@localhost", fixedTime, true); got != localAdminCode {
		t.Fatalf("admin@localhost code = %s, want %s", got, localAdminCode)
	}
	if !verifyCode(testSecret, testUUID, "admin@localhost", localAdminCode, fixedTime, true) {
		t.Fatal("fixed code was rejected for admin@localhost in development mode")
	}
}

func TestNormalizeEmail(t *testing.T) {
	if got := normalizeEmail("  Admin@LocalHost \n"); got != "admin@localhost" {
		t.Fatalf("normalizeEmail = %q", got)
	}
}
