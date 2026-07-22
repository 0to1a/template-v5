---
name: new-feature
description: Start a new feature the PRD-driven way — draft docs/prds/NNN-<slug>.md, get approval when needed, implement one vertical slice, validate with make check. Use whenever the user asks to build, add, or change application behavior.
---

This project is PRD-driven. The full workflow, PRD format, quality gate, and
lifecycle rules live in `docs/prds/README.md` — read that file and `AGENTS.md`
first; this skill only points there and adds no rules of its own.

Then:

1. Draft `docs/prds/<next-id>-<slug>.md` in the required format. Ask about
   ambiguities now, not mid-implementation. Sensitive behavior (auth,
   authorization, money, deletion, destructive migration) needs explicit
   owner approval in this session before any code.
2. Implement the smallest vertical slice that satisfies the PRD, copying the
   shape of the nearest sibling domain (`internal/auth`) or page
   (`web/src/routes/login`). Run `make gen` after proto/SQL changes.
3. Write tests named after the PRD's test-case IDs (`TC-<id>-n`).
4. `make check` must pass before reporting done. Report against the PRD:
   acceptance met, tests added, what was left out because of Out of Scope.
