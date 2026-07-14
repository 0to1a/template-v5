# Prebuilt CI Code Generation Tools

## Purpose
Allow maintainers to run CI validation with version-matched prebuilt Buf and sqlc binaries instead of compiling those tools on a cold runner.

## Acceptance:
- CI installs prebuilt Buf and sqlc versions that match the versions pinned in `go.mod`.
- CI passes the prebuilt tool commands to the existing generation and lint targets.
- Local validation continues to use the tools pinned in `go.mod` by default.
- The complete validation pipeline remains unchanged apart from tool provisioning.

## Out of Scope
- No dependency version upgrades.
- No removal of the local `go tool` workflow.
- No changes to application behavior or deployment.

## Test Cases
### TC-003-1: CI uses version-matched prebuilt tools
- Given a clean CI runner and tool versions pinned in `go.mod`
- When the CI validation job starts
- Then it installs matching prebuilt Buf and sqlc binaries and uses them for generation and linting

### TC-003-2: Local validation retains pinned defaults
- Given no Buf or sqlc command overrides
- When a maintainer runs the local validation pipeline
- Then generation and linting use the tools pinned through `go.mod`

### TC-003-3: Tool commands can be overridden
- Given compatible Buf and sqlc commands are available
- When a maintainer supplies those commands to the validation target
- Then generation and linting use the supplied commands without changing other validation steps
