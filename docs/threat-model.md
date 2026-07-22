---
type: Security
title: Threat model
description: Trust boundaries, protected assets, and every currently-accepted security risk in the template, each linked to the PRD that would close it.
tags: [security, architecture, operations]
status: active
owner: Founding Engineer
last_reviewed: 2026-07-22
---

# Threat model

This documents what this template protects, from what, and the risks it
already knowingly accepts — so "is this safe to deploy" has one answer
instead of a repository-wide search. See
[`environment-contract.md`](environment-contract.md) for the configuration
side of the same posture, and [`architecture.md`](architecture.md) for how a
request actually flows.

## Trust boundaries

```text
Browser  <-- untrusted -->  Go server  <-- trusted network -->  PostgreSQL
                                |
                                +-- trusted network --> SMTP (MAIL_URL)
```

- **Browser → server**: the only boundary that carries attacker-controlled
  input. Every Connect procedure is protected by default
  (`internal/auth/interceptor.go`); only the two login procedures are
  explicitly public.
- **Server → PostgreSQL**: assumed trusted network (same host, private
  network, or TLS-terminated connection string). The server never creates,
  drops, or resets the database — see `AGENTS.md`.
- **Server → SMTP**: assumed trusted network. `MAIL_URL` is optional; when
  unset, login codes are discarded rather than sent (`internal/auth/delivery.go`).

## Assets worth protecting

| Asset | Where it lives | Why it matters |
|---|---|---|
| `JWT_SECRET` | environment variable, process memory | Signs every bearer token and derives every user's TOTP secret (`internal/auth/otp.go`) — compromise forges both. |
| `DATABASE_URL` | environment variable, process memory | May embed a database password; grants full read/write on user data. |
| Login codes (OTP) | generated in-memory, emailed, never stored | A valid code plus a known email is equivalent to a password. |
| User email addresses | `users` table | PII; also the login identifier. |
| Bearer tokens | issued to the browser, stored in `localStorage` | Grants the holder full access as that user for the token's lifetime (24h). |

## Currently-accepted risks

Each of these is a real, present risk the code already documents; none is
implemented differently today because doing so is sensitive behavior that
needs owner approval first (see `AGENTS.md`, "Sensitive behavior... requires
owner approval before implementation"). Each links to the backlog PRD that
would close it.

- **Bearer token in `localStorage`, not an `HttpOnly` cookie**
  (`web/src/lib/auth.ts`): any XSS that runs in the app's origin can read
  the token. Tracked by
  [`prds/backlog/017-httponly-session-cookie-hardening.md`](prds/backlog/017-httponly-session-cookie-hardening.md).
- **A seeded static login code for `admin@localhost`**
  (`internal/auth/otp.go`, `db/migrations/00001_users.sql`): a fixed,
  publicly-known credential (`123456`) that must never be reachable outside
  local development. Tracked by
  [`prds/backlog/014-production-guardrail-disable-demo-login-credential.md`](prds/backlog/014-production-guardrail-disable-demo-login-credential.md).
- **TOTP replay within its 5-minute step**
  (`internal/auth/otp.go`): a valid code can be reused any number of times
  within the step it was issued for; there is no replay-tracking store.
  Tracked by
  [`prds/backlog/016-otp-replay-prevention.md`](prds/backlog/016-otp-replay-prevention.md).
- **No login throttling or lockout**
  (`internal/auth/service.go`): `SubmitLogin` does not rate-limit or lock
  out repeated failed attempts, so a fixed-length code space (6 digits) is
  brute-forceable given enough attempts within one TOTP step. Tracked by
  [`prds/backlog/015-login-throttling-and-lockout.md`](prds/backlog/015-login-throttling-and-lockout.md).
- **SMTP-only email delivery, no provider abstraction**
  (`internal/mail`): a single hardcoded transport with no failover,
  deliverability monitoring, or provider-side abuse controls. Tracked by
  [`prds/backlog/018-email-provider-abstraction.md`](prds/backlog/018-email-provider-abstraction.md).

## Explicitly out of scope for this template

- Network-layer DDoS mitigation and WAF — handled at the Cloudflare edge per
  the standard stack, not application code.
- Multi-tenant isolation, roles, or billing — not part of the template's
  core (see the Fase 6 golden-path plan); a venture adds these only after a
  cross-venture pattern is validated.

## Revisiting this document

Update this file whenever a PRD closes one of the accepted risks above
(move its bullet out) or whenever a new domain introduces a new trust
boundary or asset.
