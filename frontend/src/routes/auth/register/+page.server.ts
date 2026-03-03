import { redirect, fail } from '@sveltejs/kit';
import { getUsersClient, call, grpcMessage } from '$lib/server/grpc/clients';
import { setAuthCookie } from '$lib/server/auth';
import type { Actions, PageServerLoad } from './$types';

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

		try {
			const res = await call<{ token: string }>(getUsersClient(), 'Register', {
				email,
				password,
				display_name: displayName ?? ''
			});
			setAuthCookie(cookies, res.token);
			throw redirect(303, '/');
		} catch (err) {
			if ((err as { status?: number }).status === 303) throw err;
			return fail(400, { error: grpcMessage(err) || 'Registration failed', email, displayName });
		}
	}
};
