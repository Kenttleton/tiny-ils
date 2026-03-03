import { json, error } from '@sveltejs/kit';
import { getCuriosClient, call, grpcMessage } from '$lib/server/grpc/clients';
import type { RequestHandler } from './$types';

export const POST: RequestHandler = async ({ request }) => {
	const body = await request.json();
	try {
		const data = await call(getCuriosClient(), 'EnrichMetadata', {
			media_type: body.mediaType,
			identifier: body.identifier
		});
		return json(data);
	} catch (err) {
		throw error(400, grpcMessage(err));
	}
};
