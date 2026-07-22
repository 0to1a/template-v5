---
type: Product requirement
title: HttpOnly session/cookie hardening
description: Move the bearer token from browser localStorage to an HttpOnly, Secure, SameSite cookie so an in-origin XSS can no longer read it directly.
tags: [auth, security, sensitive, frontend]
problem_brief: waiver
waiver_owner: Founding Engineer
waiver_reason: Internal security hardening for template-v5 itself, directed by the approved ALV-7 hardening plan (Fase 3 / Wave C); not a customer-facing capability requiring product-market evidence.
waiver_expires: 2026-08-22
---

# HttpOnly session/cookie hardening

## Owner approval required before implementation
This PRD changes `internal/auth` and `web/src/lib/auth.ts` — the one
sanctioned place the bearer token is read/written per `AGENTS.md` ("the
token only moves through `web/src/lib/auth.ts`"). Per `AGENTS.md`,
"Sensitive behavior (auth, authorization...) requires owner approval before
implementation." **Do not implement any of the Acceptance below until the
owner explicitly approves this PRD.** `web/src/lib/auth.ts`'s own comment
already flags this exact risk: "moving to HttpOnly cookies is its own PRD,
not a silent change" — this is that PRD.

## Open questions for the owner
- CSRF strategy once the token moves out of a header the frontend attaches
  explicitly: `SameSite=Strict`/`Lax` alone, or an additional double-submit
  CSRF token for state-changing Connect procedures.
- Whether Connect-ES's browser client can attach/receive cookies cleanly
  with the existing `web/src/lib/client.ts` transport config, or whether
  the transport needs `credentials: 'include'` plus CORS changes if the
  frontend and API are ever served from different origins (today they are
  not — one binary serves both, per `docs/architecture.md`).
- Token/cookie expiry and refresh behavior — unchanged 24h expiry, or
  revisited alongside this change.

## Purpose
Stop an in-origin XSS from being able to read the session token directly,
since today it is deliberately stored in `localStorage`
(`web/src/lib/auth.ts`), which is readable by any script running on the
page.

## Acceptance:
- The server issues the session token as an `HttpOnly`, `Secure`, and
  `SameSite`-restricted cookie instead of a value the frontend reads and
  attaches itself.
- `web/src/lib/auth.ts` no longer reads or writes the token via
  `localStorage`; it becomes the one place that reasons about
  authentication state (e.g. "am I logged in") without ever holding the
  token value itself.
- State-changing requests remain protected against CSRF (mechanism decided
  during owner approval — see Open questions).
- Logout clears the cookie server-side (`Set-Cookie` with an expired
  value), not just client-side state.

## Out of Scope
- Refresh tokens or a token-rotation scheme — only the storage/transport
  mechanism changes, not the token's own lifecycle, unless the owner
  decides otherwise during approval.
- The other three sensitive PRDs (demo-credential guardrail, login
  throttling/lockout, OTP replay prevention).

## Test Cases
### TC-017-1: The session cookie is HttpOnly and Secure
- Given a successful `SubmitLogin`
- When the response is inspected
- Then the session cookie's `Set-Cookie` attributes include `HttpOnly` and
  `Secure`

### TC-017-2: No token is readable from browser-side JavaScript
- Given a successful login
- When `web/src/lib/auth.ts` and `localStorage` are inspected from the page
- Then no token value is present in either

### TC-017-3: Logout clears the cookie
- Given a logged-in session
- When logout runs
- Then the response clears the session cookie and a subsequent authenticated
  request is rejected
