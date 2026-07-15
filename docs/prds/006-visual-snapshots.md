# Static UI Snapshots via `make visual`

## Purpose
Let engineers run `make visual` to render every page into a deterministic, git-committed HTML file from fixed fixture data, so UI changes are visible in `git diff` and past snapshots serve as a reference for product enhancement.

## Acceptance:
- `make visual` renders every route with a `+page.svelte` into a static HTML file, without requiring the Go backend or Postgres to be running.
- Rendering is driven by hand-authored fixture data per route rather than live network responses.
- Running `make visual` twice with no code or fixture changes produces byte-identical HTML output.
- Snapshot files live under a dedicated, git-tracked directory so `git diff` on that directory shows what changed.
- A route without a matching fixture causes `make visual` to fail with an error naming the route, instead of silently skipping it.

## Out of Scope
- Screenshot/image-based visual diffing — this PRD only produces static HTML text files.
- Automatically recording or generating fixtures from a live backend — fixtures are hand-authored only.
- Wiring `make visual` into CI or blocking merges on snapshot diffs.

## Test Cases
### TC-006-1: `make visual` generates a snapshot per route
- Given every existing route has a matching fixture file
- When a maintainer runs `make visual`
- Then a static HTML file exists for each route under the git-tracked snapshot directory

### TC-006-2: Snapshots do not require a live backend
- Given the Go server and Postgres are not running
- When a maintainer runs `make visual`
- Then it completes successfully and produces snapshots using fixture data instead of live responses

### TC-006-3: Snapshots are deterministic
- Given no code or fixture changes between runs
- When `make visual` is run twice in a row
- Then the two runs produce byte-identical HTML files for every route

### TC-006-4: Missing fixture blocks generation
- Given a route's `+page.svelte` has no matching fixture file
- When a maintainer runs `make visual`
- Then the command fails with an error identifying the route missing a fixture
