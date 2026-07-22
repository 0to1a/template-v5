// The single place that constructs Connect clients. Components must call
// these clients, never build their own transport.
//
// The session travels as an HttpOnly cookie the browser attaches
// automatically (PRD 017) — this module never reads or sets it. The
// transport's fetch override makes `credentials: 'same-origin'` explicit
// because the frontend and API are one origin/one binary
// (docs/architecture.md); if that ever changes, this and CORS both need
// revisiting together.
import { createClient, type Interceptor, ConnectError, Code } from '@connectrpc/connect';
import { createConnectTransport } from '@connectrpc/connect-web';

import { AuthService } from './gen/auth/v1/auth_pb';
import { clearAuthenticated } from './auth';

const sessionExpiryInterceptor: Interceptor = (next) => async (req) => {
	try {
		return await next(req);
	} catch (err) {
		const connectError = ConnectError.from(err);
		if (connectError.code === Code.Unauthenticated) {
			clearAuthenticated();
		}
		throw connectError;
	}
};

const transport = createConnectTransport({
	baseUrl: '/',
	fetch: (input, init) => globalThis.fetch(input, { ...init, credentials: 'same-origin' }),
	interceptors: [sessionExpiryInterceptor]
});

export const authClient = createClient(AuthService, transport);
