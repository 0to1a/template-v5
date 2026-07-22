---
type: Product onboarding
title: Per-product onboarding contract
description: Defines the minimum product context and measurable first-value journey required for each SaaS venture.
tags: [product, onboarding, activation]
status: active
owner: Product
last_reviewed: 2026-07-22
---

# Per-product onboarding contract

Complete this document (or copy its sections into a venture-specific onboarding document) **after** a problem brief reaches `proceed` and **before** designing onboarding screens. The goal is to make a new target user reach measurable first value safely and quickly—not to prescribe a universal tour or checklist.

## 1. Context and persona

- **Product / problem brief:** [name and relative link to a `proceed` brief]
- **Primary persona:** [role, company/user context, relevant skill level]
- **JTBD and trigger:** [situation → motivation → desired outcome]
- **Excluded personas:** [who this path is not for]
- **Evidence basis and date:** [direct customer sources; separate inference]
- **Owner / last reviewed:** [role or name] / YYYY-MM-DD

Do not call the persona validated when it rests only on desk research or internal opinion.

## 2. Prerequisites

List only conditions required before the user can reach value.

| Prerequisite | Why required | How detected | User remediation | Can defer? |
|---|---|---|---|---|
| [account permission, source data, integration, domain knowledge] | [...] | [...] | [...] | yes/no |

Never request credentials or data before they are necessary. Identify external-contact, personal-data, regulated, payment, and legal approval gates.

## 3. First value and activation

- **First-value statement:** The user has received value when [observable user outcome, not “completed onboarding”].
- **Activation event:** `[stable_event_name]` fires once when [server-verifiable or clearly observable condition].
- **Activation window:** within [duration] of [account/workspace creation or qualifying trigger].
- **Target time-to-value (TTV):** p50 ≤ [duration], p90 ≤ [duration]. State baseline or `unknown` and how it will be measured.
- **Non-activation:** [conditions that must not count, such as viewing a page or synthetic/demo data].
- **Repeat-value signal:** [behavior showing the job recurs, with timeframe].

Activation is a product hypothesis. Link it back to evidence and revise it when behavior contradicts the hypothesis.

## 4. Smallest journey to value

| Step | User intent | Required input/action | System feedback | Instrumentation | Recovery |
|---|---|---|---|---|---|
| 1 | [...] | [...] | [...] | [...] | [...] |

Remove optional profile setup, tours, and configuration from the critical path. Offer skip/defer where they do not block first value.

## 5. Instrumentation contract

Define semantics before implementation; use no raw secrets or unnecessary personal data in events.

| Event | Trigger and once/repeat rule | Required properties | Excluded data | Owner |
|---|---|---|---|---|
| `onboarding_started` | first entry by a qualifying persona; once per workspace | source, persona_version | email, token, source content | Product |
| `[prerequisite]_completed` | prerequisite becomes valid; repeat only after invalidation | method, elapsed_ms | credentials | Product |
| `[activation_event]` | exact first-value condition; once per workspace | elapsed_ms, journey_version | customer content/PII | Product |
| `onboarding_blocked` | a blocking state is shown | state_code, step, recoverable | free-text error/secret | Product |

Minimum funnel: eligible → started → prerequisite complete → activated. Segment by approved, low-cardinality acquisition/persona fields. Define denominator, observation window, identity merge behavior, timezone, test/internal traffic exclusion, dashboard owner, and data retention before interpreting activation rate.

## 6. Empty, error, and recovery states

For every state, explain what happened, preserve valid work, and provide one safe next action.

| State | Detection | User message must convey | Primary recovery | Escalation / telemetry |
|---|---|---|---|---|
| First-use empty | valid setup, no user-created item | nothing exists yet and why first action matters | create/import minimum valid item | track empty CTA |
| Source-data empty | connection works, zero qualifying records | connection succeeded; no qualifying data found | adjust scope or use approved sample | track reason code |
| Permission blocked | required permission absent | which capability is unavailable, without leaking sensitive detail | request permission or choose allowed path | owner/admin route |
| Validation error | user input is invalid | field-level cause and valid example | correct without losing input | stable error code |
| Transient system/integration error | timeout/unavailable dependency | action did not complete; whether retry is safe | retry with idempotency/backoff | correlation ID, status path |
| Partial success | some work succeeded | what completed and what did not | retry failed subset; preserve success | counts, no sensitive payload |
| Session/interruption recovery | user returns after exit/expiry | saved progress and next incomplete step | resume or restart explicitly | resumed event |
| Irrecoverable state | policy/data makes continuation impossible | reason category and consequences | safe exit/contact owner | runbook/support path |

Never use fake success, silent data loss, dead-end errors, or automatic destructive reset as recovery.

## 7. Success and guardrails

- **Primary:** eligible-user activation rate within [window], target [x%].
- **TTV:** p50/p90 from defined start to activation, targets above.
- **Drop-off:** per required journey step; investigate when [threshold].
- **Quality guardrail:** [first-value output correctness/acceptance].
- **Safety guardrail:** [privacy, permission, error, or complaint threshold].
- **Retention proxy:** [repeat-value event within timeframe].
- **Decision date:** on YYYY-MM-DD, Product will proceed, revise, or remove the onboarding path based on [thresholds].

## 8. Pre-handoff checklist

- [ ] Persona, JTBD, and activation hypothesis trace to a `proceed` problem brief.
- [ ] Prerequisites and exclusions are explicit; optional work is off the critical path.
- [ ] First value is an observable outcome; activation semantics and TTV targets are precise.
- [ ] Events have owner, trigger, properties, privacy exclusions, denominator, and test-traffic rule.
- [ ] Empty, permission, validation, transient, partial, interruption, and irrecoverable states have recovery.
- [ ] Contradictory evidence, confidence, open risks, and next decision date are recorded.
- [ ] Behavior changes move into a small backlog PRD with acceptance and traceable `TC-*` tests.
