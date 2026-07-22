---
type: Runbook
title: "Container deployment"
description: "How to build and run the reproducible, vendor-neutral container image for this template, and how a platform's health checks should point at it."
tags: [runbook, deployment, container]
status: active
owner: Founding Engineer
last_reviewed: 2026-07-22
---

# Container deployment

## Purpose
Build one reproducible artifact that runs the same way locally, in CI, and
on any container-capable platform, without baking in a specific vendor
(Railway, Kubernetes, etc.) — this is an example a venture may adapt, not a
platform integration.

## Preconditions
- Docker (or a compatible builder) available.
- `DATABASE_URL`, `JWT_SECRET`, and any other required variables
  (`docs/environment-contract.md`) available to inject at container run
  time — never baked into the image.

## Steps
1. Build the image from the repository root:
   ```sh
   docker build -t template-v5 .
   ```
   This runs the same two build steps `make build` does — `bun run build`
   for the frontend, then `go build ./cmd/server` — inside a multi-stage
   `Dockerfile`, producing a minimal runtime image with no build toolchain
   in it.
2. Run it, providing configuration as environment variables (never baked
   into the image or committed):
   ```sh
   docker run --rm -p 8080:8080 \
     -e DATABASE_URL="postgres://user:password@host:5432/app" \
     -e JWT_SECRET="$(openssl rand -hex 32)" \
     template-v5
   ```
3. Point the platform's own health-check configuration at:
   - `GET /health` for liveness/restart decisions.
   - `GET /health/ready` for readiness/traffic-admission decisions.
   (Exact platform configuration — e.g. Railway's healthcheck path setting
   — is platform-specific and out of scope for this vendor-neutral
   example; set it to these paths in whatever mechanism the platform uses.)
4. Configure the platform to send `SIGTERM` on stop/redeploy (the container
   default for `docker stop`) and to allow at least the shutdown timeout
   defined in `cmd/server/server.go` before force-killing, so in-flight
   requests can drain (`docs/architecture.md`, "Deployment shape").

## Verification
```sh
curl -f http://localhost:8080/health
curl -f http://localhost:8080/health/ready
```
Both return 200 once the container is fully started and can reach the
configured `DATABASE_URL`.

## Rollback
Stop the container and run the previous image tag; no state is stored in
the image itself. For a bad migration shipped inside a release, follow
`migration-rollback.md`, not this runbook.

## Owner / escalation
Owned by whoever manages deployment for the specific product. See
`release.md` for the full release process this container fits into.

## Test case trace
Not implemented in code — see `docs/prds/developed/013-fase-3-operational-documentation.md`,
TC-013-4, which verifies the `Dockerfile`'s stages correspond to this
repository's own real build steps.
