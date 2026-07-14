# Template v5

AI-first full-stack template: **Go 1.26 + Connect RPC + PostgreSQL** backend,
**Svelte 5 (SvelteKit SPA)** frontend, one production binary, and a
**PRD-driven workflow** — every behavior change starts as a small PRD in
`docs/prds/` (see `docs/prds/README.md`). Working rules live in `AGENTS.md`; the full
design rationale in `docs/final.md`.

## Setup

Requirements: Go (see `go.mod`), Bun 1.3.14, and an external PostgreSQL.

```bash
cp .env.example .env     # set DATABASE_URL and a >=32-byte JWT_SECRET
make bootstrap           # installs all dependencies (the only target that does)
make check               # the done-signal: codegen, lint, tests, both builds
make run                 # build the frontend, then run the single server process
```

`make help` lists every command. Schema migrations are embedded in the
binary and applied automatically at server startup (up only); the server
never creates, drops, or resets the database itself.

## Security notes (initial version)

- The seeded `admin@localhost` account accepts the static OTP `123456`
  (exact-match only). Remove or protect it before any untrusted deployment.
- Other accounts use a 5-minute TOTP derived from `JWT_SECRET`; no email
  provider is wired up yet, so they cannot receive codes out of the box.
- An OTP can be replayed within its own 5-minute step (documented limitation).
- The bearer token is stored in `localStorage`; an XSS in this origin could
  read it. Moving to HttpOnly cookies is a future PRD.
