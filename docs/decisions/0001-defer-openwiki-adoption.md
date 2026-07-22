---
type: ADR
title: Defer OpenWiki adoption; Git Markdown stays the source of truth
description: OpenWiki is not adopted as a build/runtime dependency now; it may be piloted later against measurable go/no-go criteria once the docs/ content contract is stable.
tags: [architecture, decisions, documentation]
status: accepted
owner: Founding Engineer
last_reviewed: 2026-07-22
---

# Defer OpenWiki adoption; Git Markdown stays the source of truth

## Context
The ALV-7 hardening plan for `template-v5` considered adopting
[OpenWiki](https://github.com/langchain-ai/openwiki) to make documentation
more discoverable for AI agents. A combined Product/Engineering audit
(2026-07-22) evaluated this against the plan's goals: fast agent retrieval,
no duplicated source of truth, and low operational overhead for a
still-small docs corpus.

## Decision
Do not adopt OpenWiki as a dependency now. `docs/` in Git remains the sole
source of truth for architecture, product context, PRDs, onboarding, and
runbooks. `docs/prds/README.md`'s front matter already follows a
source-grounded, navigable style compatible with future indexing, so this
decision does not foreclose a later pilot — it only refuses to take on an
indexing dependency before the content it would index is stable and large
enough to justify the operational cost.

## Alternatives considered
- **Adopt OpenWiki immediately.** Rejected: fewer than 20 canonical
  documents exist today, no cross-repo retrieval need has materialized, and
  a two-way sync or a second copy of any answer would violate the
  single-source-of-truth goal this same plan requires.
  [`docs/README.md`](README.md) already meets the plan's five-question,
  ≤2-minute retrieval acceptance without an external indexer.
- **Never revisit it.** Rejected: as the docs corpus and number of
  repositories grow, plain-text search may stop scaling for agent
  retrieval; closing the door permanently would be a decision this ADR
  isn't evidenced enough to make either.

## Consequences
- Documentation stays reviewable and versioned through the same PR path as
  code; no separate indexing pipeline to keep fresh or secure.
- Retrieval remains `rg`/grep-based; if that stops meeting the plan's
  ≤2-minute retrieval acceptance as the corpus grows, that is the trigger
  to revisit this decision, not a fixed calendar date.
- A future OpenWiki pilot is not blocked by this ADR. It requires, before
  starting: at least 20 canonical documents (or a demonstrated cross-repo
  retrieval need), a two-week least-privilege pilot, no two-way sync, and a
  measurable go/no-go scorecard (indexing coverage, source citation rate,
  freshness lag, secret/PII exclusion, and an operational owner). Recording
  that pilot's outcome — `adopt`, `iterate`, or `reject` — belongs in a new
  ADR that supersedes this one.

## Status
accepted — 2026-07-22
