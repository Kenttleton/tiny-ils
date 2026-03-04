import { fail, redirect } from '@sveltejs/kit';
import { getUsersClient, getNetworkClient, call, grpcMessage } from '$lib/server/grpc/clients';
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

	const nodeConfig = await call<{ grpc_address: string }>(
		getNetworkClient(),
		'GetNodeConfig',
		{}
	).catch(() => ({ grpc_address: '' }));

	return { detectedPublicUrl: detectedUrl, detectedGrpcAddress: nodeConfig.grpc_address ?? '' };
};

export const actions: Actions = {
	default: async ({ request, cookies, locals }) => {
		const data = await request.formData();
		const displayName = String(data.get('displayName') ?? '').trim();
		const email = String(data.get('email') ?? '').trim();
		const password = String(data.get('password') ?? '');
		const confirm = String(data.get('confirm') ?? '');
		const publicUrl = String(data.get('publicUrl') ?? '').trim().replace(/\/$/, '');
		const grpcAddress = String(data.get('grpcAddress') ?? '').trim();

		const values = { email, displayName, publicUrl, grpcAddress };

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
			setAuthCookie(cookies, res.token, locals.nodeId);

			if (grpcAddress) {
				await call(getNetworkClient(), 'SetNodeAddress', { address: grpcAddress }).catch(
					() => {} // non-fatal — admin can correct in settings
				);
			}
		} catch (err) {
			return fail(400, { error: grpcMessage(err), ...values });
		}

		throw redirect(303, '/admin');
	}
};
