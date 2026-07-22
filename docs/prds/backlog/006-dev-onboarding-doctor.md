---
type: Product requirement
title: Deterministic developer onboarding with make doctor
description: A non-mutating make doctor preflight check and a deterministic onboarding path so a new developer reaches a healthy local app without guesswork.
tags: [developer-experience, tooling, onboarding]
---

# Deterministic developer onboarding with make doctor

## Purpose
Let a developer diagnose a broken local environment (Go, Bun, PostgreSQL, config) with one read-only command and clear remediation, so a fresh clone reaches a healthy app through one documented, deterministic path.

## Acceptance:
- `make doctor` reports the status of the Go toolchain version, the Bun version, required `.env` configuration (`DATABASE_URL`, `JWT_SECRET` length, optional `MAIL_URL` format), and PostgreSQL reachability using the configured `DATABASE_URL`.
- `make doctor` never installs, upgrades, or modifies any tool, file, or database, and never creates/drops/resets PostgreSQL — it only reads local state and the network.
- Each failing check prints a specific, actionable remediation (an exact command, pinned version, or file to edit), not a generic error; passing checks are also printed for a full picture.
- `make doctor` exits non-zero if any check fails and zero when every check passes.
- `docs/onboarding/developer.md` documents the deterministic path (clone → `cp .env.example .env` → `make bootstrap` → `make doctor` → `make check` → `make run`) and `README.md` links to it as the onboarding entry point.

## Out of Scope
- No automatic installation/upgrade of Go, Bun, or PostgreSQL, and no `.env` file creation or mutation — remediation is printed, never applied.
- No changes to `make bootstrap`, `make check`, `make run`, or `make build` behavior.
- No new dependencies; PostgreSQL connectivity reuses the already-vendored `pgxpool`, and `MAIL_URL` parsing reuses the existing `internal/mail.ParseURL`.

## Test Cases
### TC-006-1: Go toolchain version mismatch is reported with remediation
- Given a reported Go version that does not match the required `1.26.x`
- When the Go version check runs
- Then it reports a failing status naming the found and required versions and remediation pointing to the Go download page

### TC-006-2: Go toolchain version match reports OK
- Given a reported Go version of `go1.26.5`
- When the Go version check runs against required `1.26`
- Then it reports an OK status with no remediation

### TC-006-3: Missing or mismatched Bun version is reported with remediation
- Given Bun is absent or its reported version does not equal the pinned `1.3.14`
- When the Bun version check runs
- Then it reports a failing status with the exact pinned version and an install command

### TC-006-4: Pinned Bun version reports OK
- Given a reported Bun version of `1.3.14`
- When the Bun version check runs against pinned `1.3.14`
- Then it reports an OK status with no remediation

### TC-006-5: Missing .env, short JWT_SECRET, and unset DATABASE_URL are each reported with remediation
- Given no `.env` file, a `JWT_SECRET` shorter than 32 bytes, and an empty `DATABASE_URL`
- When the config checks run
- Then each check reports a failing status with a distinct, actionable remediation

### TC-006-6: Valid config reports OK
- Given an existing `.env`, a `JWT_SECRET` of at least 32 bytes, and a non-empty `DATABASE_URL`
- When the config checks run
- Then each check reports an OK status

### TC-006-7: Malformed MAIL_URL is reported with remediation
- Given a non-empty `MAIL_URL` that fails to parse (missing host, port, or a non-`smtp` scheme)
- When the mail config check runs
- Then it reports a failing status with the parse error and remediation naming the expected `smtp://user:pass@host:port` shape

### TC-006-8: PostgreSQL connection failure is reported without mutating the database
- Given a ping attempt against `DATABASE_URL` returns an error (unreachable host or failed authentication)
- When the PostgreSQL check runs
- Then it reports a failing status including the underlying error and remediation to verify the running server and credentials, without attempting to create or reset anything

### TC-006-9: PostgreSQL connection success reports OK
- Given a ping attempt against `DATABASE_URL` returns no error
- When the PostgreSQL check runs
- Then it reports an OK status
