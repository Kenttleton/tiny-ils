import { json, error } from '@sveltejs/kit';
import { getCuriosClient, call, grpcMessage } from '$lib/server/grpc/clients';
import type { RequestHandler } from './$types';

export const POST: RequestHandler = async ({ params }) => {
	try {
		const data = await call(getCuriosClient(), 'ReturnCopy', { copy_id: params.id });
		return json(data);
	} catch (err) {
		throw error(400, grpcMessage(err));
	}
};
