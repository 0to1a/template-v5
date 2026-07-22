---
type: Product requirement
title: Benchmark remediation — Postgres provisioning gap and PRD-ID collision note
description: Close the two real documentation gaps the ALV-16 fresh-clone/agent-parity benchmark found, so future evaluators and agents don't repeat the same friction.
tags: [documentation, onboarding, prds, benchmark]
problem_brief: waiver
waiver_owner: Founding Engineer
waiver_reason: Bounded remediation directly evidenced by the ALV-16 benchmark run (two fresh-clone trials, two agent capability-parity trials) required by the approved ALV-7 hardening plan, Wave D. Internal template-repo documentation fix, not a customer-facing capability requiring product-market evidence.
waiver_expires: 2026-08-21
---

# Benchmark remediation — Postgres provisioning gap and PRD-ID collision note

## Purpose
Close the two concrete, repeatable friction points the ALV-16 benchmark measured — missing guidance for provisioning a fresh local PostgreSQL role/database, and no documented handling for concurrent-agent PRD-ID collisions — so the next fresh-clone evaluator or parallel agent doesn't hit either one blind.

## Acceptance:
- `docs/onboarding/developer.md` documents the exact commands to create a local PostgreSQL role and database from scratch (for a developer or agent with a bare PostgreSQL install and no existing role/db), consistent with `.env.example`'s `DATABASE_URL` shape.
- `docs/prds/README.md` states that two PRDs drafted concurrently on separate branches can legitimately claim the same next ID, that `make doc-lint` (`duplicate ID` check) is what catches it, and that the resolution is renumbering the PRD that merges second — not a merge-time surprise.
- `make doc-lint` reports zero issues against the resulting `docs/` tree.
- No application behavior changes; this PRD is documentation only.
- Both new doc claims above are proven by an automated test, not review alone.

## Out of Scope
- Automating PostgreSQL provisioning (e.g. via `make bootstrap` or a script) — `AGENTS.md` forbids implicitly creating the database, so provisioning stays a manual, documented step.
- A locking/reservation mechanism for PRD IDs — the accepted mitigation stays "doc-lint catches it, rename on merge," matching existing precedent (the ALV-14/ALV-15 wave collision on ID 010).
- Any change to `internal/platform/doclint` behavior.

## Test Cases
### TC-020-1: Fresh Postgres provisioning steps are present and consistent with `.env.example`
- Given `docs/onboarding/developer.md` and `.env.example`
- When `internal/platform/docs`'s automated test reads both files
- Then `developer.md` contains a `CREATE ROLE` and `CREATE DATABASE` command, and its example `DATABASE_URL` uses the same `postgres://` scheme as `.env.example`

### TC-020-2: PRD-ID collision handling is documented
- Given `docs/prds/README.md`
- When `internal/platform/docs`'s automated test reads it
- Then it contains the phrase "duplicate PRD ID" and explicitly says the resolution is renumbering the PRD that merges second
