import { redirect, fail } from '@sveltejs/kit';
import { parseSetCookie } from '$lib/server/bff';
import type { Actions, PageServerLoad } from './$types';

const BFF = process.env.BFF_URL ?? 'http://localhost:3001';

export const load: PageServerLoad = async ({ locals }) => {
	if (locals.user) throw redirect(303, '/');
	return {};
};

export const actions: Actions = {
	default: async ({ request, cookies }) => {
		const data = await request.formData();
		const email = data.get('email')?.toString() ?? '';
		const password = data.get('password')?.toString() ?? '';
		const displayName = data.get('displayName')?.toString() || undefined;

		const res = await fetch(`${BFF}/auth/register`, {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ email, password, displayName })
		});

		if (!res.ok) {
			const err = await res.json().catch(() => ({ error: 'Registration failed' }));
			return fail(400, { error: err.error ?? 'Registration failed', email, displayName });
		}

		const setCookie = res.headers.get('set-cookie');
		if (setCookie) {
			const { name, value, path, httpOnly, secure, sameSite, maxAge } = parseSetCookie(setCookie);
			cookies.set(name, value, { path, httpOnly, secure, sameSite, maxAge });
		}

		throw redirect(303, '/');
	}
};
