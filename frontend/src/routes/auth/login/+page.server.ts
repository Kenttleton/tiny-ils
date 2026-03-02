import { redirect, fail } from '@sveltejs/kit';
import { parseSetCookie } from '$lib/server/bff';
import type { Actions, PageServerLoad } from './$types';

const BFF = process.env.BFF_URL ?? 'http://localhost:3001';

export const load: PageServerLoad = async ({ locals, url }) => {
	if (locals.user) throw redirect(303, url.searchParams.get('next') ?? '/');
	return { next: url.searchParams.get('next') ?? '' };
};

export const actions: Actions = {
	default: async ({ request, cookies, url }) => {
		const data = await request.formData();
		const email = data.get('email')?.toString() ?? '';
		const password = data.get('password')?.toString() ?? '';
		const next = data.get('next')?.toString() ?? url.searchParams.get('next') ?? '/';

		const res = await fetch(`${BFF}/auth/login`, {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ email, password })
		});

		if (!res.ok) {
			const err = await res.json().catch(() => ({ error: 'Login failed' }));
			return fail(400, { error: err.error ?? 'Login failed', email });
		}

		const setCookie = res.headers.get('set-cookie');
		if (setCookie) {
			const { name, value, path, httpOnly, secure, sameSite, maxAge } = parseSetCookie(setCookie);
			cookies.set(name, value, { path, httpOnly, secure, sameSite, maxAge });
		}

		throw redirect(303, next);
	}
};
