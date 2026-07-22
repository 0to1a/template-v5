// The single place that constructs Connect clients and attaches the bearer
// header. Components must call these clients, never build their own
// transport or read the access token directly.
import { createClient, type Interceptor, ConnectError, Code } from '@connectrpc/connect';
import { createConnectTransport } from '@connectrpc/connect-web';

import { AuthService } from './gen/auth/v1/auth_pb';
import { getAccessToken, clearAccessToken } from './auth';

const authInterceptor: Interceptor = (next) => async (req) => {
	const token = getAccessToken();
	if (token) {
		req.header.set('Authorization', `Bearer ${token}`);
	}

	try {
		return await next(req);
	} catch (err) {
		const connectError = ConnectError.from(err);
		if (connectError.code === Code.Unauthenticated) {
			clearAccessToken();
		}
		throw connectError;
	}
};

const transport = createConnectTransport({
	baseUrl: '/',
	interceptors: [authInterceptor]
});

export const authClient = createClient(AuthService, transport);
