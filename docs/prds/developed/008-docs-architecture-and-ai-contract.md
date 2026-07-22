---
type: Product requirement
title: Documentation Architecture and AI Contract
description: A canonical docs index, architecture/repository map, ADR and runbook templates, a worked vertical-slice example, a completion-report template, and a basic doc-lint check.
tags: [documentation, developer-experience, tooling]
problem_brief: waiver
waiver_owner: Founding Engineer
waiver_reason: Internal documentation and AI-contract infrastructure for template-v5 itself, directed by the approved ALV-7 hardening plan (Fase 2 / Wave B); not a customer-facing capability requiring product-market evidence.
waiver_expires: 2026-08-22
---

# Documentation architecture and AI contract

## Purpose
Give an AI agent or new contributor a single, navigable `docs/` index and a small doc-lint check so architecture, ownership, and rationale are found in minutes instead of a broad repository scan.

## Acceptance:
- `docs/README.md` exists as the documentation index, states the canonical `docs/` structure (`product/`, `decisions/`, `prds/`, `onboarding/`, `runbooks/`), and directly answers "who is the ICP", "what is activation", "why does this requirement exist", "how do I run this", and "who owns this" by name and link, each traceable to one authoritative document.
- `docs/architecture.md` documents the request lifecycle browser → Connect RPC → Go service → repository → PostgreSQL naming the real files/packages involved (`cmd/server`, `internal/<domain>`, `internal/gen`, `db/migrations`), and `docs/repository-map.md` lists every top-level directory with a one-line purpose and a pointer to its owning doc.
- `docs/decisions/` exists with an ADR template plus at least one real accepted ADR, `docs/runbooks/` has a runbook template, one worked vertical-slice example traces a real developed PRD across proto/backend/frontend/migration/tests, and a completion-report template exists and is linked from `docs/prds/README.md`'s Report step.
- `make doc-lint` is a new, non-mutating target that reports every markdown file under `docs/` missing a required front-matter field, every internal link that does not resolve to an existing file, and any duplicate PRD ID across `backlog/` and `developed/`; it exits non-zero if any issue is found and zero otherwise.
- Running `make doc-lint` against the `docs/` tree produced by this change reports zero issues.

## Out of Scope
- Wiring `doc-lint` into CI or into `make check` (Fase 4 / quality-automation concern).
- Adopting OpenWiki or any external documentation indexing tool — recorded here only as an ADR decision to defer (Fase 5 prerequisite not yet met).
- Rewriting or renumbering existing PRDs, runbooks, or onboarding docs beyond adding missing front matter or links.
- A full YAML parser or a new dependency; front-matter and link scanning is a minimal hand-rolled scanner sufficient for this repo's fixed format.

## Test Cases
### TC-008-1: Missing front-matter field is reported
- Given a markdown file under `docs/` with front matter missing the `description` key
- When `doc-lint` runs
- Then it reports that file and the missing field, and exits non-zero

### TC-008-2: Broken internal link is reported
- Given a markdown file under `docs/` with a relative link to a file that does not exist
- When `doc-lint` runs
- Then it reports that file and the unresolved link path, and exits non-zero

### TC-008-3: Duplicate PRD ID is reported
- Given two files under `docs/prds/backlog/` and `docs/prds/developed/` that share the same three-digit ID prefix
- When `doc-lint` runs
- Then it reports the duplicate ID and both file paths, and exits non-zero

### TC-008-4: A clean docs tree passes
- Given every markdown file under `docs/` has required front matter, every internal link resolves, and no PRD ID is duplicated
- When `doc-lint` runs
- Then it reports no issues and exits zero
