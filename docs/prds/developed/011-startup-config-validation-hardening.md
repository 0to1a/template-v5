---
type: Product requirement
title: Startup config validation hardening without secret logging
description: Reject a malformed PORT or DATABASE_URL at startup with a clear error, and give the server exactly one safe way to log its own configuration that never includes a secret value.
tags: [runtime, configuration, security]
problem_brief: waiver
waiver_owner: Founding Engineer
waiver_reason: Internal runtime hardening for template-v5 itself, directed by the approved ALV-7 hardening plan (Fase 3 / Wave C); not a customer-facing capability requiring product-market evidence.
waiver_expires: 2026-08-22
---

# Startup config validation hardening without secret logging

## Purpose
Fail fast and loud on a broken `PORT` or `DATABASE_URL` instead of surfacing a confusing runtime error later, and give the server one sanctioned, secret-free way to report its own configuration at startup.

## Acceptance:
- `config.Load` rejects a non-empty `PORT` that is not a valid TCP port number (1–65535) with an error naming the problem, and never a numeric error that echoes a secret.
- `config.Load` rejects a `DATABASE_URL` that is not a well-formed `postgres://` or `postgresql://` URL, without including the raw `DATABASE_URL` value (which may embed a password) in the returned error.
- `Config` exposes a safe-fields accessor listing only non-secret configuration (e.g. port, whether mail delivery and guest registration are enabled) for startup logging, so nothing has to (and nothing should) print the `Config` struct directly.
- Existing fail-closed behavior for a missing/short `JWT_SECRET` and a missing `DATABASE_URL` is unchanged.

## Out of Scope
- Moving `MAIL_URL` parsing (already validated in `internal/mail`) into this package.
- Any new environment variable or environment/mode concept (tracked separately as a sensitive, owner-approval-gated PRD).
- Changing the default `PORT` value or adding CLI flags.

## Test Cases
### TC-011-1: A non-numeric PORT is rejected
- Given `PORT=abc`
- When `config.Load` runs
- Then it returns an error naming `PORT` as invalid

### TC-011-2: An out-of-range PORT is rejected
- Given `PORT=70000`
- When `config.Load` runs
- Then it returns an error naming `PORT` as invalid

### TC-011-3: A malformed DATABASE_URL is rejected without leaking its value
- Given `DATABASE_URL=not-a-url`
- When `config.Load` runs
- Then it returns an error that does not contain the string `not-a-url`

### TC-011-4: Safe fields never include the JWT secret or the raw database URL
- Given a loaded `Config` with a real `JWT_SECRET` and `DATABASE_URL`
- When its safe-fields accessor is inspected
- Then neither the JWT secret value nor the raw `DATABASE_URL` value appears anywhere in it
