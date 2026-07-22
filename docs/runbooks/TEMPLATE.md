---
type: Runbook
title: "[Operation name]"
description: "One-sentence summary of what this runbook does and when to use it."
tags: [runbook]
status: draft # draft | active | superseded
owner: "[name or role]"
last_reviewed: YYYY-MM-DD
---

# [Operation name]

[If this implements a PRD, link it here: `Implements PRD` followed by a
link to `../prds/developed/<id>-<slug>.md`.]

## Purpose
[What this runbook accomplishes and the situation that calls for it —
routine operation, incident response, or release step.]

## Preconditions
- [Access, tooling, or approval required before starting]
- [Any state the system must already be in]

## Steps
1. [Exact command or UI action]
2. [Exact command or UI action]
3. [...]

## Verification
[Exact command(s) that prove the operation worked, and what output to
expect.]

## Rollback
[How to undo this operation if it goes wrong, or state explicitly that it
is not reversible and what to do instead.]

## Owner / escalation
[Who owns this procedure and who to contact if it fails. If this action is
approval-gated (e.g. requires GitHub Admin, production secrets, or DNS
access), say so explicitly and name why automation deliberately does not
hold that access — see `AGENTS.md`, "Never become the sole holder of
production access."]

## Test case trace
[If this runbook implements PRD test cases, list the `TC-<id>-n` each step
verifies.]
