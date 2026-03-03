import { json, error } from '@sveltejs/kit';
import { getCuriosClient, call, grpcMessage } from '$lib/server/grpc/clients';
import type { RequestHandler } from './$types';

export const GET: RequestHandler = async ({ params }) => {
	try {
		const data = await call(getCuriosClient(), 'GetCurio', { id: params.id });
		return json(data);
	} catch (err) {
		throw error(404, grpcMessage(err));
	}
};

export const PUT: RequestHandler = async ({ params, request }) => {
	const body = await request.json();
	try {
		const data = await call(getCuriosClient(), 'UpdateCurio', {
			id: params.id,
			title: body.title,
			description: body.description ?? '',
			format_type: body.formatType ?? 'PHYSICAL',
			tags: body.tags ?? [],
			barcode: body.barcode ?? ''
		});
		return json(data);
	} catch (err) {
		throw error(400, grpcMessage(err));
	}
};

export const DELETE: RequestHandler = async ({ params }) => {
	try {
		await call(getCuriosClient(), 'DeleteCurio', { id: params.id });
		return new Response(null, { status: 204 });
	} catch (err) {
		throw error(400, grpcMessage(err));
	}
};
