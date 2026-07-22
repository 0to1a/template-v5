// Package vulnscan implements the threshold-and-exception evaluation
// behind `make vuln-scan`. It parses govulncheck's and `bun audit`'s JSON
// output into a common Finding shape and decides which findings should
// fail the scan; it never invokes those tools itself (see cmd/vulnscan)
// and never writes anything.
package vulnscan

import (
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"
)

// Severity levels, ordered low to critical. The Go vulnerability database
// (vuln.go.dev) does not publish graded severities the way npm advisories
// do, so every govulncheck finding (which, at the default "symbol" scan
// level, only fires for vulnerable code actually called, not merely
// imported) is reported at SeverityHigh.
const (
	SeverityLow      = "low"
	SeverityModerate = "moderate"
	SeverityHigh     = "high"
	SeverityCritical = "critical"
)

var severityRank = map[string]int{
	SeverityLow:      1,
	SeverityModerate: 2,
	SeverityHigh:     3,
	SeverityCritical: 4,
}

// Finding is one vulnerability reported by a scanner, reduced to what
// Evaluate needs: an ID to check against the exception list, and a
// severity to compare against the threshold.
type Finding struct {
	ID       string
	Severity string
}

// Exception suppresses one Finding ID until Expires. Every field is
// required so an exception always names who accepted the risk, why, and
// when it must be revisited.
type Exception struct {
	ID      string
	Package string
	Owner   string
	Reason  string
	Expires time.Time
}

// Evaluate returns every finding at or above threshold whose ID has no
// exception with an Expires date after now. An exception whose Expires
// date has passed no longer suppresses its finding.
func Evaluate(findings []Finding, exceptions []Exception, threshold string, now time.Time) []Finding {
	byID := make(map[string]Exception, len(exceptions))
	for _, e := range exceptions {
		byID[e.ID] = e
	}

	min := severityRank[threshold]
	var failing []Finding
	for _, f := range findings {
		if severityRank[f.Severity] < min {
			continue
		}
		if exc, ok := byID[f.ID]; ok && exc.Expires.After(now) {
			continue
		}
		failing = append(failing, f)
	}
	return failing
}

// ParseExceptions reads the hand-rolled exceptions file format: "key:
// value" blocks separated by a line containing only "---", the same
// separator style internal/platform/doclint uses for front matter. Blank
// lines and lines starting with "#" are ignored. Every field is required.
func ParseExceptions(text string) ([]Exception, error) {
	var exceptions []Exception
	for _, block := range strings.Split(text, "\n---\n") {
		block = strings.TrimSpace(block)
		if block == "" {
			continue
		}

		fields := map[string]string{}
		for _, line := range strings.Split(block, "\n") {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			key, value, ok := strings.Cut(line, ":")
			if !ok {
				continue
			}
			fields[strings.TrimSpace(key)] = strings.TrimSpace(value)
		}
		if len(fields) == 0 {
			continue
		}

		for _, required := range []string{"id", "package", "owner", "reason", "expires"} {
			if fields[required] == "" {
				return nil, fmt.Errorf("vulnscan: exception block missing required field %q: %q", required, block)
			}
		}
		expires, err := time.Parse("2006-01-02", fields["expires"])
		if err != nil {
			return nil, fmt.Errorf("vulnscan: exception %q: invalid expires %q: %w", fields["id"], fields["expires"], err)
		}

		exceptions = append(exceptions, Exception{
			ID:      fields["id"],
			Package: fields["package"],
			Owner:   fields["owner"],
			Reason:  fields["reason"],
			Expires: expires,
		})
	}
	return exceptions, nil
}

// govulncheckEntry is the subset of govulncheck -json's streamed output
// this package cares about. govulncheck emits a sequence of top-level JSON
// values (not a single array or JSON-Lines document), each either a
// config/progress/osv/finding object; only "finding" entries at the
// default "symbol" scan level represent a vulnerability whose code is
// actually reachable, which is what should fail a scan.
type govulncheckEntry struct {
	Finding *struct {
		OSV string `json:"osv"`
	} `json:"finding"`
}

// ParseGovulncheckFindings reads a `govulncheck -json` stream and returns
// one Finding per distinct reachable OSV ID.
func ParseGovulncheckFindings(r io.Reader) ([]Finding, error) {
	dec := json.NewDecoder(r)
	seen := map[string]bool{}
	var findings []Finding
	for {
		var entry govulncheckEntry
		err := dec.Decode(&entry)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("vulnscan: parsing govulncheck output: %w", err)
		}
		if entry.Finding == nil || entry.Finding.OSV == "" || seen[entry.Finding.OSV] {
			continue
		}
		seen[entry.Finding.OSV] = true
		findings = append(findings, Finding{ID: entry.Finding.OSV, Severity: SeverityHigh})
	}
	return findings, nil
}

// bunAuditOutput mirrors `bun audit --json`'s shape: a map from package
// name to the advisories affecting it.
type bunAuditOutput map[string][]struct {
	ID       int    `json:"id"`
	Severity string `json:"severity"`
}

// ParseBunAuditFindings reads `bun audit --json`'s output and returns one
// Finding per distinct advisory ID, with bun's own graded severity
// (low/moderate/high/critical) preserved.
func ParseBunAuditFindings(r io.Reader) ([]Finding, error) {
	var output bunAuditOutput
	if err := json.NewDecoder(r).Decode(&output); err != nil {
		return nil, fmt.Errorf("vulnscan: parsing bun audit output: %w", err)
	}

	seen := map[string]bool{}
	var findings []Finding
	for _, advisories := range output {
		for _, a := range advisories {
			id := strconv.Itoa(a.ID)
			if seen[id] {
				continue
			}
			seen[id] = true
			findings = append(findings, Finding{ID: id, Severity: a.Severity})
		}
	}
	return findings, nil
}
