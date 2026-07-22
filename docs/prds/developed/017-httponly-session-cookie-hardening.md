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

## Owner approval
This PRD changes `internal/auth` and `web/src/lib/auth.ts`, so per `AGENTS.md`
it required owner approval before implementation. The owner approved
implementation via the `request_confirmation` interaction on ALV-14
("Approve 5 sensitive auth/security PRDs before Fase 3 sensitive work is
coded"), accepted 2026-07-22, and answered the CSRF-strategy question below
via the `ask_user_questions` interaction "Open questions for sensitive
PRDs 015-018 (Wave C)", answered 2026-07-22.

## Owner decisions (resolve the former "Open questions")
- CSRF strategy: **`SameSite=Strict` alone**, no additional double-submit
  CSRF token.
- Connect-ES + this transport: confirmed working — `createConnectTransport`
  takes a `fetch` override rather than a `fetchInit` option in the installed
  version (`@connectrpc/connect-web@2.1.2`); `web/src/lib/client.ts` wraps
  `fetch` to force `credentials: 'same-origin'`, matching today's one-origin
  deployment (`docs/architecture.md`). No CORS change was needed.
- Token/cookie expiry: unchanged 24h (`accessTokenTTL` in `internal/auth/jwt.go`),
  refresh behavior untouched — out of scope, as originally proposed.

## Purpose
Stop an in-origin XSS from being able to read the session token directly,
since before this PRD it was deliberately stored in `localStorage`
(`web/src/lib/auth.ts`), which is readable by any script running on the
page.

## Acceptance
- The server issues the session token as an `HttpOnly`, `Secure`,
  `SameSite=Strict` cookie (`internal/auth/cookie.go`,
  `template_v5_session`) instead of a value the frontend reads and attaches
  itself. `SubmitLoginResponse` no longer carries the token in its body
  (the message is now empty; see `proto/auth/v1/auth.proto`).
- `web/src/lib/auth.ts` no longer reads or writes anything via
  `localStorage`. It reasons about "am I logged in" via a separate,
  non-secret `template_v5_authed` indicator cookie the server sets/clears
  alongside the session cookie — the token value itself never enters
  frontend code.
- State-changing requests are protected against CSRF by `SameSite=Strict`
  alone, per the owner's decision — no double-submit token.
- A new `Logout` RPC (`internal/auth/handler.go`) clears both cookies
  server-side (`Set-Cookie` with `Max-Age=0`), not just client-side state.
  It is public (bypasses the auth interceptor) so it still succeeds for a
  caller whose session already expired.

## Out of Scope
- Refresh tokens or a token-rotation scheme — only the storage/transport
  mechanism changed, the token's own lifecycle did not.
- Any change needed only if the frontend and API ever move to different
  origins (they don't today) — that would need `credentials: 'include'`
  plus CORS configuration, revisited together if it ever happens.
- The other three sensitive PRDs (demo-credential guardrail, login
  throttling/lockout, OTP replay prevention).

## Test Cases
### TC-017-1: The session cookie is HttpOnly and Secure
- Given a successful `SubmitLogin`
- When the response is inspected
- Then the session cookie's `Set-Cookie` attributes include `HttpOnly`,
  `Secure`, and `SameSite=Strict`

### TC-017-2: No token is readable from browser-side JavaScript
- Given a successful login
- When `web/src/lib/auth.ts`, `localStorage`, and the non-HttpOnly indicator
  cookie are inspected
- Then no token value is present in any of them — the indicator cookie
  carries only a fixed, non-secret flag

### TC-017-3: Logout clears the cookie
- Given a logged-in session (or none at all)
- When `Logout` is called
- Then the response clears both cookies (`Max-Age=0`, empty value) and a
  subsequent request presenting the old cookie value is rejected as
  unauthenticated

## Verification beyond automated tests
Manually verified end to end against a disposable Docker Postgres: started
the built binary, called `SubmitLogin` via `curl` and inspected the raw
`Set-Cookie` headers (both attributes and values), then called `Logout` with
no prior cookie and confirmed it still returns `200` and clears both
cookies. A full browser-driven check (real cookie jar via `document.cookie`,
actual `fetch` credential behavior) was not additionally performed in this
session — the `curl`-level check above already exercises the exact wire
behavior a browser would rely on; flag if a browser pass is still wanted.
