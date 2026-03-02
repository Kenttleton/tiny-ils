import { error, redirect, fail } from '@sveltejs/kit';
import { serverFetch } from '$lib/server/bff';
import type { Curio } from '$lib/api';
import type { Actions, PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ request, params }) => {
	const cookie = request.headers.get('cookie') ?? '';
	try {
		const { data } = await serverFetch<Curio>('GET', `/curios/${params.id}`, cookie);
		return { curio: data };
	} catch {
		throw error(404, 'Curio not found');
	}
};

export const actions: Actions = {
	update: async ({ request, params }) => {
		const cookie = request.headers.get('cookie') ?? '';
		const form = await request.formData();
		const body = {
			title: form.get('title')?.toString() ?? '',
			description: form.get('description')?.toString() || undefined,
			mediaType: form.get('mediaType')?.toString() ?? 'THING',
			formatType: form.get('formatType')?.toString() ?? 'PHYSICAL',
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
			await serverFetch<Curio>('PUT', `/curios/${params.id}`, cookie, body);
			throw redirect(303, '/admin/curios');
		} catch (err) {
			if ((err as { status?: number }).status === 303) throw err;
			return fail(400, { error: String(err) });
		}
	}
};
