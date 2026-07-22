---
type: Product requirement
title: Fase 5 OpenWiki pilot gate check and go/no-go plan
description: Confirm ADR 0001's ≥20-canonical-document trigger is met, and define the least-privilege, two-week, no-two-way-sync OpenWiki pilot and its measurable go/no-go scorecard, without adopting OpenWiki as a dependency or provisioning it.
tags: [documentation, architecture, governance]
problem_brief: waiver
waiver_owner: Founding Engineer
waiver_reason: Internal docs-tooling evaluation for template-v5 itself, directed by the approved ALV-7 hardening plan (Fase 5); not a customer-facing capability requiring product-market evidence.
waiver_expires: 2026-08-22
---

# Fase 5 OpenWiki pilot gate check and go/no-go plan

## Purpose
Give the owner a reproducible confirmation that [`docs/decisions/0001-defer-openwiki-adoption.md`](../../decisions/0001-defer-openwiki-adoption.md)'s pilot trigger is now met, and a written, approval-ready pilot plan with a numeric go/no-go scorecard, so no OpenWiki infrastructure is provisioned before the owner explicitly approves the vendor/cost commitment.

## Acceptance:
- This PRD records a reproducible count of canonical documents under `docs/` (excluding templates, index/guide files, and worked examples) and states whether it meets ADR 0001's "at least 20 canonical documents" trigger.
- A written pilot plan specifies: a 2-week timebox; read-only, least-privilege access scoped to `docs/` only (no write-back, no credentials beyond what indexing requires); that Git Markdown remains the sole source of truth with no two-way sync; and that OpenWiki must not become a runtime or build dependency of `template-v5` at any point during or after the pilot.
- The plan defines a numeric or binary pass/fail threshold for each go/no-go scorecard dimension: source-citation rate, freshness lag, retrieval success/latency, PII/secret exclusion, a named operational owner, and hosting/operating cost — matching ADR 0001's consequences section.
- The plan states explicitly that manual duplication of any answer (a second copy of content OpenWiki already indexes) is an automatic no-go, matching this issue's stated no-go condition.
- The plan states explicitly that this PRD does not authorize provisioning any OpenWiki instance, credentials, or hosting — that is a vendor/cost commitment requiring owner approval first, obtained outside this PRD.

## Out of Scope
- Provisioning, deploying, or granting any credentials to an OpenWiki instance.
- Running the live 2-week pilot or collecting real usage metrics.
- Writing the pilot's outcome ADR (`adopt` / `iterate` / `reject`) — that only happens after the pilot actually runs, and supersedes ADR 0001.

## Test Cases
### TC-021-1: Canonical-document gate is met and reproducible
- Given `docs/` excluding files named `TEMPLATE.md`, `README.md`, matching `*-template.md`, under `docs/product/examples/`, or `vertical-slice-example.md`
- When counted with `find docs -iname '*.md' ! -iname 'TEMPLATE.md' ! -iname 'README.md' ! -iname '*-template.md' ! -path 'docs/product/examples/*' ! -iname 'vertical-slice-example.md' | wc -l`
- Then the count is 33, which is ≥20, so ADR 0001's trigger is met

### TC-021-2: Pilot plan defines a numeric go/no-go scorecard
- Given `docs/prds/backlog/021-fase-5-openwiki-pilot-gate-and-plan.md`'s Pilot plan section
- When each scorecard dimension is checked
- Then citation rate, freshness lag, retrieval, PII/secret exclusion, owner, and cost each have a stated pass/fail threshold

### TC-021-3: Plan forbids two-way sync, runtime/build dependency, and unapproved provisioning
- Given the same Pilot plan section
- When reviewed
- Then it states read-only/no-two-way-sync, states OpenWiki is not a runtime/build dependency, and states provisioning requires owner approval not granted by this PRD

## Gate check
Canonical-document count (method reproducible via the `find` command in TC-021-1): **33** documents, comfortably above ADR 0001's 20-document trigger. No cross-repo retrieval need has separately materialized (template-v5 remains the only repository in scope). The document-count trigger alone is sufficient to proceed to a pilot proposal under ADR 0001 — but proceeding to *provisioning* still requires owner approval per this PRD's Acceptance and the Founding Engineer's standing boundary against unapproved vendor/cost commitments.

## Pilot plan (proposed; not yet approved to execute)
- **Timebox:** 2 weeks from the day the owner approves provisioning.
- **Scope of access:** read-only indexing of `docs/**/*.md` only — no access to `internal/`, `db/`, secrets, or any other repository content. No write-back path is configured; Git Markdown stays the sole source of truth for every answer (no two-way sync).
- **Dependency boundary:** OpenWiki is not added to `go.mod`, `web/package.json`, `Makefile` build/check targets, CI, or any runtime path. It is an external, optional indexer that can be deleted at any time with zero impact on building, testing, or running `template-v5`.
- **Go/no-go scorecard** (measured over the 2-week window):
  | Dimension | Pass threshold |
  |---|---|
  | Source citation | ≥90% of pilot answers cite the exact `docs/` file(s) they draw from |
  | Freshness | Index lag behind the latest merged `docs/` change ≤24 hours for ≥95% of the window |
  | Retrieval | ≥90% of a fixed 10-question retrieval test set (drawn from `docs/README.md`'s five-question retrieval plus 5 harder cross-doc questions) answered correctly within the plan's ≤2-minute acceptance |
  | PII/secret exclusion | Zero instances of a secret, credential, or personal data value surfaced in any indexed answer (checked against the same asset list in `docs/threat-model.md`) |
  | Owner | A named operational owner (person, not "the team") is assigned and confirmed in writing before the pilot starts |
  | Cost | Actual 2-week hosting/operating cost stays within a dollar ceiling the owner sets at approval time |
- **No-go conditions (any one triggers `reject` or `iterate`, not `adopt`):**
  - Any manual duplication of content is found (a human or agent copies an answer instead of relying on the index) — automatic no-go per this issue.
  - Freshness, citation, or owner thresholds are not met.
- **Outcome:** within one week of the pilot ending, a new ADR (`docs/decisions/000X-openwiki-pilot-outcome.md`) records `adopt`, `iterate`, or `reject` against the scorecard above and supersedes ADR 0001.
- **Approval gate:** provisioning any OpenWiki instance, credentials, or hosting spend does not happen until the owner explicitly approves this plan (see the confirmation request on issue ALV-17).
