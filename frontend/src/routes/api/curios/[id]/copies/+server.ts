import { json, error } from '@sveltejs/kit';
import { getCuriosClient, call, grpcMessage } from '$lib/server/grpc/clients';
import type { RequestHandler } from './$types';

export const GET: RequestHandler = async ({ params }) => {
	try {
		const data = await call(getCuriosClient(), 'ListCopies', { id: params.id });
		return json(data);
	} catch (err) {
		throw error(500, grpcMessage(err));
	}
};
