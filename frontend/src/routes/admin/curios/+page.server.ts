import { fail } from '@sveltejs/kit';
import { getCuriosClient, call, grpcMessage } from '$lib/server/grpc/clients';
import type { CurioList } from '$lib/api';
import type { Actions, PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ url }) => {
	const q = url.searchParams.get('q') ?? '';
	const data = await call<CurioList>(getCuriosClient(), 'ListCurios', {
		query: q,
		media_type: '',
		limit: 100,
		offset: 0
	});
	return { curios: data.curios ?? [], total: data.total ?? 0, q };
};

export const actions: Actions = {
	delete: async ({ request }) => {
		const form = await request.formData();
		const id = form.get('id')?.toString();
		if (!id) return fail(400, { error: 'Missing id' });
		try {
			await call(getCuriosClient(), 'DeleteCurio', { id });
			return { success: true };
		} catch (err) {
			return fail(400, { error: grpcMessage(err) });
		}
	}
};
