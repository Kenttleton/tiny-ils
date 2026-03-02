import { fail } from '@sveltejs/kit';
import { serverFetch } from '$lib/server/bff';
import type { CurioList } from '$lib/api';
import type { Actions, PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ request, url }) => {
	const cookie = request.headers.get('cookie') ?? '';
	const q = url.searchParams.get('q') ?? '';
	const params = new URLSearchParams({ q, limit: '100' });
	const { data } = await serverFetch<CurioList>('GET', `/curios?${params}`, cookie);
	return { curios: data.curios ?? [], total: data.total ?? 0, q };
};

export const actions: Actions = {
	delete: async ({ request }) => {
		const cookie = request.headers.get('cookie') ?? '';
		const form = await request.formData();
		const id = form.get('id')?.toString();
		if (!id) return fail(400, { error: 'Missing id' });
		try {
			await serverFetch('DELETE', `/curios/${id}`, cookie);
			return { success: true };
		} catch (err) {
			return fail(400, { error: String(err) });
		}
	}
};
