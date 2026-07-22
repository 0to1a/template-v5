---
type: Product guide
title: PRD workflow
description: The PRD lifecycle and format every behavior change follows, from backlog draft through implementation to developed.
tags: [prds, workflow, governance]
status: active
owner: Founding Engineer
last_reviewed: 2026-07-22
---

# PRDs — the unit of work

Every behavior change starts as a backlog PRD:

```text
docs/prds/backlog/<id>-<slug>.md      e.g. docs/prds/backlog/007-logout.md
```

After the acceptance criteria are implemented and validated, move the same file to:

```text
docs/prds/developed/<id>-<slug>.md
```

- ID: three digits, monotonic across **both** lifecycle folders, never reused. The next ID is the highest existing one + 1.
- Slug: lowercase kebab-case.
- One PRD = one verifiable capability. Split big PRDs before implementing.
- A PRD is not a design document, implementation plan, or file list.
- Every new PRD must include a `problem_brief` front-matter link to a brief whose `status` is `proceed`. If evidence cannot be completed before genuinely urgent work, use `problem_brief: waiver` plus `waiver_owner`, `waiver_reason`, and `waiver_expires`; the waiver does not bypass sensitive-behavior approval. See [`docs/product/README.md`](../product/README.md).
- Test-case IDs reuse the PRD ID: `TC-007-1`, `TC-007-2`, …
- Folder location is the lifecycle signal: `backlog/` means not yet delivered; `developed/` means implemented and validated.

## Required format

```markdown
---
type: Product requirement
title: Short descriptive title
description: One-sentence summary of the capability and user value.
tags: [domain, capability]
problem_brief: ../../product/<brief>.md
---

# Title

## Purpose
[1 sentence — what and for whom]

## Acceptance:
- [2–5 verifiable criteria]
-

## Out of Scope
- [1–3 things that must not be done — broad refactoring, adding new dependencies, etc.]
-

## Test Cases
### TC-<xxx>-1: [expected behavior]
- Given
- When
- Then

### TC-<xxx>-n: [expected behavior]
- Given
- When
- Then
```

The front matter follows the source-grounded, navigable style used by [OpenWiki](https://github.com/langchain-ai/openwiki). Keep it short: it improves discovery, but does not replace the behavioral sections below it.

## Quality gate — a PRD is ready when

- front matter has a concrete title, a one-sentence description, useful domain/capability tags, and either a linked `proceed` problem brief or a complete, unexpired waiver;
- Purpose is exactly one sentence naming the behavior and its user;
- Acceptance has 2–5 criteria, each objectively verifiable;
- Out of Scope has 1–3 entries;
- every important acceptance criterion has at least one Test Case;
- Given/When/Then describe behavior, not internal function details;
- nothing subjective ("nice UI", "clean code") and no broad refactor riding along;
- new dependencies, if truly needed, are named explicitly.

If the request is ambiguous, stop after drafting the PRD and ask. Never invent business rules.

## Workflow

1. **Understand** — read `AGENTS.md`, this file, `make help`, the nearest sibling domain/page, and the PRD (if it exists). Not the whole repo.
2. **Draft PRD** — confirm the linked problem brief says `proceed` (or record the complete waiver fields), choose the next ID across `backlog/` and `developed/`, create it under `backlog/`, use the required format, and surface assumptions/questions. Small, clear changes may proceed right after the PRD is written. Sensitive behavior (auth, authorization, money, deletion, destructive migration) requires owner approval first — and approval is per-session: a PRD file existing does not prove it was approved.
3. **Map one vertical slice** — the smallest change that satisfies the PRD: migration/query (if needed) → proto → handler/service/repository (as needed) → `make gen` → Svelte page/component (if needed) → tests from the Test Cases. Skip layers the feature doesn't need.
4. **Implement** — translate Test Cases into automated tests early; put the `TC-<id>-n` ID in the test name or comment; stay inside Acceptance and Out of Scope.
5. **Validate** — run targeted checks while iterating, then `make check` as the final gate. If it fails, the work is not done.
6. **Promote** — only after implementation and validation succeed, `git mv` the PRD from `backlog/` to `developed/` in the same change. Do not copy it or renumber it.
7. **Report** — PRD path, acceptance met, tests added (with TC IDs), what was deliberately not done, and real risks/follow-ups. Use [`../report-template.md`](../report-template.md) as the copy-pasteable format.

## Lifecycle

- A PRD begins in `backlog/`; moving it to `developed/` asserts that its Acceptance and Test Cases have been implemented and validated.
- A developed PRD is never silently edited to match the code; requirement changes are an explicit owner revision or a new backlog PRD.
- If code later regresses, fix the regression against the developed PRD; do not move historical requirements back to backlog.
- Typo/non-behavior fixes need no PRD; behavior-changing bugfixes need a small backlog PRD.
- If acceptance turns out wrong mid-implementation, stop and ask for a PRD revision.

Traceability is plain text search — `rg 'TC-001'` must find the tests. No registry.
