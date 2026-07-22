// The only module allowed to reason about authentication state. Components
// must go through here, never read cookies or a token directly.
//
// The session token itself never enters this module, or any other frontend
// code (PRD 017): it lives only in an HttpOnly cookie the server sets and
// the browser attaches automatically. What this module reads is a paired,
// non-secret "authed" indicator cookie the server sets/clears alongside it,
// purely so the UI can answer "am I logged in" without ever holding a
// credential.
const AUTHED_INDICATOR_COOKIE = 'template_v5_authed';

// In-memory fallback keeps this module usable where `document` does not
// exist (unit tests under Node). Browsers always take the cookie path.
let memoryAuthed = false;

function hasDocument(): boolean {
	return typeof document !== 'undefined';
}

export function isAuthenticated(): boolean {
	if (!hasDocument()) return memoryAuthed;
	return document.cookie.split('; ').includes(`${AUTHED_INDICATOR_COOKIE}=1`);
}

// markAuthenticated/clearAuthenticated let the login and logout flows (and
// tests) track state in environments without a real cookie jar. In a
// browser, the server's Set-Cookie on the same response already updates
// document.cookie before these run; they are the source of truth only for
// the in-memory fallback above.
export function markAuthenticated(): void {
	memoryAuthed = true;
}

export function clearAuthenticated(): void {
	memoryAuthed = false;
}
