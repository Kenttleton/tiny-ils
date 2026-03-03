import { json, error } from '@sveltejs/kit';
import { getCuriosClient, call, grpcMessage } from '$lib/server/grpc/clients';
import type { RequestHandler } from './$types';

export const POST: RequestHandler = async ({ params, locals }) => {
	try {
		const data = await call(getCuriosClient(), 'PlaceHold', {
			curio_id: params.id,
			user_id: locals.user?.userId ?? '',
			user_node_id: locals.nodeId
		});
		return json(data);
	} catch (err) {
		throw error(400, grpcMessage(err));
	}
};
