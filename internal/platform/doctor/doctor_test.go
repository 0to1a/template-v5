package doctor

import (
	"errors"
	"strings"
	"testing"
)

// TC-006-1: a Go version that does not match the required minor version
// fails with remediation naming both versions and the download page.
func TestGoVersion_TC006_1_Mismatch(t *testing.T) {
	c := GoVersion("go1.24.3", "1.26")
	if c.OK() {
		t.Fatalf("expected failure, got %+v", c)
	}
	if !strings.Contains(c.Detail, "go1.24.3") || !strings.Contains(c.Detail, "1.26") {
		t.Fatalf("detail should name found and required versions, got %q", c.Detail)
	}
	if !strings.Contains(c.Remediation, "go.dev/dl") {
		t.Fatalf("remediation should point to the Go download page, got %q", c.Remediation)
	}
}

func TestGoVersion_TC006_1_NotFound(t *testing.T) {
	c := GoVersion("", "1.26")
	if c.OK() {
		t.Fatalf("expected failure, got %+v", c)
	}
	if c.Remediation == "" {
		t.Fatalf("expected remediation for a missing go binary")
	}
}

// TC-006-2: the pinned minor version (any patch) reports OK.
func TestGoVersion_TC006_2_Match(t *testing.T) {
	for _, reported := range []string{"go1.26", "go1.26.5", "go1.26.12"} {
		c := GoVersion(reported, "1.26")
		if !c.OK() {
			t.Fatalf("expected OK for %q, got %+v", reported, c)
		}
		if c.Remediation != "" {
			t.Fatalf("expected no remediation on success, got %q", c.Remediation)
		}
	}
}

// TC-006-3: a missing or mismatched Bun version fails with the pinned
// version and an install command.
func TestBunVersion_TC006_3_Mismatch(t *testing.T) {
	c := BunVersion("1.2.0", "1.3.14")
	if c.OK() {
		t.Fatalf("expected failure, got %+v", c)
	}
	if !strings.Contains(c.Remediation, "1.3.14") || !strings.Contains(c.Remediation, "bun.sh/install") {
		t.Fatalf("remediation should name the pinned version and install command, got %q", c.Remediation)
	}
}

func TestBunVersion_TC006_3_NotFound(t *testing.T) {
	c := BunVersion("", "1.3.14")
	if c.OK() {
		t.Fatalf("expected failure, got %+v", c)
	}
	if !strings.Contains(c.Detail, "not found") {
		t.Fatalf("detail should say bun was not found, got %q", c.Detail)
	}
}

// TC-006-4: the exact pinned version reports OK.
func TestBunVersion_TC006_4_Match(t *testing.T) {
	c := BunVersion("1.3.14", "1.3.14")
	if !c.OK() {
		t.Fatalf("expected OK, got %+v", c)
	}
	if c.Remediation != "" {
		t.Fatalf("expected no remediation on success, got %q", c.Remediation)
	}
}

// TC-006-5: missing .env, a short JWT_SECRET, and an unset DATABASE_URL
// each fail with a distinct remediation.
func TestConfigChecks_TC006_5_Missing(t *testing.T) {
	if c := EnvFile(false); c.OK() || c.Remediation == "" {
		t.Fatalf("expected failing EnvFile check with remediation, got %+v", c)
	}
	if c := JWTSecret("too-short"); c.OK() || c.Remediation == "" {
		t.Fatalf("expected failing JWTSecret check with remediation, got %+v", c)
	}
	if c := DatabaseURLConfigured(""); c.OK() || c.Remediation == "" {
		t.Fatalf("expected failing DatabaseURLConfigured check with remediation, got %+v", c)
	}
}

// TC-006-6: a present .env, a long enough JWT_SECRET, and a set
// DATABASE_URL each report OK.
func TestConfigChecks_TC006_6_Valid(t *testing.T) {
	if c := EnvFile(true); !c.OK() {
		t.Fatalf("expected OK, got %+v", c)
	}
	if c := JWTSecret(strings.Repeat("a", 32)); !c.OK() {
		t.Fatalf("expected OK, got %+v", c)
	}
	if c := DatabaseURLConfigured("postgres://user:pass@localhost:5432/db"); !c.OK() {
		t.Fatalf("expected OK, got %+v", c)
	}
}

// TC-006-7: a malformed MAIL_URL fails with remediation naming the expected
// shape; an unset MAIL_URL is valid.
func TestMailURLConfigured_TC006_7_Malformed(t *testing.T) {
	c := MailURLConfigured("not-a-url://missing-port")
	if c.OK() {
		t.Fatalf("expected failure, got %+v", c)
	}
	if !strings.Contains(c.Remediation, "smtp://") {
		t.Fatalf("remediation should name the expected smtp:// shape, got %q", c.Remediation)
	}
}

func TestMailURLConfigured_TC006_7_Unset(t *testing.T) {
	c := MailURLConfigured("")
	if !c.OK() {
		t.Fatalf("expected OK for unset MAIL_URL, got %+v", c)
	}
}

// TC-006-8: a PostgreSQL ping failure is reported verbatim with remediation,
// and this classification never itself attempts a connection or write.
func TestPostgresReachable_TC006_8_Failure(t *testing.T) {
	c := PostgresReachable(errors.New("password authentication failed for user \"user\""))
	if c.OK() {
		t.Fatalf("expected failure, got %+v", c)
	}
	if !strings.Contains(c.Detail, "authentication failed") {
		t.Fatalf("detail should include the underlying error, got %q", c.Detail)
	}
	if c.Remediation == "" {
		t.Fatalf("expected remediation for a failed connection")
	}
}

// TC-006-9: a nil ping error reports OK.
func TestPostgresReachable_TC006_9_Success(t *testing.T) {
	c := PostgresReachable(nil)
	if !c.OK() {
		t.Fatalf("expected OK, got %+v", c)
	}
}
