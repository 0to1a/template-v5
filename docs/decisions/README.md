---
type: Decisions index
title: Architecture decision records
description: When and how to write an ADR for a cross-cutting technical or process decision, and the index of decisions made.
tags: [architecture, decisions, governance]
status: active
owner: Founding Engineer
last_reviewed: 2026-07-22
---

# Architecture decision records

Use an ADR for a decision that is cross-cutting (affects more than one
domain or future contributors), hard to reverse, or likely to be
re-litigated without a written record — vendor/tooling choices, adopting or
rejecting an external system, or a process change. It is not for a single
PRD's implementation detail; that belongs in the PRD itself.

## Format

Copy [`TEMPLATE.md`](TEMPLATE.md) to `NNNN-<slug>.md` (four-digit,
monotonic, never reused — the next ID is the highest existing one + 1).
Front matter uses the same canonical fields as other governance docs
(`type`, `title`, `description`, `tags`, `status`, `owner`,
`last_reviewed`); `status` is one of `proposed`, `accepted`, `superseded`,
or `rejected`.

A superseded ADR is never deleted or rewritten — a new ADR records the
changed decision and links back to the one it supersedes, so the history of
*why* stays intact.

## Precedence

An ADR explains *why*; it does not override `AGENTS.md` or an approved PRD's
behavioral contract. If a decision recorded here conflicts with a later PRD,
the PRD wins for that specific behavior and the ADR should be marked
`superseded`.

## Index

- [`0001-defer-openwiki-adoption.md`](0001-defer-openwiki-adoption.md) — `accepted` — Git Markdown stays the source of truth; OpenWiki is not adopted as a build/runtime dependency now.
