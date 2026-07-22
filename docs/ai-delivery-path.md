---
type: Engineering guide
title: AI delivery path
description: Minimal context and validation route for an agent implementing one approved vertical slice.
tags: [ai, delivery, workflow]
status: active
owner: Founding Engineer
---

# AI delivery path

Use this page as the single router after reading `AGENTS.md` and the approved
PRD. It compresses discovery; it does not override either source.

## Context budget

Read only:

1. `AGENTS.md`, `docs/prds/README.md`, and the approved backlog PRD;
2. `make help`;
3. one nearest sibling for each layer the PRD actually needs, using the table
   below;
4. a registration/composition file only when wiring a new domain or procedure.

Search for a symbol before opening another file. Do not inventory the repository,
read every PRD, or reload unchanged guidance. Expand context only when a concrete
compile, test, or contract question cannot be answered by the current set.

Stop discovery and implement when the PRD has no unresolved business rule and
you can name the required layers, sibling shapes, tests, and smallest targeted
check. If a business rule is ambiguous, stop before implementation and request a
PRD decision; do not infer it from sibling code.

## Vertical-slice route

Skip layers the PRD does not require. For layers it does require, work in this
order so generated contracts are available to handwritten code:

| Layer | Location | Nearest shape to inspect |
| --- | --- | --- |
| Migration | `db/migrations/` | latest numbered migration |
| SQL | `db/queries/` | `db/queries/auth.sql` |
| Contract | `proto/<domain>/v1/` | `proto/auth/v1/auth.proto` |
| Generation | generated outputs | run `make gen`; never edit generated files |
| Backend | `internal/<domain>/` | the matching file in `internal/auth/` |
| Registration | `cmd/server/register_<domain>.go` | existing `cmd/server/register_*.go` |
| Client | `web/src/lib/client.ts` | existing exported Connect client |
| UI | `web/src/routes/` | nearest route such as `web/src/routes/login/` |
| Tests | beside touched code | nearest sibling test; include `TC-<id>-n` |

Preserve these invariants:

- Connect procedures are protected unless explicitly allowlisted in
  `cmd/server/main.go`.
- Authorization and resource scope are enforced server-side, not only in UI.
- SQL stays in `db/queries/`; migrations are embedded, up-only, and never
  create, drop, or reset PostgreSQL implicitly.
- Generated Go/TypeScript files come only from `make gen`.
- Frontend API calls use the generated Connect client via
  `web/src/lib/client.ts`; auth tokens move only through
  `web/src/lib/auth.ts`.
- Secrets, OTPs, JWTs, and database URLs are never logged.
- Svelte components remain HTML-first and use Svelte 5 runes.

## Validation ladder

Validate at the cheapest relevant level while iterating, then run the full gate
once when the candidate is complete:

1. After SQL/proto edits: `make gen`, then inspect `git status --short` to catch
   unexpected generated changes.
2. Backend iteration: `go test ./internal/<domain> ./cmd/server` (omit a package
   that was not touched).
3. Frontend iteration: from `web/`, run `bun run check`; run the nearest unit
   test when one exists.
4. Before reporting done: run `make check` exactly once on the completed
   candidate. If it fails, return to the smallest check that reproduces the
   failure; run `make check` again only after that check is green.

Do not use a passing targeted check as the done-signal. Do not repeatedly run
`make check` after edits unrelated to the reported failure.

## Stop conditions

Stop and escalate instead of widening the change when:

- the approved PRD is missing or its acceptance/out-of-scope boundary conflicts
  with the requested behavior;
- sensitive behavior lacks explicit owner approval for the current session;
- satisfying acceptance appears to require a new dependency, broad refactor, or
  destructive operation not approved by the PRD;
- an acceptance test reveals an unspecified business rule.

Done means every acceptance criterion is implemented, every PRD test-case ID is
traceable to an automated test, the PRD has moved from `backlog/` to
`developed/`, and the final `make check` passes.

## Done report

Keep the report compact:

```text
PRD: <developed path>
Acceptance: <criteria met>
Tests: <TC IDs and targeted commands>
Final gate: make check (pass)
Out of scope preserved: <items deliberately not changed>
Risks/follow-ups: <real residuals or none>
```
