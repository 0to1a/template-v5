---
type: Product requirement
title: Multi-company vertical slice (feature-delivery benchmark)
description: Let an authenticated user create a company, hold a role within it, and select it as their active company, so exactly one tenant-scoped example resource can be created, read, and denied across companies strictly at the server.
tags: [auth, authorization, multi-tenancy, benchmark]
problem_brief: waiver
waiver_owner: Founding Engineer
waiver_reason: Internal feature-delivery benchmark (ALV-39) directed by ALV-7's approved hardening plan and its attached "Feature Delivery Benchmark Thesis" issue document; identical PRD text is run unmodified against two template baselines to measure AI delivery cost/speed, not a customer-facing capability requiring product-market evidence.
waiver_expires: 2026-08-22
---

# Multi-company vertical slice (feature-delivery benchmark)

## Purpose
Give an authenticated user the ability to create a company, hold a membership role in it, select it as their active company, and manage one tenant-scoped example resource that only members of that active company can read or write, with every cross-company attempt denied by the server.

## Acceptance:
- An authenticated user can create a company; creating a company makes that user an `owner` member of it.
- Company membership has at minimum the roles `owner`, `admin`, and `member`; an authenticated user can list the companies they belong to and select one of them as their active company.
- An authenticated user can create and read one tenant-scoped example resource (`note`: an `id` and a `body` string) while their active company is set to the company that resource belongs to.
- A request to read or write a `note` belonging to a company the requester is not a member of is rejected by the server (a Connect error, not merely hidden or filtered in the UI), regardless of which company the requester currently has active.
- The slice is implemented end to end — migration, `db/queries`, `proto/company/v1` and/or a `notes` extension of an existing contract, Go handler/service/repository, and a minimal Svelte 5 UI to create a company, create a note, and see the cross-company denial — with automated Go tests traceable to this PRD's `TC-022-n` test cases.

## Out of Scope
- Billing, custom domains, SSO, SCIM, or any production deployment or production data migration.
- More than the three membership roles, invitations/email flows, or company deletion/transfer.
- Any change to files outside what this slice requires (no unrelated refactors, no touching prior PRDs' delivered behavior).

## Test Cases
### TC-022-1: Creating a company makes the creator its owner
- Given an authenticated user with no companies
- When they call the create-company RPC with a name
- Then a company row is created and a `company_membership` row links that user to it with role `owner`

### TC-022-2: Membership roles are enforced and listable
- Given a company with an `owner`, an `admin`, and a `member` added via membership rows
- When the authenticated user lists their companies and memberships
- Then each membership's role is one of `owner`, `admin`, `member` and matches what was assigned

### TC-022-3: Active-company selection scopes subsequent requests
- Given a user who is a member of two companies, A and B
- When the user selects company A as their active company and then creates a `note`
- Then the created `note` is stored under company A, independent of company B's data

### TC-022-4: Tenant-scoped resource create/read succeeds only within the active, member company
- Given a `note` created under company A and a user who is a member of company A with A set active
- When the user requests that `note` by id
- Then the server returns it successfully

### TC-022-5: Cross-company read and write are denied at the server
- Given a `note` created under company A and a user who is not a member of company A (including a user whose active company is B)
- When that user requests to read or write the company-A `note` by id
- Then the server rejects the request with a Connect permission-denied/not-found error, and no company-A data is returned

## Benchmark pins (ALV-39)
This exact PRD file is run unmodified, in separate isolated worktrees/clones with separate databases, against two `template-v5` baselines:
- **Treatment (`latest`):** `4a229ed7127d83ba48f8a81df24047e508190dc5` (tip of `main` at the moment this PRD was drafted; `make doc-lint` passes clean at this SHA).
- **Control:** `7ba2f0605930e60e3237e3c798177ca79eed6c84`, per the ALV-7 "Feature Delivery Benchmark Thesis" issue document.

No implementation against either baseline starts until an owner approves this PRD via the linked `request_confirmation`, because the slice touches authentication/authorization (membership-based access control) and adds a schema migration.
