---
type: Product requirement
title: Secure GitHub Contribution Path
description: Lock main to reviewed, CI-checked pull requests with a least-privilege automation identity and an audited break-glass path.
tags: [ci, github, security]
---

# Secure GitHub Contribution Path

## Purpose
Give maintainers and automation identities a `main` branch that only accepts changes through a pull request with a deterministic, required CI check, while keeping an audited emergency bypass for repository admins.

## Acceptance:
- The CI workflow reports a single, stably-named required check called `check` and grants its `GITHUB_TOKEN` read-only permissions by default.
- A documented branch ruleset for `main` requires an open pull request and a passing `check` status before merge, and rejects direct or force pushes from any identity without the repository Admin role.
- The ruleset's only bypass path is scoped to the repository Admin role, so any emergency direct push is inherently limited to admins and recorded in GitHub's audit trail.
- The automation identity (a collaborator without Admin) can open a pull request against `main` and have the `check` workflow run to completion without needing elevated permissions.

## Out of Scope
- No change to required reviewer/approval counts beyond the pull-request gate itself.
- No change to application runtime behavior, authentication, or business logic.
- No organization-wide policy; scoped to this one repository.
- No new dependencies.

## Test Cases
### TC-007-1: CI required check name is deterministic
- Given `.github/workflows/ci.yml` has a single job with a fixed id and no matrix
- When the workflow runs on a pull request or a push to `main`
- Then the reported check run is named `check`

### TC-007-2: CI workflow token is least-privilege
- Given the `ci.yml` workflow file
- When its `permissions` block is inspected
- Then `GITHUB_TOKEN` is scoped to `contents: read` with no other write permissions

### TC-007-3: Direct push to main is rejected for non-admins
- Given an identity with push (not Admin) access to the repository
- When it attempts to push a commit directly to `main`
- Then GitHub rejects the push per the branch ruleset

### TC-007-4: Automation identity opens a PR and CI runs
- Given the automation identity has push access and opens a branch with a change
- When it opens a pull request against `main`
- Then the `check` workflow run triggers on that PR and completes
