# template-v5 ships as a single Go binary with the SvelteKit SPA embedded
# via go:embed. Neither Railway's nor any generic buildpack's auto-detection
# can produce that: it requires Bun (pinned 1.3.14) to build the frontend,
# then buf/sqlc codegen, then a Go build that embeds the frontend output — in
# that order. This Dockerfile is that pipeline, kept in lockstep with
# Makefile's `build` target and .github/workflows/ci.yml.
FROM golang:1.26-trixie AS build

# Bun's installer needs unzip; not present in the golang base image.
RUN apt-get update && apt-get install -y --no-install-recommends unzip \
	&& rm -rf /var/lib/apt/lists/*

# Bun pinned to match web/package.json's engines and CI's oven-sh/setup-bun.
RUN curl -fsSL https://bun.sh/install | BUN_INSTALL=/usr/local bash -s "bun-v1.3.14"
ENV PATH="/usr/local/bin:${PATH}"

# buf.gen.yaml's TS plugin (protoc-gen-es) is a `#!/usr/bin/env node` script;
# bun is a compatible runtime but doesn't register itself as `node`.
RUN ln -s /usr/local/bin/bun /usr/local/bin/node

WORKDIR /src

# Cache go module downloads separately from the full source copy.
COPY go.mod go.sum ./
RUN go mod download

# Cache bun install separately: buf.gen.yaml's TS plugin
# (web/node_modules/.bin/protoc-gen-es) must exist before `buf generate`.
COPY web/package.json web/bun.lock ./web/
RUN cd web && bun install --frozen-lockfile

COPY . .

# go tool buf/sqlc compile from the exact versions pinned in go.mod's `tool`
# block, so this never drifts from what CI and local `make check` run.
RUN go tool buf generate
RUN go tool sqlc generate

RUN cd web && bun run build
RUN touch web/dist/.gitkeep

# CGO_ENABLED=0: none of this template's dependencies (pgx, modernc.org/sqlite)
# need cgo, so the binary is fully static and has no libc dependency at all —
# confirmed with `ldd` reporting "not a valid dynamic program". That's what
# makes the Alpine runtime stage below safe: Bun/buf/sqlc never run outside
# this Debian-based build stage, so their glibc dependency never touches the
# final image.
RUN mkdir -p bin && CGO_ENABLED=0 go build -o bin/server ./cmd/server

# Alpine over debian:trixie-slim: measured final image is ~48MB vs ~157MB
# for the same binary on debian:trixie-slim (content/transfer size ~15MB vs
# ~42MB), and the static binary has no glibc/musl dependency either way, so
# there's no compatibility cost to the smaller base. See the PR description
# for the measurement and the Bun/musl caveat this decision considered and
# ruled out (Bun only runs in the build stage above).
FROM alpine:3.22
RUN apk add --no-cache ca-certificates

COPY --from=build /src/bin/server /usr/local/bin/server

ENV PORT=8080
EXPOSE 8080
CMD ["server"]
