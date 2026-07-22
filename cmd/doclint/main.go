// Command doclint is the thin, impure shell behind `make doc-lint`: it
// hands the docs/ tree to internal/platform/doclint for classification.
// It only reads files; it never writes anything.
package main

import (
	"fmt"
	"os"

	"project/internal/platform/doclint"
)

func main() {
	if !run() {
		os.Exit(1)
	}
}

func run() bool {
	issues, err := doclint.Lint("docs")
	if err != nil {
		fmt.Fprintf(os.Stderr, "doc-lint: %v\n", err)
		return false
	}

	if len(issues) == 0 {
		fmt.Println("✓ doc-lint: no issues found")
		return true
	}

	for _, issue := range issues {
		fmt.Printf("✗ %s: %s\n", issue.File, issue.Message)
	}
	return false
}
