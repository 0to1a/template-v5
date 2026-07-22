# Template v5

An AI-first full-stack template with a **Go 1.26, Connect RPC, and PostgreSQL**
backend; a **Svelte 5 and SvelteKit SPA** frontend; and a single production
binary. Development follows a **PRD-driven workflow**: every behavior change
starts with a small PRD in [`docs/prds/backlog/`](docs/prds/README.md), and repository
working rules are defined in [`AGENTS.md`](AGENTS.md).

**Start here:** [`docs/onboarding/developer.md`](docs/onboarding/developer.md)
is the entry point for setting up a local environment — it's the one
deterministic path from a fresh clone to a healthy app, verified by a
non-mutating `make doctor` check.

## Setup

Requirements: Go (see `go.mod`), Bun 1.3.14, and an external PostgreSQL.

```bash
cp .env.example .env     # set DATABASE_URL and a >=32-byte JWT_SECRET
make bootstrap           # installs all dependencies (the only target that does)
make doctor               # read-only preflight: Go, Bun, PostgreSQL, config
make check               # the done-signal: codegen, lint, tests, both builds
make run                 # build the frontend, then run the single server process
```

`make help` lists every command. Schema migrations are embedded in the
binary and applied automatically at server startup (up only); the server
never creates, drops, or resets the database itself. See
[`docs/onboarding/developer.md`](docs/onboarding/developer.md) for the full
walkthrough, including what `make doctor` checks and known first-run pitfalls.

## Security notes

- The seeded `admin@localhost` account accepts the static OTP `123456`
  (exact-match only). Remove or protect it before any untrusted deployment.
- Other accounts use a 5-minute TOTP derived from `JWT_SECRET`; no email
  provider is wired up yet, so they cannot receive codes out of the box.
- An OTP can be replayed within its own 5-minute step (documented limitation).
- The bearer token is stored in `localStorage`; an XSS in this origin could
  read it. Moving to HttpOnly cookies is a future PRD.
