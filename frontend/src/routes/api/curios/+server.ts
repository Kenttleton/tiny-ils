import { json, error } from '@sveltejs/kit';
import { getCuriosClient, call, grpcMessage } from '$lib/server/grpc/clients';
import type { RequestHandler } from './$types';

export const GET: RequestHandler = async ({ url }) => {
	const q = url.searchParams.get('q') ?? '';
	const mediaType = url.searchParams.get('mediaType') ?? '';
	const limit = Number(url.searchParams.get('limit') ?? 50);
	const offset = Number(url.searchParams.get('offset') ?? 0);
	try {
		const data = await call(getCuriosClient(), 'ListCurios', {
			query: q,
			media_type: mediaType,
			limit,
			offset
		});
		return json(data);
	} catch (err) {
		throw error(500, grpcMessage(err));
	}
};

export const POST: RequestHandler = async ({ request }) => {
	const body = await request.json();
	try {
		const data = await call(getCuriosClient(), 'CreateCurio', {
			title: body.title,
			description: body.description ?? '',
			media_type: body.mediaType ?? 'THING',
			format_type: body.formatType ?? 'PHYSICAL',
			tags: body.tags ?? [],
			barcode: body.barcode ?? ''
		});
		return json(data, { status: 201 });
	} catch (err) {
		throw error(400, grpcMessage(err));
	}
};
