---
type: Documentation index
title: Documentation index
description: The canonical docs/ structure and the single-owner answer to the five most common retrieval questions.
tags: [documentation, index, architecture]
status: active
owner: Founding Engineer
last_reviewed: 2026-07-22
---

# Documentation index

This is the entry point into `docs/`. Git Markdown is the source of truth —
no external wiki or indexer holds a second copy of any answer below (see
[`decisions/0001-defer-openwiki-adoption.md`](decisions/0001-defer-openwiki-adoption.md)).
Each topic has exactly one authoritative document; if you find the same
claim in two places, that is a bug — fix it by removing the duplicate, not
by adding a third source.

## Canonical structure

```text
docs/
  product/       # thesis, problem brief, positioning, metrics
  decisions/     # ADRs — cross-cutting technical/process decisions
  prds/          # backlog/developed capability contracts (the unit of work)
  onboarding/    # developer and per-product onboarding contracts
  runbooks/      # operations, release, incident procedures
  architecture.md        # request lifecycle
  repository-map.md      # what lives where, one line each
  vertical-slice-example.md  # one real PRD traced through every layer
  report-template.md     # completion-report contract
```

## Five-question retrieval

| Question | Answer lives here |
|---|---|
| Who is the ICP (target user)? | Each venture's problem brief, following the chain in [`product/README.md`](product/README.md). This repository is the shared template, not a venture itself — it has no single ICP. |
| What is activation (first value)? | [`onboarding/product.md`](onboarding/product.md) — every product built from this template must define its activation event and time-to-value there. |
| Why does this requirement exist? | The specific PRD's `problem_brief` front-matter link (see [`prds/README.md`](prds/README.md)), or a cross-cutting [`decisions/`](decisions/) ADR for architecture/process choices. |
| How do I run this? | [`onboarding/developer.md`](onboarding/developer.md) — `make bootstrap` → `make doctor` → `make check` → `make run`. |
| Who owns this? | The `owner` field in the front matter of the specific canonical doc, PRD, or runbook; repository-wide precedence and defaults are in [`AGENTS.md`](../AGENTS.md). |

## Where to go next

- [`architecture.md`](architecture.md) — request lifecycle from browser to PostgreSQL.
- [`repository-map.md`](repository-map.md) — every top-level directory, one line each.
- [`product/README.md`](product/README.md) — problem brief → PRD context gate.
- [`prds/README.md`](prds/README.md) — the PRD workflow every behavior change follows.
- [`onboarding/developer.md`](onboarding/developer.md) — deterministic local setup.
- [`onboarding/product.md`](onboarding/product.md) — per-product onboarding contract.
- [`runbooks/`](runbooks/) — operational procedures; start from [`runbooks/TEMPLATE.md`](runbooks/TEMPLATE.md).
- [`decisions/`](decisions/) — ADRs; start from [`decisions/README.md`](decisions/README.md).
- [`vertical-slice-example.md`](vertical-slice-example.md) — one real PRD traced end to end.
- [`report-template.md`](report-template.md) — the completion-report contract for every finished PRD.

## Precedence

`AGENTS.md` → domain documentation (this tree) → the approved PRD for the
specific change. A domain doc may add detail but never override a rule
stated in `AGENTS.md`.

## Basic doc lint

Run `make doc-lint` to check every file under `docs/` for required front
matter, resolvable internal links, and unique PRD IDs across
`prds/backlog/` and `prds/developed/`. It is read-only and reports every
problem it finds; it is not yet wired into `make check` or CI (tracked for
the quality-automation phase).
