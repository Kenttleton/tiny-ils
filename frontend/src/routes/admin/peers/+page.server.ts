import { fail } from '@sveltejs/kit';
import { getNetworkClient, call, grpcMessage } from '$lib/server/grpc/clients';
import type { PeerList } from '$lib/api';
import type { Actions, PageServerLoad } from './$types';

export const load: PageServerLoad = async () => {
	const [nodeInfo, nodeConfig, peerData] = await Promise.all([
		call<{ node_id: string; public_key: string; capabilities: string[] }>(
			getNetworkClient(),
			'GetNodeInfo',
			{}
		).catch(() => ({ node_id: '', public_key: '', capabilities: [] as string[] })),
		call<{ grpc_address: string }>(getNetworkClient(), 'GetNodeConfig', {}).catch(() => ({
			grpc_address: ''
		})),
		call<PeerList>(getNetworkClient(), 'ListPeers', {}).catch(() => ({ peers: [] } as PeerList))
	]);
	return {
		nodeId: nodeInfo.node_id,
		publicKey: nodeInfo.public_key,
		capabilities: nodeInfo.capabilities ?? [],
		grpcAddress: nodeConfig.grpc_address ?? '',
		peers: peerData.peers ?? []
	};
};

export const actions: Actions = {
	connect: async ({ request }) => {
		const form = await request.formData();
		const body = {
			node_id: form.get('nodeId')?.toString() ?? '',
			public_key: form.get('publicKey')?.toString() ?? '',
			address: form.get('address')?.toString() ?? '',
			display_name: form.get('displayName')?.toString() || ''
		};
		if (!body.node_id || !body.public_key || !body.address) {
			return fail(400, {
				error: 'Library ID, public key, and address are required',
				values: {
					nodeId: body.node_id,
					publicKey: body.public_key,
					address: body.address,
					displayName: body.display_name
				}
			});
		}
		try {
			await call(getNetworkClient(), 'ConnectPeer', body);
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
	},

	approve: async ({ request }) => {
		const form = await request.formData();
		const nodeId = form.get('nodeId')?.toString() ?? '';
		if (!nodeId) {
			return fail(400, { error: 'Library ID is required' });
		}
		try {
			await call(getNetworkClient(), 'ApprovePeer', { node_id: nodeId });
			return { success: true };
		} catch (err) {
			return fail(400, { error: grpcMessage(err) });
		}
	}
};
