---
type: Developer onboarding
title: Deterministic developer onboarding
description: The one path from a fresh clone to a healthy local app, verified by a non-mutating make doctor check.
tags: [developer-experience, onboarding, tooling]
status: active
owner: Engineering
last_reviewed: 2026-07-22
---

# Deterministic developer onboarding

This is the one supported path from a fresh clone to a healthy, running app.
Follow it in order; don't skip `make doctor` when something fails, since it
tells you exactly what to fix and how.

## 1. Prerequisites

- Go, matching the version pinned in [`go.mod`](../../go.mod) (currently 1.26.x). Install from https://go.dev/dl/.
- Bun, exact version `1.3.14`. Install with:
  ```bash
  curl -fsSL https://bun.sh/install | bash -s "bun-v1.3.14"
  ```
- An external PostgreSQL server you can already authenticate to. This project never creates, drops, or resets the database server itself (see [`AGENTS.md`](../../AGENTS.md)) — a database and role must already exist.

## 2. The deterministic path

```bash
git clone <this repo> && cd <this repo>
cp .env.example .env     # then edit: DATABASE_URL to a real role/database, JWT_SECRET to >=32 random bytes
make bootstrap            # installs Go modules and web dependencies — the only step that installs anything
make doctor                # read-only preflight: Go, Bun, PostgreSQL, config — fix anything it reports before continuing
make check                 # the done-signal: codegen, lint, tests, both builds
make run                   # build the frontend, then run the single server process
```

`make help` lists every available command.

## 3. What `make doctor` checks (and doesn't do)

`make doctor` reports the status of:

- the Go toolchain version,
- the Bun version,
- whether `.env` exists,
- `DATABASE_URL` and `JWT_SECRET` (set, and long enough),
- `MAIL_URL` (optional; only validated if set),
- whether PostgreSQL is reachable with the configured `DATABASE_URL`.

For every failing check it prints a specific remediation: an exact install
command, the exact pinned version, or the file and field to edit. It never
installs or upgrades a tool, never writes `.env`, and never creates, drops,
resets, or otherwise mutates the database — it only reads local files and
opens a connection to verify reachability. It exits non-zero if anything
fails, so it composes with CI or a pre-flight script.

## 4. Known issues found during the first dry run

Running this path on a genuinely fresh machine surfaced three real failures,
which is exactly what `make doctor` now catches by name instead of an
unlabeled `make bootstrap`/`make check` failure partway through:

- **Bun not installed.** `make bootstrap` fails opaquely at `_check-tools`
  with no install instructions; `make doctor` names the pinned version and
  the exact install command.
- **No `.env` file.** The server and `make doctor` both refuse to guess
  configuration; `make doctor` names the exact `cp` command to fix it.
- **`.env.example`'s `DATABASE_URL` is a placeholder, not a working
  default.** Its `user`/`password`/`template_v5` values will not match a
  real PostgreSQL role or database on most machines, and PostgreSQL being
  merely *reachable* over TCP does not mean these credentials are valid.
  `make doctor` performs a real connection attempt (not just a port check)
  and reports the underlying authentication or database error verbatim.

## 5. Documentation precedence

When repository guidance conflicts, the order is: [`AGENTS.md`](../../AGENTS.md)
→ domain documentation (e.g. [`docs/prds/README.md`](../prds/README.md)) →
the approved PRD for the specific change being implemented. See `AGENTS.md`
for the full statement.

## 6. Where to go next

- [`docs/prds/README.md`](../prds/README.md) — the PRD workflow every behavior change follows.
- [`AGENTS.md`](../../AGENTS.md) — working rules (generated-file boundaries, testing, security).

This document is about developer setup, not end-user (customer-facing)
product onboarding.
