// Login flow module: carries the email between /login and /login/otp and
// completes the login. The token is stored only through the central auth
// module — pages never receive or handle it directly.
import { authClient } from './client';
import { setAccessToken } from './auth';

let pendingEmail = '';

export function setPendingEmail(email: string): void {
	pendingEmail = email;
}

export function getPendingEmail(): string {
	return pendingEmail;
}

export function clearPendingEmail(): void {
	pendingEmail = '';
}

// The slice of the generated client that completing a login needs. Tests
// substitute a fake; pages use the default real client.
export type SubmitLoginClient = Pick<typeof authClient, 'submitLogin'>;

// Where a successful login lands the user.
export const postLoginPath = '/';

// completeLogin exchanges email + code for a bearer token, stores it via the
// central auth module, then navigates home. Navigation is injected so pages
// pass SvelteKit's goto and tests pass a spy. Errors (invalid code, network)
// propagate to the caller for display; navigate is only reached on success.
export async function completeLogin(
	email: string,
	code: string,
	navigate: (path: string) => Promise<void>,
	client: SubmitLoginClient = authClient
): Promise<void> {
	const response = await client.submitLogin({ email, code });
	setAccessToken(response.accessToken);
	clearPendingEmail();
	await navigate(postLoginPath);
}
