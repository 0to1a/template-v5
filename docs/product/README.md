---
type: Product guide
title: Product discovery and context gate
description: Canonical workflow from customer evidence to an implementable product requirement.
tags: [product, discovery, governance]
status: active
owner: Product
last_reviewed: 2026-07-22
---

# Product discovery and context gate

Use this directory to establish **why** a capability should exist before a PRD defines **what behavior** to build. Git Markdown is the source of truth.

## Required chain

```text
customer evidence → problem brief (`proceed`) → PRD → acceptance/TC IDs → implementation
```

1. Copy [`problem-brief-template.md`](problem-brief-template.md) for one target user and problem.
2. Record direct, dated customer evidence separately from desk research and inference. Desk research alone is not validation.
3. Decide `proceed`, `pivot`, or `park`; name confidence, contradictory evidence, gaps, owner, and decision date.
4. Create a backlog PRD only after `proceed`. Its `problem_brief` front-matter field must link to that brief.
5. If urgent work genuinely has no problem brief, use a time-bounded waiver in PRD front matter: `problem_brief: waiver`, `waiver_owner`, `waiver_reason`, and `waiver_expires`. The waiver owner is accountable for replacing it with evidence or parking the capability. A waiver does not bypass approval for sensitive behavior.

A PRD with neither a `proceed` brief nor a complete waiver is not ready for implementation. Product owns the decision/evidence; engineering owns implementation feasibility. Future lint should read these fields rather than maintain a second registry.

## Decision threshold

`proceed` means there is enough evidence to justify the **next bounded experiment or MVP slice**, not proof that the venture will succeed. Before proceeding, the brief must identify:

- one target user and urgent job;
- at least one dated direct-customer evidence item, or explicitly state the evidence gap;
- current alternatives and a credible urgency or willingness-to-pay signal;
- the riskiest remaining assumption and cheapest falsification experiment;
- contradictory evidence and open risks.

Use `pivot` when the evidence supports a materially different user/problem/channel. Use `park` when pain, access, willingness to pay, or strategic fit is too weak to justify another experiment.

## Source hierarchy and ownership

For repository work, precedence is `AGENTS.md` → domain documentation → approved PRD. The problem brief explains rationale but does not override an approved PRD's behavioral contract. If new evidence changes the requirement, revise through the PRD workflow rather than silently editing a developed PRD.

## Worked example (synthetic)

The following pair demonstrates formatting and traceability only; its interviews, quotes, and commercial signals are explicitly synthetic and must never be cited as Alvin Co validation:

- [Synthetic problem brief: feedback triage](examples/feedback-triage-problem-brief.md)
- [Synthetic linked PRD: weekly feedback digest](examples/feedback-triage-prd.md)

Replace every synthetic item with auditable evidence before using the example for an investment or build decision.
