---
type: Runbook
title: GitHub branch protection for main
description: Ruleset configuration that makes main PR-only with a required check, plus the audited break-glass path.
tags: [ci, github, security]
status: active
owner: Founding Engineer
last_reviewed: 2026-07-22
---

# GitHub branch protection for `main`

Implements PRD [`007-secure-github-contribution-path`](../prds/backlog/007-secure-github-contribution-path.md).

## Why this is an owner action, not automation

Applying or editing a repository ruleset requires the GitHub **Admin** role on the
repo. The automation identity (`albot01`) intentionally holds **push**
(collaborator/write), not Admin — otherwise it could bypass the very
protection it is meant to be constrained by. Only the human repository owner
(or another Admin they designate) can run the steps below. This is by design,
not a gap: see AGENTS.md, "Never become the sole holder of production access."

## Current state

The repository already has one ruleset (`no main`, id `19521876`) that blocks
`deletion`, `non_fast_forward`, and `update` on `main` with no bypass and no
`pull_request` / `required_status_checks` rule. As configured, that ruleset
blocks **all** updates to `main`, including legitimate PR merges. It needs to
be replaced with the configuration below, not left in place alongside it.

## Target configuration

Rule types on the existing ruleset (id `19521876`), scoped to `~DEFAULT_BRANCH`:

- `deletion` — block branch deletion.
- `non_fast_forward` — block force pushes.
- `pull_request` — require an open PR before merge.
- `required_status_checks` — require the `check` context (the CI job id in
  `.github/workflows/ci.yml`) to pass before merge.

Everything except `bypass_actors` can be applied with `gh api` because it
needs no repo-specific lookup. Save this as `ruleset.json` and run it as an
Admin:

```json
{
  "name": "protect-main",
  "target": "branch",
  "enforcement": "active",
  "conditions": {
    "ref_name": { "include": ["~DEFAULT_BRANCH"], "exclude": [] }
  },
  "rules": [
    { "type": "deletion" },
    { "type": "non_fast_forward" },
    {
      "type": "pull_request",
      "parameters": {
        "required_approving_review_count": 0,
        "dismiss_stale_reviews_on_push": true,
        "require_code_owner_review": false,
        "require_last_push_approval": false,
        "required_review_thread_resolution": false,
        "allowed_merge_methods": ["merge", "squash", "rebase"]
      }
    },
    {
      "type": "required_status_checks",
      "parameters": {
        "required_status_checks": [{ "context": "check" }],
        "strict_required_status_checks_policy": false
      }
    }
  ]
}
```

```sh
gh api --method PUT repos/0to1a/template-v5/rulesets/19521876 --input ruleset.json
```

`required_approving_review_count` is `0` deliberately: today the only
contributors are the repo owner and the automation identity, and a mandatory
second human reviewer would just block automation PRs. Raise it to `1`+ once
there is a second human reviewer to assign.

## Break-glass bypass (do this step in the GitHub UI)

GitHub does not document a stable numeric id for the built-in "Admin"
`RepositoryRole` in the public API reference, so guessing it in a script here
would risk silently misconfiguring who can bypass protection — a security
control is the wrong place to guess. Instead, in the GitHub UI:

1. Go to **Settings → Rules → Rulesets → protect-main**.
2. Under **Bypass list**, add **Repository admin**, bypass mode **Always**.
3. Save.

This makes the bypass path exactly "repository Admins, at any time" — no
wider. GitHub's audit log records every bypass event (who, when, which rule).
Treat any bypass as an incident: the admin who used it must record the reason
and outcome in a follow-up issue comment or ADR entry within one business day.

## Verification (no admin required)

```sh
# Confirms the ruleset now requires PR + "check":
gh api repos/0to1a/template-v5/rulesets/19521876

# Confirms a non-admin push to main is rejected (expect a rejection, not a
# successful push):
git push origin HEAD:main
```

## Test case trace

- TC-007-3 (direct push rejected): verified by the `git push` command above
  returning a ruleset rejection.
- TC-007-1, TC-007-2, TC-007-4: verified against `.github/workflows/ci.yml`
  and a test pull request opened by the automation identity.
