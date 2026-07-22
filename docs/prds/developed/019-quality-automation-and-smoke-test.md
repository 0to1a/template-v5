---
type: Product requirement
title: Quality automation and smoke test
description: Wire doc-lint and a dependency vulnerability scan into CI, split make check into targeted targets, add a PostgreSQL-ephemeral boot/migration/health/shutdown smoke test, and close the two untested platform packages.
tags: [ci, testing, tooling, security]
problem_brief: waiver
waiver_owner: Founding Engineer
waiver_reason: Internal CI/quality infrastructure for template-v5 itself, directed by the approved ALV-7 hardening plan (Fase 4 / Wave D); not a customer-facing capability requiring product-market evidence.
waiver_expires: 2026-08-22
---

# Quality automation and smoke test

## Purpose
Make the failure modes the ALV-7 plan calls out for Fase 4 — broken docs, unreviewed dependency vulnerabilities, and a broken boot/migration/shutdown path — actually fail CI instead of relying on manual review.

## Acceptance:
- `make doc-lint` also reports (a) a PRD whose `problem_brief` is neither a link to a file with `status: proceed` nor a complete, unexpired waiver (`waiver_owner`/`waiver_reason`/`waiver_expires`), and (b) a `docs/prds/developed/` PRD with a `TC-<id>-n` that does not appear in any test file in the repository; `doc-lint` is now a step in `make check` and CI.
- A new `make vuln-scan` target runs a Go vulnerability scan (`govulncheck`) and a JS advisory scan (`bun audit`) and fails when either reports a finding at or above a documented severity threshold that is not listed, by ID, in a checked-in exception file (`security/vulnerability-exceptions.txt`, each entry naming an owner, reason, and expiry date); an expired exception no longer suppresses its finding. `vuln-scan` is a step in `make check` and CI.
- `make check` is decomposed into independently runnable targets (`doc-lint`, `vuln-scan`, `lint`, `test`, `build`) so a developer can run one targeted check while iterating; `make check` still runs all of them and remains the single done-signal/CI gate.
- A new `make smoke-test` target builds `bin/server`, boots it against a throwaway PostgreSQL container it starts and tears down itself, polls `/health` until it answers 200 (proving startup, migration, and the health route all worked), sends the process a stop signal, and fails if the process does not exit within a bounded time; CI runs this on every PR and push to `main`.
- `internal/platform/server` (SPA routing) and `internal/platform/config` (env-based config loading) — currently zero test coverage — each get table-driven tests covering their existing branches.

## Out of Scope
- Any change to `cmd/server/main.go`'s graceful-shutdown behavior itself, or to the `newHTTPServer`/`internal/platform/server.Run` lifecycle code — that is Fase 3 / Wave C (ALV-14) scope; this PRD's smoke test only exercises whatever shutdown behavior Wave C ships (today: process termination on SIGTERM), and must be re-verified, not re-implemented, once Wave C lands.
- CVSS-equivalent severity grading for Go findings: the Go vulnerability database does not publish CVSS scores, so every `govulncheck`-confirmed (called, not merely imported) finding is treated as at-threshold; only the JS/`bun audit` scan uses graded severity (`low`/`moderate`/`high`/`critical`).
- Adding `golang.org/x/vuln` as anything other than a `go.mod` `tool` dependency (mirrors the existing `buf`/`sqlc` pattern) — no other new Go or JS dependency is introduced; the exceptions file uses this repo's existing hand-rolled key/value block parsing style, not a new YAML/TOML library.
- Splitting the GitHub Actions `check` job into multiple jobs — the required-status-check name `check` stays a single job (see `.github/workflows/ci.yml`); only the `Makefile` targets are decomposed.
- Rewriting or renumbering PRDs 001–009, or retroactively adding traceability/backlinks to them. The `problem_brief` backlink check only validates the field when a PRD declares one at all (mirrors `doclint`'s existing precedent of never retroactively failing pre-existing docs against a newer, stricter convention — see `requiredFrontMatterFields` in `internal/platform/doclint/doclint.go`); PRDs 001–007 predate the problem-brief gate and have no such field, so they are silently unaffected. Likewise, the TC-traceability check only applies to developed PRDs numbered 014 and above — PRDs 002 and 003 (`docs/prds/developed/`) predate this convention and are validated by their CI pipeline's own behavior rather than a grep-able test, not by new automated tests added retroactively here.

## Test Cases
### TC-019-1: A PRD with a broken problem_brief backlink is reported
- Given a PRD file whose `problem_brief` links to a file that does not have `status: proceed`
- When `doc-lint` runs
- Then it reports that PRD file and the invalid backlink, and exits non-zero

### TC-019-2: An expired waiver is reported
- Given a PRD file with `problem_brief: waiver` and a `waiver_expires` date in the past
- When `doc-lint` runs
- Then it reports that PRD file as having an expired waiver, and exits non-zero

### TC-019-3: A developed PRD with an untraced test case is reported
- Given a file under `docs/prds/developed/` containing `TC-example-1` with no matching `TC-example-1` string in any test file in the repository
- When `doc-lint` runs
- Then it reports the missing trace for `TC-example-1`, and exits non-zero

### TC-019-4: A vulnerability at or above threshold without an exception fails the scan
- Given a scanner finding with an ID that has no entry in `security/vulnerability-exceptions.txt`
- When the vulnerability scan's evaluation runs
- Then that finding is reported as failing

### TC-019-5: An unexpired exception suppresses a finding, an expired one does not
- Given a scanner finding whose ID has an exception entry, once with `expires` in the future and once with `expires` in the past
- When the vulnerability scan's evaluation runs for each case
- Then the future-dated exception suppresses the finding and the past-dated one does not

### TC-019-6: Targeted make targets exist and `make check` composes them
- Given the `Makefile`
- When `make doc-lint`, `make vuln-scan`, `make lint`, `make test`, and `make build` are each run individually
- Then each completes independently, and `make check` runs all of them and exits non-zero if any step fails

### TC-019-7: Smoke test proves boot, migration, health, and shutdown
- Given a freshly built `bin/server` and a throwaway PostgreSQL container with no pre-existing schema
- When the smoke test starts the server against that container, polls `/health`, then sends a stop signal
- Then `/health` answers 200 only after migrations have applied, and the server process exits within the smoke test's bounded wait

### TC-019-8: SPA handler and config table-driven tests
- Given `internal/platform/server.NewSPAHandler` and `internal/platform/config.Load` with a table of request/environment cases
- When each case runs
- Then the handler/loader's observed behavior (status code, served content, or returned error) matches the expected outcome for that case
