import { fail } from '@sveltejs/kit';
import { getNetworkClient, call, grpcMessage } from '$lib/server/grpc/clients';
import type { PeerList } from '$lib/api';
import type { Actions, PageServerLoad } from './$types';

export const load: PageServerLoad = async () => {
	const data = await call<PeerList>(getNetworkClient(), 'ListPeers', {}).catch(() => ({
		peers: []
	} as PeerList));
	return { peers: data.peers ?? [] };
};

export const actions: Actions = {
	register: async ({ request }) => {
		const form = await request.formData();
		const body = {
			node_id: form.get('nodeId')?.toString() ?? '',
			public_key: form.get('publicKey')?.toString() ?? '',
			address: form.get('address')?.toString() ?? '',
			display_name: form.get('displayName')?.toString() || ''
		};
		if (!body.node_id || !body.public_key || !body.address) {
			return fail(400, {
				error: 'nodeId, publicKey, and address are required',
				values: {
					nodeId: body.node_id,
					publicKey: body.public_key,
					address: body.address,
					displayName: body.display_name
				}
			});
		}
		try {
			await call(getNetworkClient(), 'RegisterPeer', body);
			return { success: true };
		} catch (err) {
			return fail(400, {
				error: grpcMessage(err),
				values: {
					nodeId: body.node_id,
					publicKey: body.public_key,
					address: body.address,
					displayName: body.display_name
				}
			});
		}
	}
};
