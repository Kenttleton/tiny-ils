import { json, error } from '@sveltejs/kit';
import { getUsersClient, call, grpcMessage } from '$lib/server/grpc/clients';
import type { RequestHandler } from './$types';

export const GET: RequestHandler = async ({ locals }) => {
	try {
		const data = await call(getUsersClient(), 'ListClaims', { node_id: locals.nodeId });
		return json(data);
	} catch (err) {
		throw error(500, grpcMessage(err));
	}
};
