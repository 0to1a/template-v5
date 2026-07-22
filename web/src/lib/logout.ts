// Logout flow: asks the server to clear the session cookie, then records
// the signed-out state locally, then navigates. Mirrors login.ts's
// completeLogin structure.
import { authClient } from './client';
import { clearAuthenticated } from './auth';

// The slice of the generated client that logging out needs. Tests
// substitute a fake; pages use the default real client.
export type LogoutClient = Pick<typeof authClient, 'logout'>;

// Where a completed logout lands the user.
export const postLogoutPath = '/';

export async function performLogout(
	navigate: (path: string) => Promise<void>,
	client: LogoutClient = authClient
): Promise<void> {
	await client.logout({});
	clearAuthenticated();
	await navigate(postLogoutPath);
}
