---
name: new-feature
description: Start an application behavior change through an approved backlog PRD, one minimal vertical slice, targeted iteration, and one final make check.
---

This project is PRD-driven. Read, in order:

1. `AGENTS.md`;
2. `docs/prds/README.md`;
3. `docs/ai-delivery-path.md`;
4. the relevant approved PRD under `docs/prds/backlog/`.

These sources own the rules; this skill is only an entry point.

## Requirements first

Draft a new PRD at `docs/prds/backlog/<next-id>-<slug>.md` when no approved PRD
covers the requested behavior. Do not compress requirements discovery to save
execution time. Resolve ambiguities before implementation. Sensitive behavior
(auth, authorization, money, deletion, destructive migration) requires explicit
owner approval in the current session.

## Context limit

Follow the context budget and layer table in `docs/ai-delivery-path.md`. Read one
nearest sibling per required layer, search for symbols before opening more files,
and do not inventory the repository or read unrelated PRDs. Expand context only
to answer a concrete implementation or validation question.

## Execute one slice

Implement only the approved acceptance criteria in vertical-slice order. Run
`make gen` after proto or SQL changes; never edit generated files. Add automated
tests traceable to every `TC-<id>-n`. Preserve Out of Scope and stop if the work
would require an unapproved dependency, broad refactor, destructive action, or
new business rule.

## Validate without full-gate loops

Use the validation ladder in `docs/ai-delivery-path.md`: generator when needed,
then targeted backend/frontend checks while iterating. Run `make check` once on
the completed candidate as the final done-signal. If it fails, reproduce and fix
with the smallest targeted check before rerunning the final gate.

After acceptance and the final gate pass, move the same PRD from `backlog/` to
`developed/` and report with the compact done template in the delivery guide.
