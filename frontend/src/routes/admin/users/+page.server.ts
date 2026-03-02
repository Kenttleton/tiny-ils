import { fail } from '@sveltejs/kit';
import { serverFetch } from '$lib/server/bff';
import type { ClaimList } from '$lib/api';
import type { Actions, PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ request }) => {
	const cookie = request.headers.get('cookie') ?? '';
	const { data } = await serverFetch<ClaimList>('GET', '/claims', cookie).catch(() => ({
		data: { claims: [] } as ClaimList
	}));
	return { claims: data.claims ?? [] };
};

export const actions: Actions = {
	grant: async ({ request }) => {
		const cookie = request.headers.get('cookie') ?? '';
		const form = await request.formData();
		const userId = form.get('userId')?.toString() ?? '';
		const role = form.get('role')?.toString() ?? 'MANAGER';
		if (!userId) return fail(400, { error: 'User ID required' });
		try {
			await serverFetch('POST', '/claims/grant', cookie, { userId, role });
			return { success: true };
		} catch (err) {
			return fail(400, { error: String(err) });
		}
	},

	revoke: async ({ request }) => {
		const cookie = request.headers.get('cookie') ?? '';
		const form = await request.formData();
		const userId = form.get('userId')?.toString() ?? '';
		if (!userId) return fail(400, { error: 'User ID required' });
		try {
			await serverFetch('DELETE', '/claims/revoke', cookie, { userId });
			return { success: true };
		} catch (err) {
			return fail(400, { error: String(err) });
		}
	}
};
