import { describe, expect, it, beforeEach } from 'vitest';

import { clearAuthenticated, isAuthenticated } from './auth';
import {
	completeLogin,
	getPendingEmail,
	postLoginPath,
	setPendingEmail,
	type SubmitLoginClient
} from './login';

describe('login flow', () => {
	beforeEach(() => {
		clearAuthenticated();
		setPendingEmail('');
	});

	it('TC-017-2 records authenticated state through the central auth module (no token held) and redirects home', async () => {
		const client: SubmitLoginClient = {
			submitLogin: async (req) => {
				expect(req).toEqual({ email: 'admin@localhost', code: '123456' });
				return {} as Awaited<ReturnType<SubmitLoginClient['submitLogin']>>;
			}
		};
		const navigatedTo: string[] = [];
		const navigate = async (path: string) => {
			navigatedTo.push(path);
		};

		setPendingEmail('admin@localhost');
		await completeLogin('admin@localhost', '123456', navigate, client);

		expect(isAuthenticated()).toBe(true);
		expect(getPendingEmail()).toBe('');
		expect(navigatedTo).toEqual([postLoginPath]);
	});

	it('TC-001-4 stays unauthenticated and does not navigate when the code is rejected', async () => {
		const client: SubmitLoginClient = {
			submitLogin: async () => {
				throw new Error('unauthenticated');
			}
		};
		const navigatedTo: string[] = [];
		const navigate = async (path: string) => {
			navigatedTo.push(path);
		};

		await expect(completeLogin('admin@localhost', '000000', navigate, client)).rejects.toThrow();
		expect(isAuthenticated()).toBe(false);
		expect(navigatedTo).toEqual([]);
	});
});
