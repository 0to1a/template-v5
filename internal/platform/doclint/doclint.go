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

var (
	linkPattern    = regexp.MustCompile(`\[[^\]]*\]\(([^)]+)\)`)
	frontMatterKey = regexp.MustCompile(`^([A-Za-z_][A-Za-z0-9_]*):\s*(.*)$`)
	prdIDPrefix    = regexp.MustCompile(`^(\d{3})-.+\.md$`)
)

// Lint walks every *.md file under docsRoot and returns every problem
// found, sorted for deterministic output. It only reads files.
func Lint(docsRoot string) ([]Issue, error) {
	files, err := markdownFiles(docsRoot)
	if err != nil {
		return nil, err
	}

	var issues []Issue
	prdFiles := map[string][]string{}

	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			return nil, err
		}
		text := string(content)

		issues = append(issues, checkFrontMatter(file, text)...)
		issues = append(issues, checkLinks(file, text)...)

		if id, ok := prdID(docsRoot, file); ok {
			prdFiles[id] = append(prdFiles[id], file)
		}
	}

	issues = append(issues, checkDuplicatePRDIDs(prdFiles)...)

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

// checkFrontMatter requires the file to start with a "---" delimited
// front-matter block containing every field in requiredFrontMatterFields
// with a non-empty value.
func checkFrontMatter(file, text string) []Issue {
	lines := strings.Split(text, "\n")
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "---" {
		return []Issue{{File: file, Message: "missing front matter (file must start with '---')"}}
	}

	fields := map[string]string{}
	closed := false
	for _, line := range lines[1:] {
		if strings.TrimSpace(line) == "---" {
			closed = true
			break
		}
		if m := frontMatterKey.FindStringSubmatch(line); m != nil {
			fields[m[1]] = strings.TrimSpace(m[2])
		}
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

// prdID reports the three-digit PRD ID for a file under
// docsRoot/prds/backlog or docsRoot/prds/developed, if its filename
// matches that lifecycle convention.
func prdID(docsRoot, file string) (string, bool) {
	backlogDir := filepath.Join(docsRoot, "prds", "backlog")
	developedDir := filepath.Join(docsRoot, "prds", "developed")
	dir := filepath.Dir(file)
	if dir != backlogDir && dir != developedDir {
		return "", false
	}
	m := prdIDPrefix.FindStringSubmatch(filepath.Base(file))
	if m == nil {
		return "", false
	}
	return m[1], true
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
