import { json, error } from '@sveltejs/kit';
import { getNetworkClient, call, grpcMessage } from '$lib/server/grpc/clients';
import type { RequestHandler } from './$types';

export const GET: RequestHandler = async () => {
	try {
		const data = await call(getNetworkClient(), 'ListPeers', {});
		return json(data);
	} catch (err) {
		throw error(500, grpcMessage(err));
	}
};

export const POST: RequestHandler = async ({ request }) => {
	const body = await request.json();
	try {
		const data = await call(getNetworkClient(), 'RegisterPeer', {
			node_id: body.nodeId,
			public_key: body.publicKey,
			address: body.address,
			display_name: body.displayName ?? ''
		});
		return json(data, { status: 201 });
	} catch (err) {
		throw error(400, grpcMessage(err));
	}
};
