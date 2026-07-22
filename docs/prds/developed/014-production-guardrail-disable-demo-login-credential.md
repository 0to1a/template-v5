---
type: Product requirement
title: "Production guardrail: disable the static demo login credential outside development"
description: Fail closed on the seeded admin@localhost fixed login code ("123456") in any environment that isn't explicitly development.
tags: [auth, security, sensitive]
problem_brief: waiver
waiver_owner: Founding Engineer
waiver_reason: Internal security hardening for template-v5 itself, directed by the approved ALV-7 hardening plan (Fase 3 / Wave C); not a customer-facing capability requiring product-market evidence.
waiver_expires: 2026-08-22
---

# Production guardrail: disable the static demo login credential outside development

## Owner approval
This PRD changes `internal/auth` login behavior, so per `AGENTS.md` it
required owner approval before implementation. The owner approved
implementation via the `request_confirmation` interaction on ALV-14
("Approve 5 sensitive auth/security PRDs before Fase 3 sensitive work is
coded"), accepted 2026-07-22. See `docs/threat-model.md` for the risk this
closes (now removed from the currently-accepted list).

## Purpose
Stop the seeded `admin@localhost` / `123456` static login code
(`internal/auth/otp.go`) — a publicly-known credential in this template's
own source — from ever being reachable in a real deployment, for any
product built from this template.

## Acceptance:
- A new environment discriminator (proposed: `APP_ENV`, values `development`
  or `production`) is added to `internal/platform/config`. An unset or
  empty value defaults to `production` (fail closed); any value other than
  the two allowed ones is a startup error.
- When the environment is `production`, `admin@localhost` never accepts the
  fixed code `123456` — it is derived/verified through the same TOTP path
  as every other account, with no special case.
- When the environment is `development` (explicit opt-in), today's behavior
  is unchanged: the fixed code still works for local development.
- `docs/environment-contract.md` documents the new variable, and its entry
  in `docs/threat-model.md`'s "currently-accepted risks" is removed once
  this ships.

## Out of Scope
- Removing the seeded `admin@localhost` row from `db/migrations/00001_users.sql`
  (a data change, not a behavior change, and migrations are up-only).
- The other three sensitive PRDs (login throttling/lockout, OTP replay
  prevention, cookie hardening) — each is its own PRD, approved separately.
- Any per-environment behavior beyond this one guardrail.

## Test Cases
### TC-014-1: Unset APP_ENV defaults to production and disables the demo code
- Given `APP_ENV` is unset
- When `SubmitLogin` is called for `admin@localhost` with code `123456`
- Then it is rejected exactly like any other invalid code for that account

### TC-014-2: Explicit development APP_ENV preserves today's behavior
- Given `APP_ENV=development`
- When `SubmitLogin` is called for `admin@localhost` with code `123456`
- Then it succeeds, unchanged from current behavior

### TC-014-3: An unrecognized APP_ENV value fails startup
- Given `APP_ENV=staging` (or any value other than `development`/`production`)
- When `config.Load` runs
- Then it returns a startup error naming `APP_ENV`
