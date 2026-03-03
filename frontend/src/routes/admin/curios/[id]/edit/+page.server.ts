import { error, redirect, fail } from '@sveltejs/kit';
import { getCuriosClient, call, grpcMessage } from '$lib/server/grpc/clients';
import type { Curio } from '$lib/api';
import type { Actions, PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ params }) => {
	try {
		const data = await call<Curio>(getCuriosClient(), 'GetCurio', { id: params.id });
		return { curio: data };
	} catch {
		throw error(404, 'Curio not found');
	}
};

export const actions: Actions = {
	update: async ({ request, params }) => {
		const form = await request.formData();
		const body = {
			id: params.id,
			title: form.get('title')?.toString() ?? '',
			description: form.get('description')?.toString() || undefined,
			format_type: form.get('formatType')?.toString() ?? 'PHYSICAL',
			tags: form
				.get('tags')
				?.toString()
				.split(',')
				.map((t) => t.trim())
				.filter(Boolean),
			barcode: form.get('barcode')?.toString() || undefined
		};
		if (!body.title) return fail(400, { error: 'Title is required' });
		try {
			await call<Curio>(getCuriosClient(), 'UpdateCurio', body);
			throw redirect(303, '/admin/curios');
		} catch (err) {
			if ((err as { status?: number }).status === 303) throw err;
			return fail(400, { error: grpcMessage(err) });
		}
	}
};
