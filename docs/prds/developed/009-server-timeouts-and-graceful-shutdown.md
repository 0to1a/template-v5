---
type: Product requirement
title: HTTP server timeouts and graceful shutdown
description: Bound every connection with http.Server timeouts and drain in-flight requests on SIGTERM/SIGINT instead of dropping them.
tags: [runtime, reliability, operations]
problem_brief: waiver
waiver_owner: Founding Engineer
waiver_reason: Internal runtime hardening for template-v5 itself, directed by the approved ALV-7 hardening plan (Fase 3 / Wave C); not a customer-facing capability requiring product-market evidence.
waiver_expires: 2026-08-22
---

# HTTP server timeouts and graceful shutdown

## Purpose
Stop a slow client or a deploy restart from hanging connections indefinitely or dropping in-flight requests, for the operator running any product built from this template.

## Acceptance:
- `cmd/server` constructs its `http.Server` with non-zero `ReadHeaderTimeout`, `ReadTimeout`, `WriteTimeout`, and `IdleTimeout` — no unbounded connection is possible.
- On SIGINT or SIGTERM, the server stops accepting new connections and lets in-flight requests finish before the process exits.
- If in-flight requests do not finish within a bounded drain deadline, the server force-closes and exits rather than hanging forever.
- The database pool is closed only after the HTTP server has finished shutting down, so a request that is still draining can still reach the database.

## Out of Scope
- Per-route timeout overrides or configurable timeout values via environment variables — the fixed constants in this PRD are the template default.
- Kubernetes/Railway-specific readiness/termination-grace-period wiring (covered by `docs/environment-contract.md` and the container-deployment runbook, not code).
- Any change to `internal/auth` or other domain business logic.

## Test Cases
### TC-009-1: Server is constructed with non-zero timeouts
- Given `newHTTPServer` builds the production `*http.Server`
- When its fields are inspected
- Then `ReadHeaderTimeout`, `ReadTimeout`, `WriteTimeout`, and `IdleTimeout` are all greater than zero

### TC-009-2: SIGTERM drains an in-flight request before exiting
- Given a server running a handler that is mid-request when a stop signal fires
- When the stop signal arrives
- Then `Shutdown` is invoked, the in-flight handler completes, and `Run` returns nil once it has finished

### TC-009-3: An in-flight request that exceeds the drain deadline is force-closed
- Given a handler that blocks longer than the configured shutdown timeout
- When the stop signal arrives
- Then `Run` returns once the shutdown deadline elapses instead of blocking forever
