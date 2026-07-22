---
type: Product requirement
title: Fase 3 operational documentation and container example
description: Threat model, environment contract, migration/rollback, backup/restore, release, and incident-response runbooks, plus a reproducible vendor-neutral container image, for any product built from this template.
tags: [documentation, operations, security, deployment]
problem_brief: waiver
waiver_owner: Founding Engineer
waiver_reason: Internal operations documentation for template-v5 itself, directed by the approved ALV-7 hardening plan (Fase 3 / Wave C); not a customer-facing capability requiring product-market evidence.
waiver_expires: 2026-08-22
---

# Fase 3 operational documentation and container example

## Purpose
Give an operator running a product built from this template a written threat model, a full environment-variable contract, and step-by-step runbooks for the operations that are otherwise tribal knowledge: migration rollback, backup/restore, release, and incident response — plus one reproducible container image to build and run it anywhere.

## Acceptance:
- `docs/threat-model.md` names the trust boundaries (browser, server, PostgreSQL, SMTP), the assets worth protecting (`JWT_SECRET`, user emails, login codes, `DATABASE_URL`), and every currently-accepted risk already called out in code comments (bearer token in `localStorage`, the seeded `admin@localhost` static login code, TOTP replay within its step), each linked to the backlog PRD that would close it.
- `docs/environment-contract.md` documents every environment variable `config.Load` reads today, its required/optional status, its default, and what an operator should set it to in production, without inventing any variable that doesn't exist in code yet.
- `docs/runbooks/migration-rollback.md`, `docs/runbooks/backup-restore.md`, `docs/runbooks/release.md`, and `docs/runbooks/incident-response.md` each follow `docs/runbooks/TEMPLATE.md` and give exact, copy-pasteable commands with a verification step.
- A root `Dockerfile` builds a reproducible, vendor-neutral (no Railway-specific config) production image via multi-stage build (frontend build, Go build, minimal runtime), documented by `docs/runbooks/container-deployment.md`.
- `docs/README.md` and `docs/repository-map.md` are updated to list every new file, and `make doc-lint` reports zero issues against the resulting `docs/` tree.

## Out of Scope
- Any change to application behavior — this PRD is documentation and a container image only.
- Automating backup/restore or release/rollback (e.g. a script or CI job) — the runbooks are manual, human-run procedures for this phase.
- Railway-specific or Kubernetes-specific deployment manifests — the container example stays vendor-neutral per the approved plan.
- The four sensitive runtime-security capabilities (demo-credential guardrail, login throttling, OTP replay prevention, cookie hardening) — those are separate PRDs gated on owner approval, only referenced here by link.

## Test Cases
### TC-013-1: Threat model links every accepted risk to a tracking PRD
- Given `docs/threat-model.md`
- When each "accepted risk" entry is checked
- Then it links to a real file under `docs/prds/backlog/` or `docs/prds/developed/`

### TC-013-2: Environment contract matches the code
- Given `docs/environment-contract.md` and `internal/platform/config/config.go`
- When every variable name in the doc is checked against `os.Getenv` calls in the code
- Then every documented variable exists in code and no variable read in code is undocumented

### TC-013-3: doc-lint reports zero issues
- Given the full `docs/` tree after this change
- When `make doc-lint` runs
- Then it exits zero with no reported issues

### TC-013-4: The container image builds
- Given the root `Dockerfile`
- When `docker build .` runs (or, if Docker is unavailable in the environment, when the Dockerfile is reviewed step by step against `make build`'s own steps)
- Then every stage corresponds to a real, working step of this repository's own build (`bun run build`, `go build ./cmd/server`)
