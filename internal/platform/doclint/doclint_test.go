package doclint

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

const validFrontMatter = "---\ntype: Test doc\ntitle: A doc\ndescription: A description\ntags: [x]\n---\n\n# A doc\n"

// TC-008-1: a markdown file missing a required front-matter field is
// reported by name, and the overall run is non-zero (non-empty issues).
func TestLint_TC008_1_MissingFrontMatterField(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "page.md"), "---\ntype: Test doc\ntitle: A doc\ntags: [x]\n---\n\n# A doc\n")

	issues, err := Lint(root)
	if err != nil {
		t.Fatalf("Lint: %v", err)
	}
	if len(issues) != 1 {
		t.Fatalf("expected exactly 1 issue, got %d: %v", len(issues), issues)
	}
	if !strings.Contains(issues[0].Message, `"description"`) {
		t.Fatalf("expected the missing field to be named, got %q", issues[0].Message)
	}
}

// TC-008-2: a relative link to a nonexistent file is reported with the
// unresolved path.
func TestLint_TC008_2_BrokenLink(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "page.md"), validFrontMatter+"\nSee [missing](does-not-exist.md).\n")

	issues, err := Lint(root)
	if err != nil {
		t.Fatalf("Lint: %v", err)
	}
	if len(issues) != 1 {
		t.Fatalf("expected exactly 1 issue, got %d: %v", len(issues), issues)
	}
	if !strings.Contains(issues[0].Message, "does-not-exist.md") {
		t.Fatalf("expected the broken link path to be named, got %q", issues[0].Message)
	}
}

// A link to a file that does exist must not be reported.
func TestLint_ValidLinkPasses(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "other.md"), validFrontMatter)
	writeFile(t, filepath.Join(root, "page.md"), validFrontMatter+"\nSee [other](other.md).\n")

	issues, err := Lint(root)
	if err != nil {
		t.Fatalf("Lint: %v", err)
	}
	if len(issues) != 0 {
		t.Fatalf("expected no issues, got %v", issues)
	}
}

// TC-008-3: two PRD files sharing the same three-digit ID prefix across
// backlog/ and developed/ are reported as a duplicate.
func TestLint_TC008_3_DuplicatePRDID(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "prds", "backlog", "009-foo.md"), validFrontMatter)
	writeFile(t, filepath.Join(root, "prds", "developed", "009-bar.md"), validFrontMatter)

	issues, err := Lint(root)
	if err != nil {
		t.Fatalf("Lint: %v", err)
	}
	if len(issues) != 1 {
		t.Fatalf("expected exactly 1 issue, got %d: %v", len(issues), issues)
	}
	if !strings.Contains(issues[0].Message, `"009"`) {
		t.Fatalf("expected the duplicate ID to be named, got %q", issues[0].Message)
	}
	if !strings.Contains(issues[0].File, "009-foo.md") || !strings.Contains(issues[0].File, "009-bar.md") {
		t.Fatalf("expected both files to be named, got %q", issues[0].File)
	}
}

// Distinct PRD IDs must not be flagged.
func TestLint_DistinctPRDIDsPass(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "prds", "backlog", "010-foo.md"), validFrontMatter)
	writeFile(t, filepath.Join(root, "prds", "developed", "011-bar.md"), validFrontMatter)

	issues, err := Lint(root)
	if err != nil {
		t.Fatalf("Lint: %v", err)
	}
	if len(issues) != 0 {
		t.Fatalf("expected no issues, got %v", issues)
	}
}

// TC-008-4: a clean docs tree (front matter present, links resolve, no
// duplicate PRD IDs) reports zero issues.
func TestLint_TC008_4_CleanTreePasses(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "README.md"), validFrontMatter+"\nSee [prd](prds/backlog/001-foo.md).\n")
	writeFile(t, filepath.Join(root, "prds", "backlog", "001-foo.md"), validFrontMatter)
	writeFile(t, filepath.Join(root, "prds", "developed", "002-bar.md"), validFrontMatter)

	issues, err := Lint(root)
	if err != nil {
		t.Fatalf("Lint: %v", err)
	}
	if len(issues) != 0 {
		t.Fatalf("expected no issues, got %v", issues)
	}
}

func TestLint_MissingFrontMatterEntirely(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "page.md"), "# No front matter\n")

	issues, err := Lint(root)
	if err != nil {
		t.Fatalf("Lint: %v", err)
	}
	if len(issues) != 1 || !strings.Contains(issues[0].Message, "missing front matter") {
		t.Fatalf("expected a single missing-front-matter issue, got %v", issues)
	}
}

// External links, anchors, and mailto: are not treated as files to resolve.
func TestLint_NonFileLinksAreSkipped(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "page.md"), validFrontMatter+
		"\n[web](https://example.com), [anchor](#section), [mail](mailto:a@b.com).\n")

	issues, err := Lint(root)
	if err != nil {
		t.Fatalf("Lint: %v", err)
	}
	if len(issues) != 0 {
		t.Fatalf("expected no issues, got %v", issues)
	}
}

// The real repository docs/ tree produced by this change must itself be
// clean — this is the regression test that keeps future doc edits honest,
// running under the same `go test ./internal/...` that `make check` uses.
func TestLint_RepositoryDocsTree(t *testing.T) {
	root := "../../../docs"
	if _, err := os.Stat(root); err != nil {
		t.Skipf("repository docs/ tree not found at %s: %v", root, err)
	}

	issues, err := Lint(root)
	if err != nil {
		t.Fatalf("Lint: %v", err)
	}
	if len(issues) != 0 {
		var b strings.Builder
		for _, issue := range issues {
			b.WriteString(issue.String())
			b.WriteString("\n")
		}
		t.Fatalf("expected the repository docs/ tree to be clean, found:\n%s", b.String())
	}
}
