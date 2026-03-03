import { json, error } from '@sveltejs/kit';
import { getCuriosClient, call, grpcMessage } from '$lib/server/grpc/clients';
import type { RequestHandler } from './$types';

const ACTION_MAP: Record<string, string> = {
	approve: 'ApproveTransfer',
	reject: 'RejectTransfer',
	ship: 'MarkShipped',
	receive: 'ConfirmReceived',
	cancel: 'CancelTransfer'
};

export const POST: RequestHandler = async ({ params, request, locals }) => {
	const rpc = ACTION_MAP[params.action];
	if (!rpc) throw error(404, `Unknown transfer action: ${params.action}`);

	const body = await request.json().catch(() => ({}));
	try {
		const data = await call(getCuriosClient(), rpc, {
			transfer_id: params.id,
			actor_id: locals.user?.userId ?? '',
			notes: body.notes ?? ''
		});
		return json(data);
	} catch (err) {
		throw error(400, grpcMessage(err));
	}
};
