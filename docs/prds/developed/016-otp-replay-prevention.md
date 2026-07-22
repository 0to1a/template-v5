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

## Owner approval
This PRD changes `internal/auth` login behavior, so per `AGENTS.md` it
required owner approval before implementation. The owner approved
implementation via the `request_confirmation` interaction on ALV-14
("Approve 5 sensitive auth/security PRDs before Fase 3 sensitive work is
coded"), accepted 2026-07-22, and answered the storage question below via
the `ask_user_questions` interaction "Open questions for sensitive PRDs
015-018 (Wave C)", answered 2026-07-22.

## Owner decisions (resolve the former "Open questions")
- Storage: **in-memory** — one entry per account (the most recently used
  TOTP step), never per code or per attempt, so the store is bounded by
  account count and self-overwrites on every new login.
- Admin fixed-code exemption: the interaction did not carry a distinct
  answer for this sub-question, so this PRD applies replay tracking
  uniformly to every account, including the seeded `admin@localhost`
  development account — no special case. This is the simpler and more
  conservative default (one less exemption to get wrong), and its only
  practical effect is that the dev-only fixed code can complete one login
  per 5-minute step, same as every other account. Flag to the owner if a
  different dev-workflow trade-off is preferred; it's a one-line change to
  exempt it in `service.go`.

## Purpose
Ensure a login code can only complete one successful login, closing the
window where the same valid TOTP code could be reused by anyone who
intercepted it during its 5-minute step.

## Acceptance
- Once a login code has been used to complete a successful `SubmitLogin`,
  the same code is rejected for the same account for the remainder of its
  validity window.
- A code that was never successfully used remains valid for the rest of its
  window, unchanged from before this PRD.
- The replay-tracking mechanism is bounded: one entry per account
  (`internal/auth/replay.go`), overwritten on each new successful login, not
  an ever-growing structure.

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
- Then it succeeds, unchanged from before this PRD

### TC-016-3: Replay tracking does not leak across accounts
- Given account A's code has been used
- When account B submits its own, different, unused code
- Then account B's login succeeds
