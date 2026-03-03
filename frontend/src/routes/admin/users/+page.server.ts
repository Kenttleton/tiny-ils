import { fail } from '@sveltejs/kit';
import { getUsersClient, call, grpcMessage } from '$lib/server/grpc/clients';
import type { ClaimList } from '$lib/api';
import type { Actions, PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ locals }) => {
	const data = await call<ClaimList>(getUsersClient(), 'ListClaims', {
		node_id: locals.nodeId
	}).catch(() => ({ claims: [] } as ClaimList));
	return { claims: data.claims ?? [] };
};

export const actions: Actions = {
	grant: async ({ request, locals }) => {
		const form = await request.formData();
		const userId = form.get('userId')?.toString() ?? '';
		const role = form.get('role')?.toString() ?? 'MANAGER';
		if (!userId) return fail(400, { error: 'User ID required' });
		try {
			await call(getUsersClient(), 'GrantClaim', { user_id: userId, node_id: locals.nodeId, role });
			return { success: true };
		} catch (err) {
			return fail(400, { error: grpcMessage(err) });
		}
	},

	revoke: async ({ request, locals }) => {
		const form = await request.formData();
		const userId = form.get('userId')?.toString() ?? '';
		if (!userId) return fail(400, { error: 'User ID required' });
		try {
			await call(getUsersClient(), 'RevokeClaim', { user_id: userId, node_id: locals.nodeId });
			return { success: true };
		} catch (err) {
			return fail(400, { error: grpcMessage(err) });
		}
	}
};
