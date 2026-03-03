import { redirect, fail } from '@sveltejs/kit';
import { getCuriosClient, call, grpcMessage } from '$lib/server/grpc/clients';
import type { Curio, CurioMetadata } from '$lib/api';
import type { Actions, PageServerLoad } from './$types';

export const load: PageServerLoad = async () => {
	return {};
};

export const actions: Actions = {
	enrich: async ({ request }) => {
		const form = await request.formData();
		const mediaType = form.get('mediaType')?.toString() ?? '';
		const identifier = form.get('identifier')?.toString() ?? '';
		try {
			const data = await call<CurioMetadata>(getCuriosClient(), 'EnrichMetadata', {
				media_type: mediaType,
				identifier
			});
			return { enriched: data };
		} catch (err) {
			return fail(400, { error: grpcMessage(err) });
		}
	},

	create: async ({ request }) => {
		const form = await request.formData();
		const body = {
			title: form.get('title')?.toString() ?? '',
			description: form.get('description')?.toString() || undefined,
			media_type: form.get('mediaType')?.toString() ?? 'THING',
			format_type: form.get('formatType')?.toString() ?? 'PHYSICAL',
			tags: form
				.get('tags')
				?.toString()
				.split(',')
				.map((t) => t.trim())
				.filter(Boolean),
			barcode: form.get('barcode')?.toString() || undefined
		};
		if (!body.title) return fail(400, { error: 'Title is required', values: body });
		try {
			const data = await call<Curio>(getCuriosClient(), 'CreateCurio', body);
			throw redirect(303, `/admin/curios/${data.id}/edit`);
		} catch (err) {
			if ((err as { status?: number }).status === 303) throw err;
			return fail(400, { error: grpcMessage(err), values: body });
		}
	}
};
