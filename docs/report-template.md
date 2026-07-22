---
type: Report template
title: Completion report template
description: The definition-of-done contract every finished PRD reports against — copy this section into the final PR/issue update.
tags: [prds, definition-of-done, reporting]
status: active
owner: Founding Engineer
last_reviewed: 2026-07-22
---

# Completion report template

This is step 7 ("Report") of [`prds/README.md`](prds/README.md)'s workflow,
made copy-pasteable. A PRD is not done until this report can be filled in
truthfully.

```markdown
## Completion report: <PRD id and title>

- PRD: `docs/prds/developed/<id>-<slug>.md` (moved from `backlog/` in this change)
- Acceptance: met | partially met — [note any deviation and why]
- Tests added: TC-<id>-1, TC-<id>-2, ... — [file paths]
- Out of Scope respected: yes | no — [note anything that had to be revisited]
- Validation: targeted `go test`/`go build` run during iteration; `make check` result: [pass/fail + link]
- Residual risk / follow-ups: [known gaps, deferred work, or new backlog PRDs opened]
- Owner / reviewer: [who approved sensitive behavior, if applicable]
```

## Why every field is required

- **PRD path** makes the report traceable to the exact contract implemented, not a paraphrase of it.
- **Acceptance status** must be explicit — "partially met" with a reason is honest; a silent gap is not.
- **Test IDs** are what `rg 'TC-<id>'` finds; a report without them cannot be verified independently of the author's word.
- **Out of Scope respected** catches quiet scope creep before it ships.
- **Validation** distinguishes targeted iteration checks from the final `make check` gate — both matter, neither substitutes for the other.
- **Residual risk** is where a real limitation belongs, instead of being silently absorbed into "done."
