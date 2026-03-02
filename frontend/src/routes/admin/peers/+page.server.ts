import { fail } from '@sveltejs/kit';
import { serverFetch } from '$lib/server/bff';
import type { PeerList } from '$lib/api';
import type { Actions, PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ request }) => {
	const cookie = request.headers.get('cookie') ?? '';
	const { data } = await serverFetch<PeerList>('GET', '/peers', cookie).catch(() => ({
		data: { peers: [] } as PeerList
	}));
	return { peers: data.peers ?? [] };
};

export const actions: Actions = {
	register: async ({ request }) => {
		const cookie = request.headers.get('cookie') ?? '';
		const form = await request.formData();
		const body = {
			nodeId: form.get('nodeId')?.toString() ?? '',
			publicKey: form.get('publicKey')?.toString() ?? '',
			address: form.get('address')?.toString() ?? '',
			displayName: form.get('displayName')?.toString() || undefined
		};
		if (!body.nodeId || !body.publicKey || !body.address) {
			return fail(400, { error: 'nodeId, publicKey, and address are required', values: body });
		}
		try {
			await serverFetch('POST', '/peers', cookie, body);
			return { success: true };
		} catch (err) {
			return fail(400, { error: String(err), values: body });
		}
	}
};
