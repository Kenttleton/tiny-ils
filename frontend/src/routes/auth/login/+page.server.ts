import { redirect, fail } from '@sveltejs/kit';
import { getUsersClient, call, grpcMessage } from '$lib/server/grpc/clients';
import { setAuthCookie } from '$lib/server/auth';
import type { Actions, PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ locals, url }) => {
	if (locals.user) throw redirect(303, url.searchParams.get('next') ?? '/');
	return { next: url.searchParams.get('next') ?? '' };
};

export const actions: Actions = {
	default: async ({ request, cookies, url, locals }) => {
		const data = await request.formData();
		const email = data.get('email')?.toString() ?? '';
		const password = data.get('password')?.toString() ?? '';
		const next = data.get('next')?.toString() || url.searchParams.get('next') || '/';

		try {
			const res = await call<{ token: string }>(getUsersClient(), 'Login', { email, password });
			setAuthCookie(cookies, res.token, locals.nodeId);
			throw redirect(303, next);
		} catch (err) {
			if ((err as { status?: number }).status === 303) throw err;
			return fail(401, { error: grpcMessage(err) || 'Invalid email or password', email });
		}
	}
};
