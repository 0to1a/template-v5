# Working Rules

- Every behavior change starts with a small PRD under `docs/prds/` (see `docs/prds/README.md`).
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
- Run `make check` before declaring work complete.
