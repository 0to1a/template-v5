---
type: Operations
title: Environment contract
description: Every environment variable the server reads, whether it's required, its default, and what to set it to in production.
tags: [operations, configuration, security]
status: active
owner: Founding Engineer
last_reviewed: 2026-07-22
---

# Environment contract

The server reads configuration once, at startup, in `internal/platform/config/config.go`.
A missing `.env` file is not an error; `.env` values never override variables
already set in the process environment (see `.env.example` for a local
template). Startup fails closed — a missing or malformed required value
aborts before the server binds a port — see
[`prds/developed/011-startup-config-validation-hardening.md`](prds/developed/011-startup-config-validation-hardening.md).

| Variable | Required | Default | Production guidance |
|---|---|---|---|
| `PORT` | No | `8080` | Must be a valid TCP port (1–65535); set to whatever the platform expects the process to bind. |
| `DATABASE_URL` | Yes | none | A real `postgres://` or `postgresql://` URL. Never log this value — it may embed a password. |
| `JWT_SECRET` | Yes | none | At least 32 random bytes, generated per environment, never reused between environments, never committed. Rotating it invalidates every issued token and every user's derived TOTP secret. |
| `MAIL_URL` | No | unset (login codes are discarded, not sent) | An `smtp://` URL. Leaving this unset in production means no user can ever receive a login code — verify it is set before relying on login in a real deployment. |
| `IS_GUEST_REGISTRATION` | No | `0` (disabled) | Set to `1` only if the product intentionally auto-creates an account for any email that requests a login code. Treat enabling this in production as a product decision, not a default. |
| `APP_ENV` | No | `production` (fail closed) | `development` or `production` — any other value aborts startup. `development` is the only setting where the seeded `admin@localhost` static login code (`123456`) works; leave this unset (or explicitly `production`) in every real deployment. See [`prds/developed/014-production-guardrail-disable-demo-login-credential.md`](prds/developed/014-production-guardrail-disable-demo-login-credential.md). |

## Health endpoints

- `GET /health` — liveness. No dependencies, always 200 if the process is
  up. Use this for "should this instance be restarted".
- `GET /health/ready` — readiness. Pings the database pool with a 2-second
  timeout; non-2xx means the process is up but cannot serve traffic yet
  (e.g. database still starting). Use this for "should this instance
  receive traffic" — see
  [`prds/developed/012-liveness-and-readiness-health-endpoints.md`](prds/developed/012-liveness-and-readiness-health-endpoints.md).

## Logging

Every request produces one structured (JSON) log line to stdout carrying a
correlation ID (`X-Request-Id`, echoed on the response), method, path,
status, and duration — never header values or bodies, so secrets cannot
leak through it. See
[`prds/developed/010-structured-logging-and-correlation-id.md`](prds/developed/010-structured-logging-and-correlation-id.md).
