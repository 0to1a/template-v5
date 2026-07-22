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

## Owner approval
This PRD changes `internal/auth` login behavior, so per `AGENTS.md` it
required owner approval before implementation. The owner approved
implementation via the `request_confirmation` interaction on ALV-14
("Approve 5 sensitive auth/security PRDs before Fase 3 sensitive work is
coded"), accepted 2026-07-22, and answered the three open business-rule
questions below via the `ask_user_questions` interaction "Open questions
for sensitive PRDs 015-018 (Wave C)", answered 2026-07-22.

## Owner decisions (resolve the former "Open questions")
- Threshold/window/lockout: **5 failed attempts within a 5-minute window,
  then a 15-minute lockout.**
- Counter scope: **per-account only** (no per-source-IP tracking).
- Storage: **in-memory** (resets on restart, single-instance only — an
  accepted trade-off, not a gap; see Out of Scope).

## Purpose
Stop a 6-digit login code from being brute-forceable within its validity
window, since before this PRD `SubmitLogin` (`internal/auth/service.go`)
did not limit the number of attempts.

## Acceptance
- After 5 failed attempts for one account within a 5-minute window, further
  `SubmitLogin` attempts for that account are rejected without attempting
  verification, for 15 minutes, using the same generic error as an invalid
  code (no distinct "locked out" signal that would reveal an account
  exists).
- A successful login resets that account's failure counter.
- Throttle/lockout state for one account never affects another account's
  ability to log in.
- The mechanism does not introduce an unbounded-growth data structure: it
  is keyed by account (`PublicUUID`), populated only after an account is
  confirmed to exist, and pruned on every successful login or expired
  window — so its size is bounded by the number of real accounts with
  recent failures, never by attacker-controlled input.

## Out of Scope
- Notifying the user or an operator of a lockout event (could be a
  follow-up PRD).
- IP-based throttling (the owner chose per-account only).
- Database-backed counters that survive a restart or scale past one
  instance (the owner chose in-memory; a future PRD can revisit this if
  the service needs to scale horizontally).
- The other three sensitive PRDs (demo-credential guardrail, OTP replay
  prevention, cookie hardening).

## Test Cases
### TC-015-1: Failure threshold blocks further attempts
- Given an account has failed login 5 times within the 5-minute window
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
