// Package doclint implements the read-only checks behind `make doc-lint`.
// Lint walks a docs/ tree and classifies problems it finds; it never
// writes anything, matching the same "diagnose, don't mutate" shape as
// internal/platform/doctor. It deliberately hand-rolls a minimal
// front-matter scanner instead of a full YAML parser: this repo's front
// matter is a flat set of "key: value" lines, and that is all that needs
// checking here.
package doclint

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

// Issue is one problem found in the docs tree.
type Issue struct {
	File    string
	Message string
}

func (i Issue) String() string { return fmt.Sprintf("%s: %s", i.File, i.Message) }

// requiredFrontMatterFields is the minimum every markdown file under docs/
// must declare. Canonical guide docs additionally use status/owner/
// last_reviewed, but that convention isn't enforced here to keep this
// check basic and to avoid retroactively failing existing PRDs and
// examples that predate it.
var requiredFrontMatterFields = []string{"type", "title", "description", "tags"}

// traceabilityEnforcedFromID is the first PRD ID that must have every one
// of its TC-<id>-n test cases traced to a test/script file in the
// repository. PRDs 002-013 predate this convention: 002 and 003 were
// validated by their CI pipeline's own behavior rather than a grep-able
// test, and 009-013 (Wave C) shipped documentation- and manual-review-only
// test cases before this check existed. Like requiredFrontMatterFields
// above, an already-developed PRD is never retroactively failed against a
// newer, stricter check. Compares as a three-digit zero-padded string,
// which sorts the same as numerically.
const traceabilityEnforcedFromID = "014"

var (
	linkPattern    = regexp.MustCompile(`\[[^\]]*\]\(([^)]+)\)`)
	frontMatterKey = regexp.MustCompile(`^([A-Za-z_][A-Za-z0-9_]*):\s*(.*)$`)
	prdIDPrefix    = regexp.MustCompile(`^(\d{3})-.+\.md$`)
	tcIDPattern    = regexp.MustCompile(`TC-\d{3}-\d+`)
)

// Lint walks every *.md file under docsRoot and returns every problem
// found, sorted for deterministic output. repoRoot is only used to search
// for TC-<id>-n traceability (see traceableTCIDs); it is typically the
// repository root, one level above docsRoot. now is compared against
// waiver_expires dates. Lint only reads files.
func Lint(docsRoot, repoRoot string, now time.Time) ([]Issue, error) {
	files, err := markdownFiles(docsRoot)
	if err != nil {
		return nil, err
	}

	var issues []Issue
	prdFiles := map[string][]string{}
	type developedPRD struct {
		file string
		tcs  []string
	}
	var developedPRDs []developedPRD

	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			return nil, err
		}
		text := string(content)

		fields, closed := parseFrontMatter(text)
		issues = append(issues, checkFrontMatter(file, fields, closed)...)
		issues = append(issues, checkLinks(file, text)...)

		id, lifecycle, isPRD := prdInfo(docsRoot, file)
		if isPRD {
			prdFiles[id] = append(prdFiles[id], file)
			if closed {
				issues = append(issues, checkProblemBrief(file, fields, now)...)
			}
			if lifecycle == "developed" && id >= traceabilityEnforcedFromID {
				developedPRDs = append(developedPRDs, developedPRD{file: file, tcs: uniqueTCIDs(text)})
			}
		}
	}

	issues = append(issues, checkDuplicatePRDIDs(prdFiles)...)

	if len(developedPRDs) > 0 {
		traced, err := traceableTCIDs(repoRoot)
		if err != nil {
			return nil, err
		}
		for _, prd := range developedPRDs {
			for _, tc := range prd.tcs {
				if !traced[tc] {
					issues = append(issues, Issue{File: prd.file, Message: fmt.Sprintf("developed PRD test case %q has no trace in any test/script file in the repository", tc)})
				}
			}
		}
	}

	sort.Slice(issues, func(i, j int) bool {
		if issues[i].File != issues[j].File {
			return issues[i].File < issues[j].File
		}
		return issues[i].Message < issues[j].Message
	})
	return issues, nil
}

func markdownFiles(root string) ([]string, error) {
	var files []string
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(path, ".md") {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(files)
	return files, nil
}

// parseFrontMatter extracts the "key: value" front-matter block a file
// must start with. closed reports whether a terminating "---" was found;
// fields is empty (never nil) when there is no front matter at all.
func parseFrontMatter(text string) (fields map[string]string, closed bool) {
	fields = map[string]string{}
	lines := strings.Split(text, "\n")
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "---" {
		return fields, false
	}

	for _, line := range lines[1:] {
		if strings.TrimSpace(line) == "---" {
			return fields, true
		}
		if m := frontMatterKey.FindStringSubmatch(line); m != nil {
			fields[m[1]] = strings.TrimSpace(m[2])
		}
	}
	return fields, false
}

// checkFrontMatter requires the file to start with a "---" delimited
// front-matter block containing every field in requiredFrontMatterFields
// with a non-empty value.
func checkFrontMatter(file string, fields map[string]string, closed bool) []Issue {
	if len(fields) == 0 && !closed {
		return []Issue{{File: file, Message: "missing front matter (file must start with '---')"}}
	}
	if !closed {
		return []Issue{{File: file, Message: "front matter is never closed with a second '---'"}}
	}

	var issues []Issue
	for _, field := range requiredFrontMatterFields {
		if fields[field] == "" {
			issues = append(issues, Issue{File: file, Message: fmt.Sprintf("front matter missing required field %q", field)})
		}
	}
	return issues
}

// checkLinks resolves every relative markdown link against the file's own
// directory and reports any target that doesn't exist on disk. External
// links, mailto:, and in-page anchors are skipped.
func checkLinks(file, text string) []Issue {
	var issues []Issue
	dir := filepath.Dir(file)

	for _, m := range linkPattern.FindAllStringSubmatch(text, -1) {
		target := strings.TrimSpace(m[1])
		if target == "" || strings.HasPrefix(target, "#") ||
			strings.Contains(target, "://") || strings.HasPrefix(target, "mailto:") {
			continue
		}
		// Strip an in-page anchor from a same-file link, e.g. "foo.md#section".
		if i := strings.Index(target, "#"); i >= 0 {
			target = target[:i]
		}
		if target == "" {
			continue
		}

		resolved := filepath.Join(dir, target)
		if _, err := os.Stat(resolved); err != nil {
			issues = append(issues, Issue{File: file, Message: fmt.Sprintf("broken internal link %q", m[1])})
		}
	}
	return issues
}

// prdInfo reports the three-digit PRD ID and lifecycle folder
// ("backlog" or "developed") for a file under docsRoot/prds/, if its
// filename matches that lifecycle convention.
func prdInfo(docsRoot, file string) (id, lifecycle string, ok bool) {
	backlogDir := filepath.Join(docsRoot, "prds", "backlog")
	developedDir := filepath.Join(docsRoot, "prds", "developed")
	dir := filepath.Dir(file)
	switch dir {
	case backlogDir:
		lifecycle = "backlog"
	case developedDir:
		lifecycle = "developed"
	default:
		return "", "", false
	}
	m := prdIDPrefix.FindStringSubmatch(filepath.Base(file))
	if m == nil {
		return "", "", false
	}
	return m[1], lifecycle, true
}

func checkDuplicatePRDIDs(prdFiles map[string][]string) []Issue {
	var issues []Issue
	for id, files := range prdFiles {
		if len(files) > 1 {
			sort.Strings(files)
			issues = append(issues, Issue{
				File:    strings.Join(files, ", "),
				Message: fmt.Sprintf("duplicate PRD ID %q used by multiple files", id),
			})
		}
	}
	return issues
}

// checkProblemBrief validates a PRD's problem_brief front-matter field,
// per docs/product/README.md's contract, but only when the field is
// present at all: PRDs 001-007 predate this gate and have no such field,
// and are intentionally left unflagged, matching requiredFrontMatterFields'
// same "never retroactively fail pre-existing docs" precedent.
func checkProblemBrief(file string, fields map[string]string, now time.Time) []Issue {
	brief := fields["problem_brief"]
	if brief == "" {
		return nil
	}

	if brief == "waiver" {
		var issues []Issue
		for _, field := range []string{"waiver_owner", "waiver_reason", "waiver_expires"} {
			if fields[field] == "" {
				issues = append(issues, Issue{File: file, Message: fmt.Sprintf("problem_brief waiver missing required field %q", field)})
			}
		}
		if expires := fields["waiver_expires"]; expires != "" {
			parsed, err := time.Parse("2006-01-02", expires)
			if err != nil {
				issues = append(issues, Issue{File: file, Message: fmt.Sprintf("waiver_expires %q is not a YYYY-MM-DD date", expires)})
			} else if parsed.Before(now) {
				issues = append(issues, Issue{File: file, Message: fmt.Sprintf("problem_brief waiver expired on %s", expires)})
			}
		}
		return issues
	}

	resolved := filepath.Join(filepath.Dir(file), brief)
	briefContent, err := os.ReadFile(resolved)
	if err != nil {
		return []Issue{{File: file, Message: fmt.Sprintf("problem_brief %q does not resolve to a file", brief)}}
	}
	briefFields, closed := parseFrontMatter(string(briefContent))
	if !closed || briefFields["status"] != "proceed" {
		return []Issue{{File: file, Message: fmt.Sprintf("problem_brief %q does not have status: proceed", brief)}}
	}
	return nil
}

func uniqueTCIDs(text string) []string {
	seen := map[string]bool{}
	var ids []string
	for _, id := range tcIDPattern.FindAllString(text, -1) {
		if !seen[id] {
			seen[id] = true
			ids = append(ids, id)
		}
	}
	return ids
}

// isTraceableFile reports whether path is the kind of file a TC-<id>-n
// case can be proven from: an automated test, or a script/Makefile a
// human or CI can read to see the case exercised (e.g. a smoke test or a
// `make` target that names the ID it satisfies).
func isTraceableFile(path string) bool {
	base := filepath.Base(path)
	switch {
	case strings.HasSuffix(base, "_test.go"),
		strings.HasSuffix(base, ".test.ts"),
		strings.HasSuffix(base, ".test.js"),
		strings.HasSuffix(base, ".spec.ts"),
		strings.HasSuffix(base, ".spec.js"),
		strings.HasSuffix(base, ".sh"):
		return true
	case base == "Makefile":
		return true
	default:
		return false
	}
}

// skippedTraceDirs are never descended into while searching for TC-id
// traces: build output, dependency, and VCS directories that are large
// and, for node_modules, not owned by this repo.
var skippedTraceDirs = map[string]bool{
	".git":         true,
	"node_modules": true,
	"dist":         true,
	"bin":          true,
	".svelte-kit":  true,
}

// traceableTCIDs walks repoRoot and returns the set of every TC-<id>-n
// string found in any isTraceableFile.
func traceableTCIDs(repoRoot string) (map[string]bool, error) {
	found := map[string]bool{}
	err := filepath.WalkDir(repoRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if skippedTraceDirs[d.Name()] {
				return filepath.SkipDir
			}
			return nil
		}
		if !isTraceableFile(path) {
			return nil
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		for _, id := range tcIDPattern.FindAllString(string(content), -1) {
			found[id] = true
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return found, nil
}
