# Working Rules

- When repository guidance conflicts, precedence is: this file (`AGENTS.md`) → domain documentation (e.g. `docs/prds/README.md`, `docs/onboarding/`) → the approved PRD for the specific change being implemented. A domain doc or PRD may add detail but never override a rule stated here.
- Every behavior change starts with a small PRD under `docs/prds/backlog/` (see `docs/prds/README.md`).
- PRD drafting is a separate, unhurried step from implementation — speed optimizations apply to the coding phase only, never to shortcut requirements gathering or approval.
- Implement only its Acceptance criteria and respect Out of Scope.
- Sensitive behavior (auth, authorization, money, deletion, destructive migration) requires owner approval before implementation.
- Handwritten backend code lives under `internal/<domain>`; contracts under `proto/<domain>/v1`.
- Register each domain once through `cmd/server/register_<domain>.go`; never add a second registry.
- All Connect procedures are protected by default; public procedures require the explicit allowlist in `cmd/server/main.go`.
- SQL lives under `db/queries`; schema history lives under `db/migrations` (embedded, applied at server startup, up only).
- Never edit `internal/gen` or `web/src/lib/gen` manually; regenerate with `make gen`.
- Frontend requests use the generated Connect client through `web/src/lib/client.ts`; the token only moves through `web/src/lib/auth.ts`.
- Keep Svelte components HTML-first and use Svelte 5 runes.
- Never log OTPs, JWTs, database URLs, or secrets.
- Never create, drop, or reset PostgreSQL implicitly.
- Do not add dependencies or broad refactors unless the PRD requires and approves them.
- Add automated tests traceable to PRD test-case IDs (`TC-<id>-n`).
- While iterating on code, prefer targeted `go build`/`go test` on the touched packages; run the full `make check` once, as the final gate, before declaring work complete.
