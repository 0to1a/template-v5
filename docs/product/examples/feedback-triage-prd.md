---
type: Product requirement example
canonical: false
title: Weekly source-linked feedback digest
description: Synthetic example showing how a PRD links to a proceed problem brief.
tags: [feedback, digest, example]
problem_brief: ./feedback-triage-problem-brief.md
---

# Weekly source-linked feedback digest

> **Synthetic, non-canonical example.** Do not implement this file. A real capability belongs in `docs/prds/backlog/` and requires real evidence.

## Purpose
Enable a B2B SaaS product lead in the concierge experiment to review recurring customer problems with source context in under 30 minutes.

## Acceptance:
- Given redacted feedback entries supplied for one week, when a digest is prepared, then each reported problem includes its frequency and links to every supporting source entry.
- Given a product lead reviews a digest, when they accept, edit, or reject a reported problem, then that disposition is recorded for experiment measurement.
- Given no feedback entries are supplied for a week, when the digest is prepared, then it explicitly reports an empty period rather than inventing a problem.

## Out of Scope
- Automated ingestion or AI classification.
- A production UI, integrations, or notifications.
- Storage of unredacted customer personal data.

## Test Cases
### TC-EXAMPLE-1: Reported problems retain source traceability
- Given a redacted weekly input set
- When the digest is prepared
- Then every reported problem links to all supporting input entries and states its frequency

### TC-EXAMPLE-2: Reviewer disposition is measurable
- Given a prepared digest
- When the product lead accepts, edits, or rejects an item
- Then the chosen disposition is recorded

### TC-EXAMPLE-3: Empty input does not create claims
- Given no entries for the week
- When the digest is prepared
- Then the digest reports an empty period and contains no inferred customer problem
