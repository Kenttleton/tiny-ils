import { error } from '@sveltejs/kit';
import { getCuriosClient, call, grpcMessage } from '$lib/server/grpc/clients';
import type { RequestHandler } from './$types';

export const DELETE: RequestHandler = async ({ params }) => {
	try {
		await call(getCuriosClient(), 'CancelHold', { id: params.id });
		return new Response(null, { status: 204 });
	} catch (err) {
		throw error(400, grpcMessage(err));
	}
};
