import { fail } from '@sveltejs/kit';
import { getUsersClient, call, grpcMessage } from '$lib/server/grpc/clients';
import type { Actions, PageServerLoad } from './$types';

interface User {
	id: string;
	email: string;
	display_name: string;
	sso_provider: string;
	has_password: boolean;
}

export const load: PageServerLoad = async ({ locals, url }) => {
	const user = await call<User>(getUsersClient(), 'GetMe', { id: locals.user!.userId });
	return {
		profile: user,
		googleConfigured: !!process.env.GOOGLE_CLIENT_ID,
		linked: url.searchParams.get('linked') ?? null,
		linkError: url.searchParams.has('link_error')
	};
};

export const actions: Actions = {
	update: async ({ request, locals }) => {
		const form = await request.formData();
		const displayName = form.get('displayName')?.toString().trim() ?? '';
		const email = form.get('email')?.toString().trim() ?? '';
		const newPassword = form.get('newPassword')?.toString() ?? '';
		const currentPassword = form.get('currentPassword')?.toString() ?? '';

		if (!displayName) return fail(400, { error: 'Display name is required' });

		try {
			const updated = await call<User>(getUsersClient(), 'UpdateUser', {
				id: locals.user!.userId,
				display_name: displayName,
				email,
				new_password: newPassword,
				current_password: currentPassword
			});
			return { success: true, profile: updated };
		} catch (err) {
			return fail(400, { error: grpcMessage(err) });
		}
	},

	unlinkSso: async ({ locals }) => {
		try {
			const updated = await call<User>(getUsersClient(), 'UpdateUser', {
				id: locals.user!.userId,
				unlink_sso: true
			});
			return { success: true, unlinked: true, profile: updated };
		} catch (err) {
			return fail(400, { error: grpcMessage(err) });
		}
	}
};
