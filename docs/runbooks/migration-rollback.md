---
type: Runbook
title: "Migration rollback"
description: "How to recover from a bad schema migration without ever running goose down or resetting the database, matching this template's up-only migration policy."
tags: [runbook, database, migration]
status: active
owner: Founding Engineer
last_reviewed: 2026-07-22
---

# Migration rollback

## Purpose
Recover from a migration that shipped a bug, without violating this
repository's up-only migration policy (`AGENTS.md`: "Never create, drop, or
reset PostgreSQL implicitly"; `db/migrations/*.sql` is "schema history...
applied up-only"). There is no `goose down` step in this template's
operational path — rollback here means "write a new corrective migration",
not "undo the last one".

## Preconditions
- Access to the production database connection the server itself uses.
- The bad migration has already been identified (its filename under
  `db/migrations/`).
- Whoever runs this has authority to modify production schema — treat this
  the same as any other destructive-migration-adjacent action requiring
  care (`AGENTS.md`: "Sensitive behavior... destructive migration").

## Steps
1. Confirm which migrations have actually applied on the target database:
   ```sh
   psql "$DATABASE_URL" -c "select id, is_applied, tstamp from goose_db_version_metadata order by id;"
   ```
   (Table name may be goose's default `goose_db_version` on older schemas —
   check with `\dt` if the above errors.)
2. Do not hand-edit or delete the bad migration file. It already ran
   against a real database; rewriting history makes the file and the
   database disagree the next time the binary starts.
3. Write a new migration under `db/migrations/` that corrects the damage
   forward — e.g. drop the wrongly-added column, backfill a value, or
   recreate an index with the right definition. Follow the same `-- +goose
   Up` / `-- +goose Down` shape as `00001_users.sql`.
4. Test the new migration against a disposable local database first:
   ```sh
   make run
   ```
   (the server applies all pending migrations, including the new one, at
   startup — a failure here aborts startup loudly instead of leaving the
   schema half-migrated).
5. Ship the corrective migration through the normal PRD → PR → CI path like
   any other change — it is still schema history, not a special case.
6. Deploy. The corrective migration applies automatically at the next
   server startup, same as any other migration.

## Verification
- `make run` (or the deployed server's own startup log) shows the new
  migration applying with no error.
- The specific bug the corrective migration targeted is confirmed fixed
  against a real query, not just "the migration ran".

## Rollback
This procedure's own "rollback" is itself: if the corrective migration is
also wrong, write another corrective migration. There is no destructive
undo path by design — schema state must always be reproducible forward from
the binary that is running (`docs/architecture.md`, "Security posture").

## Owner / escalation
Owned by whoever holds production database access for the deployment in
question. Automation identities do not hold standing production database
credentials by default — see `AGENTS.md`, "Never become the sole holder of
production access". Escalate to the founder immediately if a migration has
already caused data loss; do not attempt a destructive fix under time
pressure.

## Test case trace
Not implemented in code — this runbook is a documented human procedure
(see `docs/prds/developed/013-fase-3-operational-documentation.md`'s
Acceptance for what is verified: exact commands, a verification step, and
no destructive rollback path).
