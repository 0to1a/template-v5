---
type: Product requirement
title: Clean Checkout CI Validation
description: Reliable full validation from a clean checkout before generated frontend build artifacts exist.
tags: [ci, build, validation]
---

# Clean Checkout CI Validation

## Purpose
Allow maintainers to run the complete validation pipeline successfully from a clean repository checkout before frontend build artifacts exist.

## Acceptance:
- The repository contains the minimal tracked frontend distribution placeholder required by the Go embed directive.
- `make check` reaches frontend production build and Go build from a checkout with no pre-existing generated frontend distribution files.
- Production frontend build artifacts remain ignored by Git.

## Out of Scope
- No changes to CI step ordering or validation coverage.
- No changes to frontend runtime behavior or server embedding behavior.
- No new dependencies.

## Test Cases
### TC-002-1: Go validation accepts a clean frontend distribution directory
- Given a clean checkout with no generated frontend distribution files
- When the Go validation steps run before the frontend production build
- Then the frontend embed package is valid because its tracked distribution placeholder exists

### TC-002-2: Generated frontend distribution files remain untracked
- Given the frontend production build has completed
- When repository status is inspected
- Then generated distribution files are ignored and the tracked placeholder remains unchanged
