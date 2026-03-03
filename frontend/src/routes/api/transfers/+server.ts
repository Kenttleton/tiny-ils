import { json, error } from '@sveltejs/kit';
import { getCuriosClient, getNetworkClient, call, grpcMessage } from '$lib/server/grpc/clients';
import type { RequestHandler } from './$types';

export const GET: RequestHandler = async ({ url, locals }) => {
	const status = url.searchParams.get('status') ?? '';
	const nodeId = url.searchParams.get('nodeId') ?? locals.nodeId;
	const transferType = url.searchParams.get('transferType') ?? '';
	try {
		const data = await call(getCuriosClient(), 'ListTransfers', {
			status,
			node_id: nodeId,
			transfer_type: transferType
		});
		return json(data);
	} catch (err) {
		throw error(500, grpcMessage(err));
	}
};

export const POST: RequestHandler = async ({ request, locals }) => {
	const body = await request.json();
	const nodeId = locals.nodeId;
	const sourceNode: string = body.sourceNode || nodeId;

	try {
		if (sourceNode === nodeId || !sourceNode) {
			const data = await call(getCuriosClient(), 'RequestTransfer', {
				copy_id: body.copyId,
				transfer_type: body.transferType,
				source_node: sourceNode,
				dest_node: body.destNode,
				initiated_by: locals.user?.userId ?? '',
				notes: body.notes ?? ''
			});
			return json(data, { status: 201 });
		} else {
			const ack = await call(getNetworkClient(), 'InitiateRemoteTransfer', {
				transfer_id: '',
				copy_id: body.copyId,
				transfer_type: body.transferType,
				source_node: sourceNode,
				dest_node: body.destNode || nodeId,
				initiated_by: locals.user?.userId ?? '',
				user_jwt: '',
				notes: body.notes ?? ''
			});
			return json(ack, { status: 201 });
		}
	} catch (err) {
		throw error(400, grpcMessage(err));
	}
};
