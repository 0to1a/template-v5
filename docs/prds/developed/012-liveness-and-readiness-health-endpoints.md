---
type: Product requirement
title: Liveness and readiness health endpoints
description: Add a database-backed GET /health/ready alongside the existing dependency-free GET /health, so a load balancer can tell "process is up" apart from "process can serve traffic".
tags: [runtime, operations, reliability]
problem_brief: waiver
waiver_owner: Founding Engineer
waiver_reason: Internal runtime hardening for template-v5 itself, directed by the approved ALV-7 hardening plan (Fase 3 / Wave C); not a customer-facing capability requiring product-market evidence.
waiver_expires: 2026-08-22
---

# Liveness and readiness health endpoints

## Purpose
Let a deployment platform distinguish "the process is alive" from "the process can actually reach its database", for the operator running any product built from this template.

## Acceptance:
- `GET /health` keeps its existing behavior unchanged: always 200, no dependency on the database.
- A new `GET /health/ready` returns 200 when the database pool can be reached within a short timeout, and a non-2xx status when it cannot.
- `/health/ready` never blocks indefinitely — an unreachable database is bounded by a fixed timeout, not a hang.
- `docs/architecture.md` documents the distinction between the two endpoints.

## Out of Scope
- Readiness checks for anything beyond the database (mail delivery, disk space, etc.).
- Wiring these endpoints into a specific platform's health-check configuration (Railway, Kubernetes) — that belongs in the container-deployment runbook, not code.
- Authentication on either health endpoint — both remain public, matching the existing `/health` contract.

## Test Cases
### TC-012-1: Liveness is unaffected by database state
- Given no database dependency is passed to the liveness handler
- When `GET /health` is requested
- Then it returns 200 with the existing body regardless of database reachability

### TC-012-2: Readiness succeeds when the database is reachable
- Given a pinger that succeeds
- When `GET /health/ready` is requested
- Then it returns 200

### TC-012-3: Readiness fails when the database is unreachable
- Given a pinger that returns an error
- When `GET /health/ready` is requested
- Then it returns a non-2xx status and a body distinct from the healthy response
