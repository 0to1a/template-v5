---
type: Product requirement
title: Pluggable email-provider abstraction beyond SMTP
description: Let a product built from this template deliver login-code email through a transactional API provider, not only direct SMTP, without changing internal/auth call sites.
tags: [auth, delivery, sensitive]
problem_brief: waiver
waiver_owner: Founding Engineer
waiver_reason: Internal delivery-infrastructure hardening for template-v5 itself, directed by the approved ALV-7 hardening plan (Fase 3 / Wave C); not a customer-facing capability requiring product-market evidence.
waiver_expires: 2026-08-22
---

# Pluggable email-provider abstraction beyond SMTP

## Owner decision: deferred (2026-07-22)
Via the `ask_user_questions` interaction "Open questions for sensitive
PRDs 015-018 (Wave C)" on ALV-14, the owner chose **"Defer — keep SMTP
only for now"** for the provider-selection question below. No provider was
named, so per this PRD's own gate ("do not add any dependency... until the
owner explicitly approves this PRD and names the specific provider") and
`AGENTS.md`'s vendor-approval boundary, none of the Acceptance criteria are
implemented and no dependency has been added. This PRD stays in the
backlog, unimplemented, until a future request names a provider. The
corresponding entry in `docs/threat-model.md`'s "currently-accepted risks"
remains — this decision does not close that risk, it accepts it for now.

## Owner approval required before implementation
This PRD affects the delivery path for login codes, called out explicitly
alongside the other sensitive auth items in the approved ALV-7 plan (Fase 3).
Per `AGENTS.md`, "Sensitive behavior (auth...) requires owner approval
before implementation" and "Do not add dependencies... unless the PRD
requires and approves them" — this PRD would very likely add one (a
provider SDK or HTTP client for a transactional email API). **Do not
implement any of the Acceptance below, and do not add any dependency, until
the owner explicitly approves this PRD and names the specific provider.**

## Open questions for the owner
- Which specific provider(s) to support first (e.g. Postmark, SES, Resend)
  — this determines the dependency this PRD would add.
  Vendor selection is a founder decision, not an engineering default
  (`AGENTS.md`: "Do not independently decide... commit to major vendors").
  Vendor selection is also a "major vendor" commitment per `AGENTS.md`'s
  permissions boundary.
- Whether `MAIL_URL`/SMTP stays as one supported option alongside the new
  provider, or is replaced.
- Failure/retry/fallback behavior across providers, if more than one is
  ever configured at once.

## Purpose
Let a product built from this template send login-code email through a
managed transactional-email API instead of only direct SMTP
(`internal/mail`), for the operational reliability and deliverability a raw
SMTP connection doesn't provide.

## Acceptance:
- `internal/auth.LoginCodeSender` (the existing interface) is satisfied by
  at least one new implementation that talks to the owner-selected provider,
  with no change required to `internal/auth/service.go`'s call sites.
- Provider credentials are read from configuration the same way `MAIL_URL`
  is today — never logged (see `docs/environment-contract.md`).
- Which sender implementation is active is chosen by configuration, not a
  code change, mirroring today's `MAIL_URL`-set-or-not switch in
  `cmd/server/main.go`.
- Existing SMTP delivery (`internal/mail`) continues to work unless the
  owner's approval explicitly says to replace it.

## Out of Scope
- Building a generic multi-provider failover/routing system — one new
  provider implementation is in scope; a provider-agnostic router is not,
  unless the owner's approval says otherwise.
- Any change to the login-code generation or verification logic in
  `internal/auth/otp.go` — this PRD is delivery only.
- The other three sensitive PRDs (demo-credential guardrail, login
  throttling/lockout, OTP replay prevention).

## Test Cases
### TC-018-1: The new provider implementation satisfies LoginCodeSender
- Given the owner-approved provider's sender implementation
- When it is used in place of `internal/mail.SMTPSender`
- Then `internal/auth.Service` requires no code change to send through it

### TC-018-2: Provider credentials are never logged
- Given a startup or delivery failure with the new provider configured
- When the resulting log output is inspected
- Then no credential or API key value appears in it

### TC-018-3: SMTP delivery still works unless explicitly replaced
- Given `MAIL_URL` is set and the new provider is not configured
- When a login code is requested
- Then it is still delivered via SMTP exactly as today
