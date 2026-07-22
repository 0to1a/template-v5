---
type: Repository map
title: Repository map
description: Every top-level directory and file, one line each, with a pointer to its owning document.
tags: [architecture, repository-map]
status: active
owner: Founding Engineer
last_reviewed: 2026-07-22
---

# Repository map

One line per top-level entry. See [`architecture.md`](architecture.md) for
how these pieces connect at request time.

## Root

- [`AGENTS.md`](../AGENTS.md) (`CLAUDE.md` is a symlink to it) — working rules; highest precedence in this repo.
- [`README.md`](../README.md) — entry point; links to onboarding.
- `Makefile` — `make help`, `bootstrap`, `doctor`, `doc-lint`, `gen`, `check`, `run`, `build`.
- `go.mod` / `go.sum` — pinned Go toolchain version and module dependencies.
- `buf.yaml` / `buf.gen.yaml` — protobuf lint/generation config for `make gen`.
- `sqlc.yaml` — SQL-to-Go codegen config for `make gen`.
- `.env.example` — template for local `.env` (`DATABASE_URL`, `JWT_SECRET`, optional `MAIL_URL`).
- `Dockerfile` / `.dockerignore` — reproducible, vendor-neutral container image; see [`runbooks/container-deployment.md`](runbooks/container-deployment.md).
- `.github/workflows/` — CI; runs the equivalent of `make check` on pull requests.
- `.claude/skills/` — repository-specific agent skills (e.g. `new-feature`, the PRD-driven implementation workflow).

## `cmd/` — composition roots (thin, impure shells)

- `cmd/server` — the production server: `main.go` wires config, database, migrations, domain handlers, and the HTTP mux; `register_<domain>.go` mounts one domain's Connect handler.
- `cmd/doctor` — thin CLI shell behind `make doctor`; classification logic lives in `internal/platform/doctor`.
- `cmd/doclint` — thin CLI shell behind `make doc-lint`; classification logic lives in `internal/platform/doclint`.

## `internal/` — handwritten backend code

- `internal/<domain>` (currently `auth`, `health`, `mail`) — handler/service/repository for one domain; never a second registry outside `cmd/server/register_<domain>.go`.
- `internal/gen` — generated proto stubs and sqlc repository code. Never edit by hand; regenerate with `make gen`.
- `internal/platform/config` — environment/config loading and startup validation.
- `internal/platform/database` — connection pooling and migration runner.
- `internal/platform/server` — SPA static-file handler and the `Run` graceful-shutdown helper used by `cmd/server/main.go`.
- `internal/platform/observability` — vendor-neutral `Logger` interface, request-correlation-ID middleware, structured access logs.
- `internal/platform/doctor` — pure classification logic behind `make doctor`.
- `internal/platform/doclint` — pure classification logic behind `make doc-lint`.

## `proto/` — Connect/protobuf contracts

- `proto/<domain>/v1/*.proto` — source of truth for one domain's RPC contract; `make gen` produces both Go and TypeScript stubs from it.

## `db/` — schema and queries

- `db/migrations/*.sql` — schema history, embedded in the binary, applied up-only at server startup.
- `db/queries/*.sql` — handwritten SQL; source for the sqlc-generated repository code in `internal/gen/db`.
- `db/embed.go` — embeds the migrations directory into the binary.

## `web/` — Svelte 5 / SvelteKit frontend

- `web/src/routes` — SPA pages.
- `web/src/lib/gen` — generated Connect-ES client stubs. Never edit by hand; regenerate with `make gen`.
- `web/src/lib/client.ts` — the one place Connect requests are constructed.
- `web/src/lib/auth.ts` — the one place the bearer token is read or written.

## `docs/` — documentation

See [`README.md`](README.md) for the full index; the canonical subtrees are
`product/`, `decisions/`, `prds/`, `onboarding/`, and `runbooks/`.
