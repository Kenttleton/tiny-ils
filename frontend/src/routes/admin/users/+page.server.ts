import { fail } from '@sveltejs/kit';
import { getUsersClient, call, grpcMessage } from '$lib/server/grpc/clients';
import type { Actions, PageServerLoad } from './$types';

interface User {
	id: string;
	email: string;
	display_name: string;
	sso_provider: string;
	created_at: number;
	has_password: boolean;
	role: string;
}

export const load: PageServerLoad = async ({ url }) => {
	const page = Math.max(1, parseInt(url.searchParams.get('page') ?? '1', 10));
	const limit = 50;
	const offset = (page - 1) * limit;

	const data = await call<{ users: User[]; total: number }>(getUsersClient(), 'ListUsers', {
		limit,
		offset
	}).catch(() => ({ users: [], total: 0 }));

	return { users: data.users ?? [], total: data.total ?? 0, page, limit };
};

export const actions: Actions = {
	create: async ({ request }) => {
		const form = await request.formData();
		const email = form.get('email')?.toString().trim() ?? '';
		const displayName = form.get('displayName')?.toString().trim() ?? '';
		const password = form.get('password')?.toString() ?? '';

		if (!email || !password) return fail(400, { error: 'Email and password are required' });
		if (password.length < 8) return fail(400, { error: 'Password must be at least 8 characters' });

		try {
			await call(getUsersClient(), 'Register', {
				email,
				display_name: displayName || email,
				password
			});
			return { success: true, action: 'created' };
		} catch (err) {
			return fail(400, { error: grpcMessage(err) });
		}
	},

	promote: async ({ request, locals }) => {
		const form = await request.formData();
		const userId = form.get('userId')?.toString() ?? '';
		if (!userId) return fail(400, { error: 'User ID required' });
		try {
			await call(getUsersClient(), 'GrantClaim', {
				user_id: userId,
				node_id: locals.nodeId,
				role: 'MANAGER'
			});
			return { success: true, action: 'promoted' };
		} catch (err) {
			return fail(400, { error: grpcMessage(err) });
		}
	},

	demote: async ({ request, locals }) => {
		const form = await request.formData();
		const userId = form.get('userId')?.toString() ?? '';
		if (!userId) return fail(400, { error: 'User ID required' });
		try {
			await call(getUsersClient(), 'GrantClaim', {
				user_id: userId,
				node_id: locals.nodeId,
				role: 'USER'
			});
			return { success: true, action: 'demoted' };
		} catch (err) {
			return fail(400, { error: grpcMessage(err) });
		}
	},

	delete: async ({ request }) => {
		const form = await request.formData();
		const userId = form.get('userId')?.toString() ?? '';
		if (!userId) return fail(400, { error: 'User ID required' });
		try {
			await call(getUsersClient(), 'DeleteUser', { id: userId });
			return { success: true, action: 'deleted' };
		} catch (err) {
			return fail(400, { error: grpcMessage(err) });
		}
	}
};
