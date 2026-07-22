// Package docs contains test-only assertions that specific canonical
// docs/ claims stay true, for PRD acceptance criteria that doc-lint's
// generic front-matter/link/traceability checks can't verify on their
// own. It ships no production code.
package docs

import (
	"os"
	"strings"
	"testing"
)

func readRepoFile(t *testing.T, relPath string) string {
	t.Helper()
	// internal/platform/docs -> repo root is three levels up.
	data, err := os.ReadFile("../../../" + relPath)
	if err != nil {
		t.Fatalf("reading %s: %v", relPath, err)
	}
	return string(data)
}

// TC-020-1: the onboarding doc gives a fresh-provisioning path for a
// developer/agent with no existing PostgreSQL role or database, using
// the same postgres:// URL scheme .env.example already documents.
func TestOnboardingDoc_TC020_1_FreshPostgresProvisioning(t *testing.T) {
	developerDoc := readRepoFile(t, "docs/onboarding/developer.md")
	envExample := readRepoFile(t, ".env.example")

	if !strings.Contains(developerDoc, "CREATE ROLE") {
		t.Error("docs/onboarding/developer.md is missing a CREATE ROLE command for provisioning a fresh Postgres role")
	}
	if !strings.Contains(developerDoc, "CREATE DATABASE") {
		t.Error("docs/onboarding/developer.md is missing a CREATE DATABASE command for provisioning a fresh Postgres database")
	}

	const scheme = "postgres://"
	if !strings.Contains(envExample, scheme) {
		t.Fatalf(".env.example does not use the expected %q scheme; test fixture assumption is stale", scheme)
	}
	if !strings.Contains(developerDoc, scheme) {
		t.Error("docs/onboarding/developer.md's provisioning example does not use the same postgres:// scheme as .env.example")
	}
}

// TC-020-2: the PRD workflow doc names the concurrent-PRD-ID-collision
// scenario explicitly, and states the resolution (renumber whichever PRD
// merges second), rather than leaving it as an undocumented surprise.
func TestPRDReadme_TC020_2_DocumentsIDCollisionHandling(t *testing.T) {
	prdReadme := readRepoFile(t, "docs/prds/README.md")

	if !strings.Contains(prdReadme, "duplicate PRD ID") {
		t.Error(`docs/prds/README.md does not mention the "duplicate PRD ID" doc-lint check in the context of concurrent PRD drafting`)
	}
	if !strings.Contains(prdReadme, "renumbering whichever PRD merges second") {
		t.Error("docs/prds/README.md does not state that the resolution to a PRD ID collision is renumbering whichever PRD merges second")
	}
}
