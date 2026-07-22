import { describe, expect, it, beforeEach } from 'vitest';

import { isAuthenticated, markAuthenticated } from './auth';
import { performLogout, postLogoutPath, type LogoutClient } from './logout';

describe('logout flow', () => {
	beforeEach(() => {
		markAuthenticated();
	});

	it('TC-017-3 clears local authenticated state and navigates home', async () => {
		let called = false;
		const client: LogoutClient = {
			logout: async () => {
				called = true;
				return {} as Awaited<ReturnType<LogoutClient['logout']>>;
			}
		};
		const navigatedTo: string[] = [];
		const navigate = async (path: string) => {
			navigatedTo.push(path);
		};

		expect(isAuthenticated()).toBe(true);
		await performLogout(navigate, client);

		expect(called).toBe(true);
		expect(isAuthenticated()).toBe(false);
		expect(navigatedTo).toEqual([postLogoutPath]);
	});
});
