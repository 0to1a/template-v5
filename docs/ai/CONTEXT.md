# AI task context map

Read this after `AGENTS.md`. Load only the row needed for the task; do not scan the whole repository.

| Task | Read first | Typical validation |
|---|---|---|
| Backend/domain slice | nearest `internal/<domain>`, `proto/<domain>/v1`, `db/queries`, one `cmd/server/register_*.go` | `make check-go` |
| Database change | latest migration, nearest query file, `db/embed.go`, `sqlc.yaml` | `make gen && make check-go` |
| Connect contract | nearest proto, `buf.gen.yaml`, registration file | `make gen && make check-go` |
| Frontend route | nearest `web/src/routes` page, `web/src/lib/client.ts` | `make check-web` |
| Auth/security | `internal/auth`, allowlist in `cmd/server/main.go`, approved PRD | targeted Go tests, then `make check` |
| Docs-only | the target document and its index | inspect diff; no generated files |

## Vertical-slice order

For an approved PRD, take one narrow path through the system:

1. Translate each `TC-<id>-n` into a test checklist.
2. Add migration/query only when persistence is required.
3. Add the smallest Connect contract, then run `make gen` once.
4. Implement repository/service/handler by copying the nearest domain's shape.
5. Register the domain once and keep procedures protected by default.
6. Add the minimal Svelte route using the generated client.
7. Iterate with `make check-go` and/or `make check-web`; run `make check` once as the final gate.

## Stop conditions

Stop and request an owner decision when acceptance is ambiguous or the work changes auth, authorization, money, deletion, or a destructive migration without explicit approval. Never broaden the PRD while implementing it.
