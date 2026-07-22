# Vendor-neutral, reproducible container image for this template. No
# platform-specific config (Railway, Kubernetes, ...) belongs in this file —
# see docs/runbooks/container-deployment.md for how a platform should point
# its own health checks and shutdown signal at the resulting image.
#
# Mirrors `make build`'s own steps exactly: regenerate proto/sqlc code
# (`make gen`), build the frontend, then compile the single Go binary that
# embeds it.

# ---- builder ----
FROM golang:1.26-bookworm AS builder
WORKDIR /src

# Bun, pinned to the same version the Makefile requires (BUN_PINNED). The
# installer needs unzip, which the base golang image doesn't ship.
ARG BUN_VERSION=1.3.14
RUN apt-get update && apt-get install --no-install-recommends -y unzip \
    && rm -rf /var/lib/apt/lists/*
RUN curl -fsSL https://bun.sh/install | BUN_INSTALL=/usr/local bash -s "bun-v${BUN_VERSION}"
ENV PATH="/usr/local/bin:${PATH}"
# buf's generated-code plugins (e.g. web/node_modules/.bin/protoc-gen-es)
# are plain "#!/usr/bin/env node" scripts; bun runs them node-compatibly.
RUN ln -s /usr/local/bin/bun /usr/local/bin/node

COPY go.mod go.sum ./
RUN go mod download

COPY web/package.json web/bun.lock ./web/
RUN cd web && bun install --frozen-lockfile

COPY . .

# make gen
RUN go tool buf generate
RUN go tool sqlc generate

# frontend production build, embedded by cmd/server/register_frontend.go
RUN cd web && bun run build
RUN touch web/dist/.gitkeep

# single production binary
RUN CGO_ENABLED=0 go build -o /out/server ./cmd/server

# ---- runtime: no build toolchain, no shell, non-root ----
FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=builder /out/server /server
EXPOSE 8080
ENTRYPOINT ["/server"]
