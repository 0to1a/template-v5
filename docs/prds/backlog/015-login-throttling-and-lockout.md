---
type: Product requirement
title: Login throttling and lockout after repeated failed attempts
description: Bound how many times an attacker can try a login code per account before being throttled or locked out, without creating an account-existence oracle.
tags: [auth, security, sensitive]
problem_brief: waiver
waiver_owner: Founding Engineer
waiver_reason: Internal security hardening for template-v5 itself, directed by the approved ALV-7 hardening plan (Fase 3 / Wave C); not a customer-facing capability requiring product-market evidence.
waiver_expires: 2026-08-22
---

# Login throttling and lockout after repeated failed attempts

## Owner approval required before implementation
This PRD changes `internal/auth` login behavior. Per `AGENTS.md`, "Sensitive
behavior (auth, authorization, money, deletion, destructive migration)
requires owner approval before implementation." **Do not implement any of
the Acceptance below until the owner explicitly approves this PRD.** See
`docs/threat-model.md` for the risk this closes.

## Open questions for the owner
This PRD deliberately does not invent these business rules — they need an
explicit decision before Acceptance can be finalized:
- The failure threshold and window (e.g. "5 failed attempts in 5 minutes")
  and the throttle/lockout duration.
- Whether counters are per-account only, or also per-source-IP (per-IP adds
  protection against distributed low-and-slow attempts but risks throttling
  shared corporate/NAT IPs).
- Where counters are stored (in-memory doesn't survive a restart or scale
  past one instance; database-backed does but adds a write on every login
  attempt).

## Purpose
Stop a 6-digit login code from being brute-forceable within its validity
window, since today `SubmitLogin` (`internal/auth/service.go`) does not
limit the number of attempts.

## Acceptance:
- After the owner-decided failure threshold is reached for one account
  within the owner-decided window, further `SubmitLogin` attempts for that
  account are rejected without attempting verification, using the same
  generic error as an invalid code (no distinct "locked out" signal that
  would reveal an account exists).
- A successful login resets that account's failure counter.
- Throttle/lockout state for one account never affects another account's
  ability to log in.
- The mechanism does not introduce an unbounded-growth data structure (e.g.
  an expiring/bounded store, not an ever-growing in-memory map).

## Out of Scope
- Notifying the user or an operator of a lockout event (could be a
  follow-up PRD).
- IP-based throttling, unless the owner decides it belongs in this PRD's
  Acceptance during approval.
- The other three sensitive PRDs (demo-credential guardrail, OTP replay
  prevention, cookie hardening).

## Test Cases
### TC-015-1: Failure threshold blocks further attempts
- Given an account has failed login the owner-decided threshold number of
  times within the window
- When one more `SubmitLogin` attempt is made with any code
- Then it is rejected with the same generic error as an invalid code

### TC-015-2: A successful login resets the counter
- Given an account has some failed attempts below the threshold
- When a subsequent attempt succeeds
- Then the failure counter is reset and the account is not throttled

### TC-015-3: Throttling is isolated per account
- Given account A is throttled
- When account B attempts to log in with a correct code
- Then account B's login succeeds
