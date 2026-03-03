import { redirect } from '@sveltejs/kit';
import { Google, generateCodeVerifier, generateState } from 'arctic';
import type { RequestHandler } from './$types';

function getGoogle() {
	const clientId = process.env.GOOGLE_CLIENT_ID ?? '';
	const clientSecret = process.env.GOOGLE_CLIENT_SECRET ?? '';
	const redirectUri =
		process.env.GOOGLE_REDIRECT_URI ?? 'http://localhost:3000/auth/callback/google';
	return new Google(clientId, clientSecret, redirectUri);
}

export const GET: RequestHandler = async ({ cookies, url, locals }) => {
	if (!process.env.GOOGLE_CLIENT_ID) {
		throw redirect(303, '/auth/login');
	}

	const isLink = url.searchParams.get('link') === 'true';
	// Link flow requires an active session
	if (isLink && !locals.user) {
		throw redirect(303, '/auth/login');
	}

	const state = generateState();
	const codeVerifier = generateCodeVerifier();
	const google = getGoogle();
	const authUrl = google.createAuthorizationURL(state, codeVerifier, ['openid', 'email', 'profile']);

	cookies.set('google_oauth_state', state, {
		path: '/',
		httpOnly: true,
		maxAge: 60 * 10,
		sameSite: 'lax'
	});
	cookies.set('google_code_verifier', codeVerifier, {
		path: '/',
		httpOnly: true,
		maxAge: 60 * 10,
		sameSite: 'lax'
	});

	if (isLink && locals.user) {
		cookies.set('google_link_user_id', locals.user.userId, {
			path: '/',
			httpOnly: true,
			maxAge: 60 * 10,
			sameSite: 'lax'
		});
	}

	throw redirect(302, authUrl.toString());
};
