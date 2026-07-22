# Public commands (run `make help` for the list):
#
#   make bootstrap  install/download all dependencies — the ONLY target that installs
#   make gen        regenerate code from proto and SQL (explicit, no watcher)
#   make check      done-signal: codegen, lint, tests, both builds
#   make run        build the frontend once, then run the single Go process
#   make build      produce bin/server with the SPA embedded
#
# PostgreSQL is external. Migrations are embedded in the binary and applied
# by the server itself at startup (up only). No target here may create,
# reset, or destroy the database server.
#
# `.NOTPARALLEL` is a defensive guarantee that `make -j` cannot reorder or
# overlap these steps.
.NOTPARALLEL:

# The Go process reads .env itself via godotenv; `-include` + `export` makes
# every Make recipe's shell see the same values, silently doing nothing if
# .env doesn't exist.
-include .env
export

.PHONY: help bootstrap gen check check-go check-web run build _check-tools

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

check-go: _check-tools gen ## Fast backend iteration gate: generation, formatting, vet, tests, build
	@unformatted="$$(gofmt -l $(GOFMT_PATHS))"; test -z "$$unformatted" || { echo "gofmt found unformatted files:" >&2; echo "$$unformatted" >&2; exit 1; }
	$(BUF) lint
	go vet $(GO_PKGS)
	go test $(GO_PKGS)
	go build $(GO_PKGS)

check-web: _check-tools gen ## Fast frontend iteration gate: lint, typecheck, unit tests, build
	cd web && bun run lint
	cd web && bun run check
	cd web && bun run test:unit -- --run --passWithNoTests
	cd web && bun run build
	@touch web/dist/.gitkeep

check: _check-tools gen ## Done-signal: codegen, lint, tests, both builds (installs nothing)
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
	@echo "==> go test"
	go test $(GO_PKGS)
	@echo "==> web lint"
	cd web && bun run lint
	@echo "==> web check (svelte-kit sync + svelte-check)"
	cd web && bun run check
	@echo "==> web unit tests (vitest)"
	cd web && bun run test:unit -- --run --passWithNoTests
	@echo "==> web production build"
	cd web && bun run build
	@touch web/dist/.gitkeep
	@echo "==> go build"
	go build $(GO_PKGS)
	@echo "==> make check passed"

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
