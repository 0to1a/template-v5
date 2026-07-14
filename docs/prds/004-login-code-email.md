# Login code email delivery

## Purpose
Deliver login codes to users as HTML email over SMTP instead of discarding them, so a user requesting a login code actually receives it.

## Acceptance:
- Given `MAIL_URL` is set to a valid `smtp://user:password@host:port` value, when a user requests a login code, an HTML email rendered from an embedded template is sent to the user's address through that SMTP server, containing the code.
- Given `MAIL_URL` is unset, requesting a login code behaves exactly as it does today: no SMTP send is attempted and the request still succeeds.
- Given `MAIL_URL` is set but malformed (missing scheme or host), the server fails to start with a clear error rather than failing silently on the first login request.
- Given the SMTP send fails (e.g. server unreachable), the failure is logged without exposing the login code or SMTP credentials, and `RequestLogin` still returns success to the caller.

## Out of Scope
- Retry/backoff, queuing, or async delivery for failed sends.
- Any email type other than the login-code email (no other templates, no attachments, no inline images).
- New third-party dependencies — SMTP delivery uses the standard library only.

## Test Cases
### TC-004-1: Valid MAIL_URL sends the login code as HTML email
- Given `MAIL_URL` is set to a valid SMTP URL and a fake SMTP server is listening on it
- When a login code is requested for an existing user
- Then the fake server receives a message addressed to that user's email whose body is HTML and contains the code

### TC-004-2: Unset MAIL_URL preserves current no-op behavior
- Given `MAIL_URL` is not set in the environment
- When the server loads configuration and wires login code delivery
- Then no SMTP sender is constructed and requesting a login code succeeds with no SMTP attempt

### TC-004-3: Malformed MAIL_URL fails startup
- Given `MAIL_URL` is set to a value missing a scheme or host
- When the server constructs the SMTP sender from it
- Then construction returns an error instead of a usable sender

### TC-004-4: Send failure is logged without leaking the code
- Given `MAIL_URL` points at an SMTP server that rejects or is unreachable
- When a login code is requested
- Then the failure is logged, the log output does not contain the login code, and `RequestLogin` still returns no error
