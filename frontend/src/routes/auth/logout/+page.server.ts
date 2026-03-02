import { redirect } from '@sveltejs/kit';
import type { Actions } from './$types';

const BFF = process.env.BFF_URL ?? 'http://localhost:3001';

export const actions: Actions = {
	default: async ({ request, cookies }) => {
		const cookie = request.headers.get('cookie') ?? '';
		await fetch(`${BFF}/auth/logout`, { method: 'POST', headers: { cookie } }).catch(() => null);
		// Clear the session cookie
		cookies.delete('connect.sid', { path: '/' });
		throw redirect(303, '/auth/login');
	}
};
