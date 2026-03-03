import { json, error } from '@sveltejs/kit';
import { getCuriosClient, call, grpcMessage } from '$lib/server/grpc/clients';
import type { RequestHandler } from './$types';

export const GET: RequestHandler = async ({ url }) => {
	const q = url.searchParams.get('q') ?? '';
	const mediaType = url.searchParams.get('mediaType') ?? '';
	try {
		const data = await call(getCuriosClient(), 'ListCurios', {
			query: q,
			media_type: mediaType,
			limit: 50,
			offset: 0
		});
		return json(data);
	} catch (err) {
		throw error(500, grpcMessage(err));
	}
};
