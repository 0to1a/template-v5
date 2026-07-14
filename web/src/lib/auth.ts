// The only module allowed to read or write the bearer token. Components and
// API code must go through here, never touch localStorage directly, and
// never construct an Authorization header themselves.
//
// The token lives in localStorage: an XSS that runs in this origin can read
// it. That risk is accepted and documented for the initial version; moving
// to HttpOnly cookies is its own PRD, not a silent change.
const ACCESS_TOKEN_KEY = 'template-v5.access_token';

// In-memory fallback keeps this module usable where localStorage does not
// exist (unit tests under Node). Browsers always take the localStorage path.
let memoryToken: string | null = null;

function hasLocalStorage(): boolean {
	return typeof localStorage !== 'undefined';
}

export function getAccessToken(): string | null {
	if (!hasLocalStorage()) return memoryToken;
	return localStorage.getItem(ACCESS_TOKEN_KEY);
}

export function setAccessToken(token: string): void {
	if (!hasLocalStorage()) {
		memoryToken = token;
		return;
	}
	localStorage.setItem(ACCESS_TOKEN_KEY, token);
}

export function clearAccessToken(): void {
	if (!hasLocalStorage()) {
		memoryToken = null;
		return;
	}
	localStorage.removeItem(ACCESS_TOKEN_KEY);
}

export function isAuthenticated(): boolean {
	return getAccessToken() !== null;
}
