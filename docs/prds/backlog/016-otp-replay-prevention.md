---
type: Product requirement
title: OTP replay prevention
description: Make a login code single-use, so it cannot be replayed again within its own validity window once it has succeeded once.
tags: [auth, security, sensitive]
problem_brief: waiver
waiver_owner: Founding Engineer
waiver_reason: Internal security hardening for template-v5 itself, directed by the approved ALV-7 hardening plan (Fase 3 / Wave C); not a customer-facing capability requiring product-market evidence.
waiver_expires: 2026-08-22
---

# OTP replay prevention

## Owner approval required before implementation
This PRD changes `internal/auth` login behavior. Per `AGENTS.md`, "Sensitive
behavior (auth, authorization, money, deletion, destructive migration)
requires owner approval before implementation." **Do not implement any of
the Acceptance below until the owner explicitly approves this PRD.** See
`docs/threat-model.md` for the risk this closes, and `internal/auth/otp.go`'s
own `currentTOTP` comment, which documents this as "a documented limitation,
not a bug; closing it needs a replay-tracking store this template does not
implement" — this PRD is that follow-up.

## Open questions for the owner
- Where the "already used" record is stored (a new table vs. an in-memory
  store — in-memory doesn't survive a restart or scale past one instance,
  which matters more here than for throttling since a code is only valid
  for one 5-minute step anyway).
- Whether the local admin's fixed code (`123456`) is exempt from replay
  tracking, or whether this PRD ships after
  [`014-production-guardrail-disable-demo-login-credential.md`](014-production-guardrail-disable-demo-login-credential.md)
  makes that moot outside development.

## Purpose
Ensure a login code can only complete one successful login, closing the
window where the same valid TOTP code can be reused by anyone who
intercepts it during its 5-minute step.

## Acceptance:
- Once a login code has been used to complete a successful `SubmitLogin`,
  the same code is rejected for the same account for the remainder of its
  validity window.
- A code that was never successfully used remains valid for the rest of its
  window, unchanged from today's behavior.
- The replay-tracking mechanism is bounded (records expire with the step
  they belong to) rather than growing without limit.

## Out of Scope
- Changing the TOTP step length or digit count.
- The other three sensitive PRDs (demo-credential guardrail, login
  throttling/lockout, cookie hardening).

## Test Cases
### TC-016-1: A used code cannot be replayed
- Given a login code that already completed one successful `SubmitLogin`
- When the same code is submitted again for the same account within its
  original validity window
- Then it is rejected

### TC-016-2: An unused code still works within its window
- Given a login code that has not yet been used
- When it is submitted once within its validity window
- Then it succeeds, unchanged from today

### TC-016-3: Replay tracking does not leak across accounts
- Given account A's code has been used
- When account B submits its own, different, unused code
- Then account B's login succeeds
