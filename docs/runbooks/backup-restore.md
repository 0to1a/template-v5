---
type: Runbook
title: "Backup and restore"
description: "How to take and verify a PostgreSQL backup, and how to restore it, without the server ever creating, dropping, or resetting the database itself."
tags: [runbook, database, backup, disaster-recovery]
status: active
owner: Founding Engineer
last_reviewed: 2026-07-22
---

# Backup and restore

## Purpose
Recover from data loss or corruption using a database-level backup. The
application itself takes no backups and never touches the database's
existence (`AGENTS.md`: "Never create, drop, or reset PostgreSQL
implicitly") — backup and restore are always an operator action against
PostgreSQL directly, independent of the server binary.

## Preconditions
- `pg_dump`/`pg_restore` (or the managed database provider's equivalent —
  e.g. Railway's PostgreSQL plugin ships its own backup tooling) available
  to whoever runs this.
- Access to `DATABASE_URL` for the target environment. Never print this
  value in a shared log — see `docs/environment-contract.md`.
- Restoring to production requires the same approval posture as any other
  destructive-migration-adjacent action (`AGENTS.md`).

## Steps

### Backup
1. Take a full logical backup:
   ```sh
   pg_dump --format=custom --file="backup-$(date +%Y%m%d-%H%M%S).dump" "$DATABASE_URL"
   ```
2. Store the resulting `.dump` file somewhere durable and access-controlled
   (not committed to the repository, not a public bucket).
3. Record the backup's timestamp and the migration version it was taken at
   (see `migration-rollback.md` step 1 for how to read the applied
   migration version) so a later restore knows exactly what schema it will
   bring back.

### Restore
1. Restore into a **new, empty** database first — never restore directly
   over a live database you might still need:
   ```sh
   createdb restore_check
   pg_restore --dbname=postgres://user:password@localhost:5432/restore_check backup-<timestamp>.dump
   ```
2. Verify the restored data (row counts, a few known records) before doing
   anything with it.
3. Only after verification, and only with explicit approval for the target
   environment, point `DATABASE_URL` at the restored database (or restore
   into the real target if the provider's tooling requires that instead of
   a side-by-side check — follow the provider's own documented restore
   path in that case).
4. Start the server against the restored database. Migrations are up-only
   and idempotent against an already-migrated schema, so this is safe to do
   even if the backup predates the latest migration — the server brings it
   forward automatically.

## Verification
- Row counts and a handful of known records in the restored database match
  what was expected at backup time.
- The server starts cleanly against the restored database (`make run` or
  the deployed process), migrations apply without error, and `GET
  /health/ready` returns 200.

## Rollback
If a restore turns out to be wrong (wrong backup, wrong target), do not
attempt to "undo" it in place — restore again from a known-good backup into
a fresh database, following the same Steps above.

## Owner / escalation
Owned by whoever holds production database access. Automation identities do
not hold standing production database credentials or backup credentials by
default (`AGENTS.md`, "Never become the sole holder of production access").
Escalate to the founder immediately for any suspected data loss before
attempting a restore under time pressure.

## Test case trace
Not implemented in code — this runbook is a documented human procedure. See
`docs/prds/developed/013-fase-3-operational-documentation.md`'s Acceptance.
