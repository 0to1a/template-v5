---
type: Architecture
title: Request lifecycle and system architecture
description: How a request travels from the browser through Connect RPC, the Go service layer, and the repository layer to PostgreSQL, and how the app ships as one binary.
tags: [architecture, backend, frontend]
status: active
owner: Founding Engineer
last_reviewed: 2026-07-22
---

# Request lifecycle and system architecture

Single-binary Go + Svelte, protected-by-default RPC, migrations embedded and
applied up-only at startup. See [`repository-map.md`](repository-map.md) for
what lives in each directory named below.

## Request lifecycle

```text
Browser (SvelteKit SPA, web/src/routes/**)
  -> generated Connect-ES client (web/src/lib/gen, never hand-edited)
  -> web/src/lib/client.ts constructs the call; web/src/lib/auth.ts is the
     only place the bearer token is read/attached
  -> HTTP -> Go server (cmd/server/main.go, the composition root)
  -> auth interceptor (internal/auth/interceptor.go): every Connect
     procedure is protected by default; only procedures explicitly listed
     in main.go's publicProcedures allowlist skip it
  -> domain handler (internal/<domain>/handler.go) implements the
     generated Connect service interface
  -> domain service (internal/<domain>/service.go) holds business logic
  -> repository (internal/<domain>/repository.go) wraps sqlc-generated
     queries (internal/gen/db) built from db/queries/*.sql
  -> PostgreSQL
```

`GET /health` (`internal/health`) is the one unauthenticated route and
performs no database query — it reflects process liveness, not database
health.

## Domain registration

Each domain registers itself exactly once:

1. Handwritten logic under `internal/<domain>` (handler, service,
   repository, and anything domain-specific like `internal/auth/otp.go`).
2. A single `cmd/server/register_<domain>.go` that mounts the domain's
   generated Connect handler on the shared `http.ServeMux`.
3. One call to that `register_<domain>` function from `cmd/server/main.go`.

There is never a second registry — `main.go` is the only place handlers are
wired together, and it stays a thin composition root: load config, connect
the database, apply migrations, construct dependencies, register handlers,
listen.

## Data layer

- `db/queries/*.sql` is handwritten SQL; `sqlc` (invoked by `make gen`)
  generates the typed Go repository code in `internal/gen/db`.
- `db/migrations/*.sql` is schema history, embedded into the binary via
  `db/embed.go`, and applied automatically at server startup — up only.
  The server never creates, drops, or resets the database itself; a
  database and role must already exist (see
  [`onboarding/developer.md`](onboarding/developer.md)).

## Contracts

- `proto/<domain>/v1/*.proto` is the Connect/protobuf contract, the single
  source `make gen` uses to produce both the Go server stubs
  (`internal/gen/<domain>`) and the TypeScript client stubs
  (`web/src/lib/gen`). Neither generated tree is hand-edited.

## Deployment shape

`make build` runs the frontend production build (`cd web && bun run
build`) and embeds its output via `cmd/server/register_frontend.go`, then
compiles a single `bin/server` binary that serves the API and the SPA from
one process. There is no separate frontend server in production.

## Security posture baked into this lifecycle

- Protected-by-default Connect procedures — see the interceptor and
  allowlist above.
- The bearer token only moves through `web/src/lib/auth.ts`; nothing else
  in the frontend reads or writes it directly.
- Migrations are additive/up-only and embedded, so schema state is always
  reproducible from the binary that is running.

See [`README.md`](README.md) for how this fits into the wider documentation
set, and [`vertical-slice-example.md`](vertical-slice-example.md) for one
real PRD traced through every layer described here.
