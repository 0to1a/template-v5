---
type: Worked example
title: "Worked vertical-slice example: initial email OTP login"
description: One real, already-developed PRD traced through proto, migration, backend, frontend, and tests, as a map for where to add each layer of a new capability.
tags: [architecture, example, prds]
status: active
owner: Founding Engineer
last_reviewed: 2026-07-22
---

# Worked vertical-slice example: initial email OTP login

This walks [`prds/developed/001-initial-auth.md`](prds/developed/001-initial-auth.md)
through every layer it touched, in the order [`prds/README.md`](prds/README.md)'s
"Map one vertical slice" step describes. Use it as a template for where a
new capability's pieces belong — skip whichever layers your feature
doesn't need.

> This PRD predates the `problem_brief` front-matter requirement introduced
> in Fase 1 of the ALV-7 hardening plan, so it has no linked brief. For the
> problem-brief → PRD hand-off format itself, see the synthetic example at
> [`product/examples/feedback-triage-problem-brief.md`](product/examples/feedback-triage-problem-brief.md)
> and [`product/examples/feedback-triage-prd.md`](product/examples/feedback-triage-prd.md)
> — synthetic evidence only, never cite it as real validation.

## 1. PRD
[`docs/prds/developed/001-initial-auth.md`](prds/developed/001-initial-auth.md) —
Purpose, Acceptance, Out of Scope, and `TC-001-1`..`TC-001-5`.

## 2. Contract
`proto/auth/v1/auth.proto` defines the `AuthService` RPCs (`RequestLogin`,
`SubmitLogin`) the Acceptance criteria describe.

## 3. Generated stubs
`make gen` produces `internal/gen/auth/v1` (Go server stubs) and
`web/src/lib/gen` (TypeScript client stubs) from that proto file. Neither is
hand-edited.

## 4. Migration
`db/migrations/00001_users.sql` creates the `users` table this feature
authenticates against. It is embedded via `db/embed.go` and applied
up-only at server startup.

## 5. Backend domain package
`internal/auth/`:
- `handler.go` implements the generated Connect service interface.
- `service.go` holds the login/OTP business logic.
- `repository.go` wraps the sqlc-generated queries built from
  `db/queries/auth.sql`.
- `otp.go` and `jwt.go` are the OTP and token pieces the service composes.
- `interceptor.go` is the protected-by-default auth interceptor described in
  [`architecture.md`](architecture.md).

## 6. Registration
`cmd/server/register_auth.go` mounts the `AuthService` Connect handler;
`cmd/server/main.go` calls it once and lists `RequestLogin`/`SubmitLogin` in
its `publicProcedures` allowlist, since a fresh visitor has no token yet.

## 7. Frontend
`web/src/routes/login/+page.svelte` and
`web/src/routes/login/otp/+page.svelte` call the generated client through
`web/src/lib/client.ts`; the issued token is stored only through
`web/src/lib/auth.ts`, matching Acceptance's "central auth module" wording.

## 8. Tests
`internal/auth/*_test.go` implement `TC-001-1`..`TC-001-5` (health handler,
non-enumerating login, the seeded local OTP, incorrect-OTP rejection, and
token storage through the central module). The PRD's Out of Scope
explicitly excludes browser end-to-end tests, so frontend coverage stops at
what the Go-side tests and the acceptance criteria require.

## Using this as a template

For a new capability, follow [`prds/README.md`](prds/README.md)'s workflow:
write the PRD first, then touch only the layers above that the PRD's
Acceptance actually requires — most small PRDs need far fewer than all
eight.
