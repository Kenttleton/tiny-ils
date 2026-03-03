import { fail, redirect } from '@sveltejs/kit';
import { getUsersClient, call, grpcMessage } from '$lib/server/grpc/clients';
import { setAuthCookie } from '$lib/server/auth';
import type { Actions, PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ request }) => {
	// Auto-detect the public URL from the incoming request so operators behind a
	// reverse proxy see the correct origin pre-filled in the setup form.
	const proto = request.headers.get('x-forwarded-proto') ?? 'http';
	const host =
		request.headers.get('x-forwarded-host') ??
		request.headers.get('host') ??
		'localhost:3000';
	const detectedUrl = `${proto}://${host}`;
	return { detectedPublicUrl: detectedUrl };
};

export const actions: Actions = {
	default: async ({ request, cookies }) => {
		const data = await request.formData();
		const displayName = String(data.get('displayName') ?? '').trim();
		const email = String(data.get('email') ?? '').trim();
		const password = String(data.get('password') ?? '');
		const confirm = String(data.get('confirm') ?? '');
		const publicUrl = String(data.get('publicUrl') ?? '').trim().replace(/\/$/, '');

		const values = { email, displayName, publicUrl };

		if (!email || !password) {
			return fail(400, { error: 'Email and password are required.', ...values });
		}
		if (password !== confirm) {
			return fail(400, { error: 'Passwords do not match.', ...values });
		}
		if (password.length < 8) {
			return fail(400, { error: 'Password must be at least 8 characters.', ...values });
		}
		if (!publicUrl) {
			return fail(400, { error: 'Public URL is required.', ...values });
		}

		try {
			const res = await call<{ token: string }>(getUsersClient(), 'BootstrapManager', {
				email,
				password,
				display_name: displayName || email,
				public_url: publicUrl
			});
			setAuthCookie(cookies, res.token);
		} catch (err) {
			return fail(400, { error: grpcMessage(err), ...values });
		}

		throw redirect(303, '/admin');
	}
};
