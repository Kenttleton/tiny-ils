import { json, error } from '@sveltejs/kit';
import { getUsersClient, call, grpcMessage } from '$lib/server/grpc/clients';
import type { RequestHandler } from './$types';

export const POST: RequestHandler = async ({ request, locals }) => {
	const body = await request.json();
	try {
		await call(getUsersClient(), 'GrantClaim', {
			user_id: body.userId,
			node_id: body.nodeId ?? locals.nodeId,
			role: body.role
		});
		return json({ ok: true });
	} catch (err) {
		throw error(400, grpcMessage(err));
	}
};
