---
type: Runbook
title: "Release"
description: "How to ship a change from a merged PR to a running production process, and what to check before and after."
tags: [runbook, release, deployment]
status: active
owner: Founding Engineer
last_reviewed: 2026-07-22
---

# Release

## Purpose
Ship a merged change to production with a known-good verification path, and
know what "done" looks like for a deploy of this template.

## Preconditions
- The change is merged to `main` via a PR with a passing `check` CI run
  (`docs/runbooks/github-branch-protection.md`).
- `JWT_SECRET`, `DATABASE_URL`, and any other required environment variables
  are already set in the target environment (`docs/environment-contract.md`)
  — a release never introduces new required configuration silently.

## Steps
1. Confirm CI is green on the commit being released:
   ```sh
   gh pr checks <pr-number>
   ```
2. Build the production artifact the same way CI/`make check` does:
   ```sh
   make build
   ```
   or, for a container-based deploy, build the image described in
   `container-deployment.md`.
3. Deploy the new binary/image to the target environment following the
   platform's own deploy mechanism (e.g. Railway's git-push or image
   deploy). This runbook is deliberately platform-agnostic; a
   platform-specific deploy runbook may be added later per Fase 6.
4. On startup, the server applies any pending migrations automatically
   (up-only) before it starts accepting traffic — watch the startup log
   for a migration failure, which aborts startup loudly rather than
   running against an unknown schema.
5. Confirm the new process is live and ready (Verification below) before
   considering the release complete.

## Verification
- `GET /health` returns 200 (process is up).
- `GET /health/ready` returns 200 (database reachable).
- The specific behavior the release was for is confirmed working end to
  end, not just the health endpoints — the health checks only prove the
  process started, not that the shipped change works.
- No `.dump`, `.env`, JWT secret, or database URL value appears anywhere in
  deploy logs.

## Rollback
- If the new release is broken, redeploy the previous known-good
  binary/image — the server's migrations are additive/up-only, so an older
  binary can run against a schema a newer migration already applied (as
  long as the older code doesn't depend on a column/table the new migration
  removed — this template's migrations only ever add, never remove or
  rename, for exactly this reason).
- If the release included a migration that itself needs correcting, follow
  `migration-rollback.md` — do not attempt to run a migration backward.

## Owner / escalation
Owned by whoever has deploy access to the target environment for that
product. Escalate immediately for a release that causes customer-visible
downtime or data loss (`AGENTS.md`: "Escalate immediately for... critical
downtime, unexpected material cost").

## Test case trace
Not implemented in code — this runbook is a documented human procedure. See
`docs/prds/developed/013-fase-3-operational-documentation.md`'s Acceptance.
