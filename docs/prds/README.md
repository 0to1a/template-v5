# PRDs — the unit of work

Every behavior change starts as a file here:

```text
docs/prds/<id>-<slug>.md      e.g. docs/prds/002-logout.md
```

- ID: three digits, monotonic, never reused. The next ID is the highest existing one + 1.
- Slug: lowercase kebab-case.
- One PRD = one verifiable capability. Split big PRDs before implementing.
- A PRD is not a design document, implementation plan, or file list.
- Test-case IDs reuse the PRD ID: `TC-002-1`, `TC-002-2`, …

## Required format

```markdown
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

Do not add implementation checklists, file lists, or extra required metadata.

## Quality gate — a PRD is ready when

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
2. **Draft PRD** — next ID, short slug, required format; surface assumptions and questions. Small, clear changes may proceed right after the PRD is written. Sensitive behavior (auth, authorization, money, deletion, destructive migration) requires owner approval first — and approval is per-session: a PRD file existing does not prove it was approved.
3. **Map one vertical slice** — the smallest change that satisfies the PRD: migration/query (if needed) → proto → handler/service/repository (as needed) → `make gen` → Svelte page/component (if needed) → tests from the Test Cases. Skip layers the feature doesn't need.
4. **Implement** — translate Test Cases into automated tests early; put the `TC-<id>-n` ID in the test name or comment; stay inside Acceptance and Out of Scope.
5. **Validate** — `make check`. If it fails, the work is not done.
6. **Report** — PRD path, acceptance met, tests added (with TC IDs), what was deliberately not done, real risks/follow-ups.

## Lifecycle

- An implemented PRD is never silently edited to match the code; requirement changes are an explicit owner revision or a new PRD.
- Typo/non-behavior fixes need no PRD; behavior-changing bugfixes need a small one.
- If acceptance turns out wrong mid-implementation, stop and ask for a PRD revision.

Traceability is plain text search — `rg 'TC-001'` must find the tests. No registry.
