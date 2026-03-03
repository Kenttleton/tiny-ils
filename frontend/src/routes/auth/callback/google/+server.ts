import { redirect } from '@sveltejs/kit';
import { Google } from 'arctic';
import { getUsersClient, call, grpcMessage } from '$lib/server/grpc/clients';
import { setAuthCookie } from '$lib/server/auth';
import type { RequestHandler } from './$types';

function getGoogle() {
	const clientId = process.env.GOOGLE_CLIENT_ID ?? '';
	const clientSecret = process.env.GOOGLE_CLIENT_SECRET ?? '';
	const redirectUri =
		process.env.GOOGLE_REDIRECT_URI ?? 'http://localhost:3000/auth/callback/google';
	return new Google(clientId, clientSecret, redirectUri);
}

export const GET: RequestHandler = async ({ url, cookies }) => {
	const code = url.searchParams.get('code');
	const state = url.searchParams.get('state');
	const storedState = cookies.get('google_oauth_state');
	const codeVerifier = cookies.get('google_code_verifier');

	if (!code || !state || !storedState || !codeVerifier || state !== storedState) {
		throw redirect(303, '/auth/login');
	}

	cookies.delete('google_oauth_state', { path: '/' });
	cookies.delete('google_code_verifier', { path: '/' });

	try {
		const google = getGoogle();
		const tokens = await google.validateAuthorizationCode(code, codeVerifier);
		const accessToken = tokens.accessToken();

		const userinfo = await fetch('https://openidconnect.googleapis.com/v1/userinfo', {
			headers: { Authorization: `Bearer ${accessToken}` }
		}).then((r) => r.json() as Promise<{ sub: string; email: string; name: string }>);

		const res = await call<{ token: string }>(getUsersClient(), 'UpsertSSOUser', {
			provider: 'google',
			subject: userinfo.sub,
			email: userinfo.email,
			display_name: userinfo.name ?? userinfo.email
		});

		setAuthCookie(cookies, res.token);
		throw redirect(303, '/');
	} catch (err) {
		if ((err as { status?: number }).status === 303) throw err;
		console.error('Google OAuth callback error:', grpcMessage(err));
		throw redirect(303, '/auth/login');
	}
};
