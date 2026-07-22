---
type: Runbook
title: "Incident response"
description: "First steps when a product built from this template is down, degraded, or has a security/data incident, and what to check first using this template's own structured logs and health endpoints."
tags: [runbook, incident, operations, security]
status: active
owner: Founding Engineer
last_reviewed: 2026-07-22
---

# Incident response

## Purpose
Give whoever is on call a starting checklist for the first minutes of an
incident, before deeper investigation, using the diagnostics this template
already ships (structured logs, correlation IDs, health endpoints).

## Preconditions
- Access to the server's logs (structured JSON, one line per request —
  `docs/environment-contract.md`, "Logging").
- Access to `GET /health` and `GET /health/ready` for the affected
  environment.

## Steps
1. **Triage liveness vs. readiness first.**
   - `GET /health` failing (or not responding) — the process itself is
     down; this is a process/infrastructure incident, not an
     application-logic one. Check the platform's process status
     (restarted? crash-looping? out of memory?).
   - `GET /health` OK but `GET /health/ready` failing — the process is up
     but cannot reach the database; check database connectivity/status
     before looking at application code at all.
   - Both OK but a specific feature is broken — the incident is in
     application logic; proceed to step 2.
2. **Find the affected requests.** Every request logs one structured line
   with a correlation ID (`request_id`), method, path, status, and
   duration. Filter logs for non-2xx `status` values in the incident
   window, or for a specific `request_id` reported by an affected user (it
   is also returned to the client on the `X-Request-Id` response header, so
   ask the reporter for it).
3. **Check for a recent release or migration.** Compare the incident's
   start time against the last deploy (`release.md`) and the last migration
   applied (`migration-rollback.md` step 1) — a large share of incidents
   correlate with one of these.
4. **If data was corrupted or lost**, stop making further writes to the
   affected rows/tables and follow `backup-restore.md` rather than
   improvising a fix live against production.
5. **If this looks like a security incident** (suspected credential
   compromise, unexpected data access, unauthorized login), escalate
   immediately per `AGENTS.md` ("Escalate immediately for security/privacy
   incidents") before attempting remediation — do not rotate
   `JWT_SECRET` or lock out users unilaterally without the founder's
   awareness, since rotating `JWT_SECRET` invalidates every issued token
   and every user's derived TOTP secret for every user, not just the
   affected one.

## Verification
- `GET /health` and `GET /health/ready` both return 200 again.
- The specific user-reported symptom is confirmed resolved, not just the
  health endpoints.
- A short incident note is recorded: what broke, when it started (from
  logs), what fixed it, and the follow-up (a PRD if it needs a behavior
  change, an ADR if it needs an architecture decision).

## Rollback
See `release.md`'s Rollback section for reverting a bad deploy, and
`migration-rollback.md` for a bad migration. This runbook itself has no
destructive action to undo.

## Owner / escalation
Whoever is on call for the affected product. Security/privacy incidents,
data loss, and critical downtime escalate to the founder immediately
regardless of who is on call (`AGENTS.md`).

## Test case trace
Not implemented in code — this runbook is a documented human procedure. See
`docs/prds/developed/013-fase-3-operational-documentation.md`'s Acceptance.
