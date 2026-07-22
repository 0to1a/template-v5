// Command vulnscan is the thin, impure shell behind `make vuln-scan`: it
// runs govulncheck (Go) and `bun audit` (JS), hands their output to
// internal/platform/vulnscan for evaluation against the checked-in
// exception file, and reports every finding at or above threshold that
// has no unexpired exception. It never edits go.mod/go.sum or
// package.json/bun.lock.
package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"time"

	"project/internal/platform/vulnscan"
)

func main() {
	exceptionsPath := flag.String("exceptions", "security/vulnerability-exceptions.txt", "path to the vulnerability exceptions file")
	threshold := flag.String("threshold", vulnscan.SeverityHigh, "minimum severity that fails the scan (low, moderate, high, critical)")
	flag.Parse()

	if !run(*exceptionsPath, *threshold) {
		os.Exit(1)
	}
}

func run(exceptionsPath, threshold string) bool {
	exceptionsText, err := os.ReadFile(exceptionsPath)
	if err != nil && !os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "vuln-scan: reading %s: %v\n", exceptionsPath, err)
		return false
	}
	exceptions, err := vulnscan.ParseExceptions(string(exceptionsText))
	if err != nil {
		fmt.Fprintf(os.Stderr, "vuln-scan: %v\n", err)
		return false
	}

	goFindings, err := runGovulncheck()
	if err != nil {
		fmt.Fprintf(os.Stderr, "vuln-scan: govulncheck: %v\n", err)
		return false
	}
	jsFindings, err := runBunAudit()
	if err != nil {
		fmt.Fprintf(os.Stderr, "vuln-scan: bun audit: %v\n", err)
		return false
	}

	now := time.Now()
	goOK := report("Go (govulncheck)", goFindings, exceptions, threshold, now)
	jsOK := report("JS (bun audit)", jsFindings, exceptions, threshold, now)
	return goOK && jsOK
}

// runGovulncheck scans every package for vulnerabilities whose code is
// actually reachable (govulncheck's default "symbol" scan level). It
// exits non-zero when it finds any, which is expected and not itself an
// error; any error other than that expected exit status is.
func runGovulncheck() ([]vulnscan.Finding, error) {
	cmd := exec.Command("go", "tool", "govulncheck", "-json", "./...")
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		if _, ok := err.(*exec.ExitError); !ok {
			return nil, err
		}
	}
	return vulnscan.ParseGovulncheckFindings(&stdout)
}

// runBunAudit checks web/'s installed dependencies against the npm
// advisory database. `bun audit --json` writes its banner to stderr and
// pure JSON to stdout, and exits non-zero when it finds any advisory,
// which (like govulncheck above) is expected, not itself an error.
func runBunAudit() ([]vulnscan.Finding, error) {
	cmd := exec.Command("bun", "audit", "--json")
	cmd.Dir = "web"
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		if _, ok := err.(*exec.ExitError); !ok {
			return nil, err
		}
	}
	return vulnscan.ParseBunAuditFindings(&stdout)
}

func report(ecosystem string, findings []vulnscan.Finding, exceptions []vulnscan.Exception, threshold string, now time.Time) bool {
	failing := vulnscan.Evaluate(findings, exceptions, threshold, now)
	if len(failing) == 0 {
		fmt.Printf("✓ vuln-scan: %s: no findings at or above %q without an unexpired exception\n", ecosystem, threshold)
		return true
	}
	for _, f := range failing {
		fmt.Printf("✗ vuln-scan: %s: %s (severity %s) has no unexpired exception\n", ecosystem, f.ID, f.Severity)
	}
	return false
}
