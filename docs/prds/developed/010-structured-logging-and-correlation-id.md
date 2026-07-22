---
type: Product requirement
title: Structured logging, request correlation ID, and a vendor-neutral observability interface
description: Every HTTP request gets a correlation ID and one structured access-log line, behind a small logging interface no specific observability vendor is baked into.
tags: [runtime, observability, operations]
problem_brief: waiver
waiver_owner: Founding Engineer
waiver_reason: Internal runtime hardening for template-v5 itself, directed by the approved ALV-7 hardening plan (Fase 3 / Wave C); not a customer-facing capability requiring product-market evidence.
waiver_expires: 2026-08-22
---

# Structured logging, request correlation ID, and observability interface

## Purpose
Let an operator trace one request end to end across log lines and swap in a real observability vendor later without touching call sites, for whoever is on call for a product built from this template.

## Acceptance:
- `internal/platform/observability` defines a small `Logger` interface (not tied to any specific vendor SDK) with a default implementation that writes structured (JSON) log lines.
- Every HTTP response carries an `X-Request-Id` header: the incoming request's own header value if present, otherwise a freshly generated one.
- Every HTTP request produces exactly one structured log line with the correlation ID, method, path, status code, and duration.
- The request-logging middleware never logs header values or request/response bodies, so bearer tokens, login codes, and other secrets cannot leak into logs through it.

## Out of Scope
- Migrating existing `log.Printf` calls inside `internal/auth` or `internal/mail` to the new interface — those stay as they are.
- Adding a metrics or tracing vendor SDK (Prometheus, OpenTelemetry, Sentry, etc.) — the interface only makes room for one later.
- Log shipping, retention, or aggregation infrastructure.

## Test Cases
### TC-010-1: Response carries a correlation ID
- Given a request with no `X-Request-Id` header
- When the request-logging middleware handles it
- Then the response has a non-empty `X-Request-Id` header

### TC-010-2: An incoming correlation ID is reused, not replaced
- Given a request that already sets `X-Request-Id: existing-id`
- When the request-logging middleware handles it
- Then the response's `X-Request-Id` header is exactly `existing-id`

### TC-010-3: One structured log line per request, with no header or body content
- Given a request carrying an `Authorization` header
- When the request-logging middleware handles it
- Then exactly one log call is recorded, its fields include method, path, status, duration, and the correlation ID, and none of its argument values contain the `Authorization` header's value
