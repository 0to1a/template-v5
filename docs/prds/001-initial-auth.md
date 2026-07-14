# Initial Email OTP Login

## Purpose
Allow active registered users to sign in to the application using their email address and a five-minute OTP.

## Acceptance:
- `GET /health` returns a healthy status without authentication and its handler performs no database query; the server still requires the database during startup.
- `AuthService/RequestLogin` accepts an email address and always returns a generic response without revealing whether the user is registered.
- `AuthService/SubmitLogin` issues a valid HS256 JWT for an active user with the correct OTP, and rejects invalid login attempts (an unknown user or incorrect OTP) with the same generic error.
- A user can enter an email address at `/login`, enter an OTP at `/login/otp`, and return to `/` after a successful login; the token is stored only through the central auth module.
- The development account `admin@localhost` can sign in using OTP `123456`.

## Out of Scope
- No self-registration, password login, refresh tokens, OAuth, roles, companies, or invitations.
- No real email provider integration or test-only endpoint for reading an OTP.
- No Playwright/browser E2E tests, broad UI redesign, or new frontend dependencies beyond the chosen foundation.

## Test Cases
### TC-001-1: Health handler is available without authentication or database queries
- Given the server has started successfully and the health handler has no database dependency
- When the client calls `GET /health`
- Then the handler returns a healthy response without authentication or a database query

### TC-001-2: Requesting a login does not reveal account existence
- Given one registered email address and one unregistered email address
- When each email address is sent to `AuthService/RequestLogin`
- Then both requests receive the same generic response shape

### TC-001-3: The local administrator signs in with the development OTP
- Given the active user `admin@localhost` exists
- When the user submits OTP `123456` to `AuthService/SubmitLogin`
- Then the server returns a valid HS256 JWT whose subject is the user's public identity

### TC-001-4: An incorrect OTP is rejected with a generic error
- Given an active user exists
- When the user submits an incorrect OTP to `AuthService/SubmitLogin`
- Then the server returns a generic authentication error and does not issue a token

### TC-001-5: The frontend login flow stores the token through the central auth module
- Given the user entered an email address at `/login` and is now at `/login/otp`
- When the user submits a valid OTP
- Then the token is stored through the central auth module and the user is redirected to `/`
