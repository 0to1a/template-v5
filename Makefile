# Public commands (run `make help` for the list):
#
#   make bootstrap   install/download all dependencies — the ONLY target that installs
#   make doctor      diagnose Go/Bun/PostgreSQL/config readiness — read-only, installs nothing
#   make doc-lint    validate docs/ front matter, links, PRD IDs/backlinks/TC traces — read-only
#   make vuln-scan   Go + JS dependency vulnerability scan, threshold + exception policy
#   make lint        static checks: buf lint, gofmt, go vet, web lint/check
#   make test        go test + web unit tests
#   make gen         regenerate code from proto and SQL (explicit, no watcher)
#   make check       done-signal: composes doc-lint, vuln-scan, lint, test, build
#   make smoke-test  build bin/server and prove boot/migration/health/shutdown against ephemeral PostgreSQL (needs docker)
#   make run         build the frontend once, then run the single Go process
#   make build       produce bin/server with the SPA embedded
#
# TC-010-6: doc-lint, vuln-scan, lint, and test are split out of check so a
# developer can run one targeted, fast check while iterating; check itself
# still runs every one of them plus build, and stays the single
# done-signal/CI gate.
#
# PostgreSQL is external. Migrations are embedded in the binary and applied
# by the server itself at startup (up only). No target here may create,
# reset, or destroy the database server — smoke-test is the one exception,
# and only ever against a throwaway container it starts and stops itself.
#
# `.NOTPARALLEL` is a defensive guarantee that `make -j` cannot reorder or
# overlap these steps.
.NOTPARALLEL:

# The Go process reads .env itself via godotenv; `-include` + `export` makes
# every Make recipe's shell see the same values, silently doing nothing if
# .env doesn't exist.
-include .env
export

.PHONY: help bootstrap doctor doc-lint vuln-scan lint test gen check smoke-test run build _check-tools

# Scoped explicitly to our own Go code. A bare "./..." would also crawl
# web/node_modules, which can contain vendored .go files shipped inside npm
# packages — wasted time at best, a spurious failure on code we don't own at
# worst. web/embed.go and db/embed.go are included explicitly so their only
# Go files are covered without recursing into siblings.
GO_PKGS := ./cmd/... ./internal/... ./db ./web
GOFMT_PATHS := cmd internal db/embed.go web/embed.go

GO_REQUIRED := 1.26
BUN_PINNED := 1.3.14

help: ## List available commands
	@grep -hE '^[a-z-]+:.*## ' $(MAKEFILE_LIST) | awk -F':.*## ' '{printf "  make %-10s %s\n", $$1, $$2}'

# Fail fast with a clear message if required tools are missing. This never
# installs anything itself.
_check-tools:
	@command -v go >/dev/null 2>&1 || { echo "error: 'go' is required but was not found in PATH" >&2; exit 1; }
	@go_version="$$(go env GOVERSION)"; \
	case "$$go_version" in \
		go$(GO_REQUIRED)|go$(GO_REQUIRED).*) ;; \
		*) echo "error: this project requires Go $(GO_REQUIRED), but 'go env GOVERSION' reports $$go_version" >&2; exit 1;; \
	esac
	@command -v bun >/dev/null 2>&1 || { echo "error: 'bun' is required but was not found in PATH" >&2; exit 1; }
	@bun_version="$$(bun --version)"; \
	if [ "$$bun_version" != "$(BUN_PINNED)" ]; then \
		echo "error: this project pins bun to $(BUN_PINNED), but 'bun --version' reports $$bun_version" >&2; \
		exit 1; \
	fi

bootstrap: _check-tools ## Install/download all dependencies (the only target that installs)
	@echo "==> go mod download"
	go mod download
	@echo "==> installing web dependencies"
	cd web && bun install --frozen-lockfile
	@echo "==> bootstrap done"

# Deliberately independent of _check-tools: that target aborts on the first
# missing tool, but doctor's whole point is to report every problem at once
# (Go, Bun, config, PostgreSQL) before a developer starts fixing things one
# at a time. It never installs a tool, writes .env, or touches the database
# — see cmd/doctor and internal/platform/doctor. Like every other target
# here, `go run` follows this repo's normal GOTOOLCHAIN=auto behavior (see
# go.mod's `go` directive), not a doctor-specific exception.
doctor: ## Diagnose Go/Bun/PostgreSQL/config readiness (read-only, installs nothing)
	@command -v go >/dev/null 2>&1 || { \
		echo "✗ Go toolchain: go was not found in PATH"; \
		echo "    → install Go $(GO_REQUIRED).x from https://go.dev/dl/"; \
		exit 1; \
	}
	@go run ./cmd/doctor

# Read-only, like doctor: never edits a doc, never writes .env, never
# touches the database. A step in `check`/CI (Fase 4 / PRD 010); PRD 008
# deliberately left this unwired, which is why this comment moved here.
doc-lint: ## Validate docs/ front matter, links, PRD ID/backlink/TC-trace (read-only)
	@go run ./cmd/doclint

# Go side uses govulncheck (a go.mod tool dependency, same pattern as buf/
# sqlc below); JS side uses bun's built-in `bun audit`. Both go through
# internal/platform/vulnscan so one exceptions file/threshold policy
# applies to both ecosystems. Read-only: never edits go.mod/go.sum or
# package.json/bun.lock.
vuln-scan: _check-tools ## Go + JS dependency vulnerability scan (threshold + exception policy)
	@go run ./cmd/vulnscan

# buf and sqlc default to the go.mod-pinned tools, compiled on demand by the
# Go toolchain. CI overrides these with prebuilt binaries of the same versions
# because compiling buf and sqlc from source costs minutes on a cold cache.
BUF ?= go tool buf
SQLC ?= go tool sqlc

# bun install must have run before gen: buf.gen.yaml's TS plugin is
# web/node_modules/.bin/protoc-gen-es, which only exists once web deps are
# installed. Run `make bootstrap` first on a fresh clone.
gen: ## Regenerate code from proto and SQL (buf + sqlc)
	@echo "==> generating code (buf, sqlc)"
	$(BUF) generate
	$(SQLC) generate

lint: _check-tools ## Static checks: buf lint, gofmt, go vet, web lint/check
	@echo "==> buf lint"
	$(BUF) lint
	@echo "==> gofmt"
	@unformatted="$$(gofmt -l $(GOFMT_PATHS))"; \
	if [ -n "$$unformatted" ]; then \
		echo "gofmt found unformatted files:" >&2; \
		echo "$$unformatted" >&2; \
		exit 1; \
	fi
	@echo "==> go vet"
	go vet $(GO_PKGS)
	@echo "==> web lint"
	cd web && bun run lint
	@echo "==> web check (svelte-kit sync + svelte-check)"
	cd web && bun run check

test: _check-tools ## go test + web unit tests (vitest)
	@echo "==> go test"
	go test $(GO_PKGS)
	@echo "==> web unit tests (vitest)"
	cd web && bun run test:unit -- --run --passWithNoTests

check: _check-tools gen doc-lint vuln-scan lint test build ## Done-signal: composes doc-lint, vuln-scan, lint, test, build (installs nothing)
	@echo "==> go build (every package, not just cmd/server)"
	go build $(GO_PKGS)
	@echo "==> make check passed"

# Requires docker: starts its own throwaway PostgreSQL container (never the
# developer's configured DATABASE_URL) and always tears it down, even on
# failure. Not a `check` prerequisite — CI runs it as its own step so a
# missing/broken docker locally never blocks the fast targeted checks above.
smoke-test: build ## Build bin/server and prove boot/migration/health/shutdown against ephemeral PostgreSQL (needs docker)
	@./scripts/smoke-test.sh

run: _check-tools gen ## Build the frontend once, then run the single Go process
	@echo "==> building web frontend"
	cd web && bun run build
	@touch web/dist/.gitkeep
	@echo "==> running server on $${PORT:-8080}"
	go run ./cmd/server

build: _check-tools gen ## Produce the single production artifact: bin/server
	@echo "==> web production build"
	cd web && bun run build
	@touch web/dist/.gitkeep
	@echo "==> building bin/server"
	mkdir -p bin
	go build -o bin/server ./cmd/server
