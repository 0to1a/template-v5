package vulnscan

import (
	"strings"
	"testing"
	"time"
)

var fixedNow = time.Date(2026, 7, 22, 0, 0, 0, 0, time.UTC)

// TC-019-4: a finding with no exception entry fails.
func TestEvaluate_TC010_4_NoExceptionFails(t *testing.T) {
	findings := []Finding{{ID: "GO-2024-9999", Severity: SeverityHigh}}

	failing := Evaluate(findings, nil, SeverityHigh, fixedNow)

	if len(failing) != 1 || failing[0].ID != "GO-2024-9999" {
		t.Fatalf("expected the finding to fail, got %v", failing)
	}
}

// TC-019-5: an unexpired exception suppresses a finding; an expired one
// does not.
func TestEvaluate_TC010_5_ExceptionExpiry(t *testing.T) {
	finding := Finding{ID: "GO-2024-9999", Severity: SeverityHigh}

	unexpired := Exception{ID: "GO-2024-9999", Expires: fixedNow.AddDate(0, 1, 0)}
	if failing := Evaluate([]Finding{finding}, []Exception{unexpired}, SeverityHigh, fixedNow); len(failing) != 0 {
		t.Fatalf("expected the unexpired exception to suppress the finding, got %v", failing)
	}

	expired := Exception{ID: "GO-2024-9999", Expires: fixedNow.AddDate(0, -1, 0)}
	if failing := Evaluate([]Finding{finding}, []Exception{expired}, SeverityHigh, fixedNow); len(failing) != 1 {
		t.Fatalf("expected the expired exception to no longer suppress the finding, got %v", failing)
	}
}

// A finding below threshold never fails, exception or not.
func TestEvaluate_BelowThresholdPasses(t *testing.T) {
	findings := []Finding{{ID: "GHSA-low-1", Severity: SeverityLow}}

	failing := Evaluate(findings, nil, SeverityHigh, fixedNow)

	if len(failing) != 0 {
		t.Fatalf("expected no failing findings, got %v", failing)
	}
}

func TestParseExceptions_SingleBlock(t *testing.T) {
	text := "id: GO-2024-9999\npackage: golang.org/x/text\nowner: Founding Engineer\nreason: accepted risk\nexpires: 2026-09-01\n"

	exceptions, err := ParseExceptions(text)
	if err != nil {
		t.Fatalf("ParseExceptions: %v", err)
	}
	if len(exceptions) != 1 {
		t.Fatalf("expected exactly 1 exception, got %d: %v", len(exceptions), exceptions)
	}
	e := exceptions[0]
	if e.ID != "GO-2024-9999" || e.Package != "golang.org/x/text" || e.Owner != "Founding Engineer" || e.Reason != "accepted risk" {
		t.Fatalf("unexpected exception fields: %+v", e)
	}
	if !e.Expires.Equal(time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("unexpected expires: %v", e.Expires)
	}
}

func TestParseExceptions_MultipleBlocksAndComments(t *testing.T) {
	text := "# leading comment, ignored\n" +
		"id: GO-1\npackage: a\nowner: A\nreason: r1\nexpires: 2026-01-01\n" +
		"\n---\n\n" +
		"id: GO-2\npackage: b\nowner: B\nreason: r2\nexpires: 2026-02-01\n"

	exceptions, err := ParseExceptions(text)
	if err != nil {
		t.Fatalf("ParseExceptions: %v", err)
	}
	if len(exceptions) != 2 {
		t.Fatalf("expected exactly 2 exceptions, got %d: %v", len(exceptions), exceptions)
	}
	if exceptions[0].ID != "GO-1" || exceptions[1].ID != "GO-2" {
		t.Fatalf("unexpected exception order/ids: %v", exceptions)
	}
}

func TestParseExceptions_EmptyTextIsNotAnError(t *testing.T) {
	exceptions, err := ParseExceptions("")
	if err != nil {
		t.Fatalf("ParseExceptions: %v", err)
	}
	if len(exceptions) != 0 {
		t.Fatalf("expected no exceptions, got %v", exceptions)
	}
}

func TestParseExceptions_MissingFieldIsAnError(t *testing.T) {
	_, err := ParseExceptions("id: GO-1\npackage: a\nowner: A\nreason: r1\n")
	if err == nil || !strings.Contains(err.Error(), `"expires"`) {
		t.Fatalf("expected an error naming the missing expires field, got %v", err)
	}
}

func TestParseGovulncheckFindings_DistinctReachableOSVIDs(t *testing.T) {
	stream := `
{"config":{"protocol_version":"v1.0.0"}}
{"progress":{"message":"scanning"}}
{"osv":{"id":"GO-2024-1111","summary":"something"}}
{"finding":{"osv":"GO-2024-1111","fixed_version":"v1.2.3","trace":[{"module":"m"}]}}
{"finding":{"osv":"GO-2024-1111","fixed_version":"v1.2.3","trace":[{"module":"m","function":"f"}]}}
{"finding":{"osv":"GO-2024-2222","fixed_version":"v2.0.0","trace":[{"module":"n"}]}}
`
	findings, err := ParseGovulncheckFindings(strings.NewReader(stream))
	if err != nil {
		t.Fatalf("ParseGovulncheckFindings: %v", err)
	}
	if len(findings) != 2 {
		t.Fatalf("expected exactly 2 distinct findings, got %d: %v", len(findings), findings)
	}
	for _, f := range findings {
		if f.Severity != SeverityHigh {
			t.Fatalf("expected every Go finding to be SeverityHigh, got %+v", f)
		}
	}
}

func TestParseGovulncheckFindings_NoFindings(t *testing.T) {
	stream := `{"config":{"protocol_version":"v1.0.0"}}` + "\n"

	findings, err := ParseGovulncheckFindings(strings.NewReader(stream))
	if err != nil {
		t.Fatalf("ParseGovulncheckFindings: %v", err)
	}
	if len(findings) != 0 {
		t.Fatalf("expected no findings, got %v", findings)
	}
}

func TestParseBunAuditFindings(t *testing.T) {
	json := `{"cookie":[{"id":1103907,"url":"https://github.com/advisories/GHSA-pxg6-pf52-xh8x","title":"t","severity":"low","vulnerable_versions":"<0.7.0","cwe":["CWE-74"],"cvss":{"score":0,"vectorString":null}}]}`

	findings, err := ParseBunAuditFindings(strings.NewReader(json))
	if err != nil {
		t.Fatalf("ParseBunAuditFindings: %v", err)
	}
	if len(findings) != 1 || findings[0].ID != "1103907" || findings[0].Severity != "low" {
		t.Fatalf("unexpected findings: %v", findings)
	}
}

func TestParseBunAuditFindings_Empty(t *testing.T) {
	findings, err := ParseBunAuditFindings(strings.NewReader("{}"))
	if err != nil {
		t.Fatalf("ParseBunAuditFindings: %v", err)
	}
	if len(findings) != 0 {
		t.Fatalf("expected no findings, got %v", findings)
	}
}
