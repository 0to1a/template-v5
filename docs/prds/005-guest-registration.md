# Guest self-registration on login request

## Purpose
Let an operator-controlled flag turn an unrecognized email into a new active account during login request, so a first-time user can get in without a separate signup step.

## Acceptance:
- Given `IS_GUEST_REGISTRATION=1` and an email with no active user, when `RequestLogin` is called, then a new active user is created with that email and a login code is delivered to it, using the same generic response shape as today.
- Given `IS_GUEST_REGISTRATION` is unset or `0`, when `RequestLogin` is called with an unrecognized email, then behavior is unchanged from today: no user is created, no code is sent, and the same generic response is returned.
- Given `IS_GUEST_REGISTRATION=1` and an email that already belongs to an active user, when `RequestLogin` is called, then no duplicate account is created and the existing user receives the login code as today.
- Given a user was auto-registered by this flag, when they submit the delivered code to `SubmitLogin`, then they receive a valid JWT exactly as any other active user would.

## Out of Scope
- Any new registration UI, page, or fields (name, password, etc.) — the login form and its email-only input are unchanged.
- Email format validation beyond the existing normalization (trim/lowercase).
- Rate limiting or abuse prevention for repeated guest registration.

## Test Cases
### TC-005-1: Guest registration creates a user when the flag is enabled
- Given `IS_GUEST_REGISTRATION=1` and no user exists for `new@example.com`
- When `RequestLogin` is called with `new@example.com`
- Then an active user now exists for that email and a login code was sent to it

### TC-005-2: Disabled flag preserves current no-registration behavior
- Given `IS_GUEST_REGISTRATION` is unset (or `0`) and no user exists for `new@example.com`
- When `RequestLogin` is called with `new@example.com`
- Then no user is created and no login code is sent, matching today's behavior

### TC-005-3: Enabled flag does not duplicate an existing account
- Given `IS_GUEST_REGISTRATION=1` and an active user already exists for `admin@localhost`
- When `RequestLogin` is called with `admin@localhost`
- Then exactly one user still exists for that email and the existing account receives the login code

### TC-005-4: An auto-registered guest can complete login
- Given `IS_GUEST_REGISTRATION=1` caused a new user to be created for `new@example.com` and a code was issued
- When that code is submitted to `SubmitLogin`
- Then a valid JWT is returned for that user
