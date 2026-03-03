import { fail } from '@sveltejs/kit';
import { getCuriosClient, getNetworkClient, call, grpcMessage } from '$lib/server/grpc/clients';
import type { PageServerLoad, Actions } from './$types';
import type { CopyTransfer, TransferList } from '$lib/api';

export const load: PageServerLoad = async ({ locals, url }) => {
	const nodeId = locals.nodeId;
	const tab = url.searchParams.get('tab') ?? 'incoming';

	const [activeRes, receivedRes, rejectedRes, cancelledRes] = await Promise.all([
		call<TransferList>(getCuriosClient(), 'ListTransfers', { node_id: nodeId, status: '' }).catch(
			() => ({ transfers: [] } as TransferList)
		),
		call<TransferList>(getCuriosClient(), 'ListTransfers', {
			node_id: nodeId,
			status: 'RECEIVED'
		}).catch(() => ({ transfers: [] } as TransferList)),
		call<TransferList>(getCuriosClient(), 'ListTransfers', {
			node_id: nodeId,
			status: 'REJECTED'
		}).catch(() => ({ transfers: [] } as TransferList)),
		call<TransferList>(getCuriosClient(), 'ListTransfers', {
			node_id: nodeId,
			status: 'CANCELLED'
		}).catch(() => ({ transfers: [] } as TransferList))
	]);

	const active = (activeRes.transfers ?? []) as CopyTransfer[];
	const incoming = active.filter(
		(t) =>
			t.destNode === nodeId && ['PENDING', 'APPROVED', 'IN_TRANSIT'].includes(t.status)
	);
	const outgoing = active.filter(
		(t) =>
			t.sourceNode === nodeId &&
			t.destNode !== nodeId &&
			['PENDING', 'APPROVED', 'IN_TRANSIT'].includes(t.status)
	);

	const history = [
		...((receivedRes.transfers ?? []) as CopyTransfer[]),
		...((rejectedRes.transfers ?? []) as CopyTransfer[]),
		...((cancelledRes.transfers ?? []) as CopyTransfer[])
	].sort((a, b) => (b.receivedAt ?? b.requestedAt) - (a.receivedAt ?? a.requestedAt));

	return { nodeId, incoming, outgoing, history, tab };
};

export const actions: Actions = {
	approve: async ({ request, locals }) => {
		const data = await request.formData();
		const id = String(data.get('id'));
		try {
			await call(getCuriosClient(), 'ApproveTransfer', {
				transfer_id: id,
				actor_id: locals.user?.userId ?? '',
				notes: ''
			});
		} catch (e) {
			return fail(400, { error: grpcMessage(e) });
		}
	},

	reject: async ({ request, locals }) => {
		const data = await request.formData();
		const id = String(data.get('id'));
		const notes = String(data.get('notes') ?? '');
		try {
			await call(getCuriosClient(), 'RejectTransfer', {
				transfer_id: id,
				actor_id: locals.user?.userId ?? '',
				notes
			});
		} catch (e) {
			return fail(400, { error: grpcMessage(e) });
		}
	},

	ship: async ({ request, locals }) => {
		const data = await request.formData();
		const id = String(data.get('id'));
		try {
			await call(getCuriosClient(), 'MarkShipped', {
				transfer_id: id,
				actor_id: locals.user?.userId ?? '',
				notes: ''
			});
		} catch (e) {
			return fail(400, { error: grpcMessage(e) });
		}
	},

	receive: async ({ request, locals }) => {
		const data = await request.formData();
		const id = String(data.get('id'));
		try {
			await call(getCuriosClient(), 'ConfirmReceived', {
				transfer_id: id,
				actor_id: locals.user?.userId ?? '',
				notes: ''
			});
		} catch (e) {
			return fail(400, { error: grpcMessage(e) });
		}
	},

	cancel: async ({ request, locals }) => {
		const data = await request.formData();
		const id = String(data.get('id'));
		try {
			await call(getCuriosClient(), 'CancelTransfer', {
				transfer_id: id,
				actor_id: locals.user?.userId ?? '',
				notes: ''
			});
		} catch (e) {
			return fail(400, { error: grpcMessage(e) });
		}
	},

	request: async ({ request, locals }) => {
		const data = await request.formData();
		const nodeId = locals.nodeId;
		const sourceNode = String(data.get('sourceNode') || nodeId);
		const body = {
			copy_id: String(data.get('copyId')),
			transfer_type: String(data.get('transferType')),
			source_node: sourceNode,
			dest_node: String(data.get('destNode')),
			initiated_by: locals.user?.userId ?? '',
			notes: String(data.get('notes') ?? '')
		};
		try {
			if (sourceNode === nodeId || !sourceNode) {
				await call(getCuriosClient(), 'RequestTransfer', body);
			} else {
				await call(getNetworkClient(), 'InitiateRemoteTransfer', {
					transfer_id: '',
					copy_id: body.copy_id,
					transfer_type: body.transfer_type,
					source_node: sourceNode,
					dest_node: body.dest_node || nodeId,
					initiated_by: body.initiated_by,
					user_jwt: '',
					notes: body.notes
				});
			}
		} catch (e) {
			return fail(400, { error: grpcMessage(e) });
		}
	}
};
