import { redirect, fail } from '@sveltejs/kit';
import { getNetworkClient, call, grpcMessage } from '$lib/server/grpc/clients';
import { setAuthCookie } from '$lib/server/auth';
import type { Actions, PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ locals, url }) => {
	if (locals.user) throw redirect(303, url.searchParams.get('next') ?? '/');
	return { next: url.searchParams.get('next') ?? '' };
};

export const actions: Actions = {
	default: async ({ request, cookies, url, locals }) => {
		const data = await request.formData();
		const homeNodeAddress = data.get('home_node_address')?.toString().trim() ?? '';
		const homeNodeId = data.get('home_node_id')?.toString().trim() ?? '';
		const userId = data.get('user_id')?.toString().trim() ?? '';
		const next = data.get('next')?.toString() ?? url.searchParams.get('next') ?? '/';

		if (!homeNodeAddress || !homeNodeId || !userId) {
			return fail(400, { error: 'All fields are required.', homeNodeAddress, homeNodeId, userId });
		}

		try {
			const res = await call<{ token: string; display_name: string }>(
				getNetworkClient(),
				'AuthenticateGuest',
				{
					user_id: userId,
					home_node_address: homeNodeAddress,
					home_node_id: homeNodeId
				}
			);
			setAuthCookie(cookies, res.token, locals.nodeId);
			throw redirect(303, next);
		} catch (err) {
			if ((err as { status?: number }).status === 303) throw err;
			return fail(401, {
				error: grpcMessage(err) || 'Cross-node login failed. Check the address and ID.',
				homeNodeAddress,
				homeNodeId,
				userId
			});
		}
	}
};
