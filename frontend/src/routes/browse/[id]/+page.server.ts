import { error, fail } from '@sveltejs/kit';
import { serverFetch } from '$lib/server/bff';
import type { Curio, CopyList, Hold } from '$lib/api';
import type { Actions, PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ request, params }) => {
	const cookie = request.headers.get('cookie') ?? '';
	const [curioRes, copiesRes] = await Promise.allSettled([
		serverFetch<Curio>('GET', `/curios/${params.id}`, cookie),
		serverFetch<CopyList>('GET', `/curios/${params.id}/copies`, cookie)
	]);

	if (curioRes.status === 'rejected') throw error(404, 'Curio not found');

	return {
		curio: curioRes.value.data,
		copies: copiesRes.status === 'fulfilled' ? (copiesRes.value.data.copies ?? []) : []
	};
};

export const actions: Actions = {
	checkout: async ({ request, params }) => {
		const cookie = request.headers.get('cookie') ?? '';
		const form = await request.formData();
		const copyId = form.get('copyId')?.toString();
		if (!copyId) return fail(400, { error: 'Missing copy ID' });
		try {
			await serverFetch('POST', `/copies/${copyId}/checkout`, cookie, {});
			return { success: true, action: 'checkout' };
		} catch (err) {
			return fail(400, { error: String(err) });
		}
	},

	hold: async ({ request, params }) => {
		const cookie = request.headers.get('cookie') ?? '';
		try {
			await serverFetch<Hold>('POST', `/curios/${params.id}/hold`, cookie, {});
			return { success: true, action: 'hold' };
		} catch (err) {
			return fail(400, { error: String(err) });
		}
	}
};
