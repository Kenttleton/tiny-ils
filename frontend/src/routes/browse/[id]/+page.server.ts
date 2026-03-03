import { error, fail, redirect } from '@sveltejs/kit';
import { getCuriosClient, call, grpcMessage } from '$lib/server/grpc/clients';
import type { Curio, CopyList } from '$lib/api';
import type { Actions, PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ params }) => {
	const [curioRes, copiesRes] = await Promise.allSettled([
		call<Curio>(getCuriosClient(), 'GetCurio', { id: params.id }),
		call<CopyList>(getCuriosClient(), 'ListCopies', { id: params.id })
	]);

	if (curioRes.status === 'rejected') throw error(404, 'Curio not found');

	return {
		curio: curioRes.value,
		copies: copiesRes.status === 'fulfilled' ? (copiesRes.value.copies ?? []) : []
	};
};

export const actions: Actions = {
	checkout: async ({ request, params, locals }) => {
		if (!locals.user) throw redirect(303, `/auth/login?next=/browse/${params.id}`);
		const form = await request.formData();
		const copyId = form.get('copyId')?.toString();
		if (!copyId) return fail(400, { error: 'Missing copy ID' });
		try {
			await call(getCuriosClient(), 'CheckoutCopy', {
				copy_id: copyId,
				user_id: locals.user.userId,
				user_node_id: locals.nodeId,
				due_date: 0
			});
			return { success: true, action: 'checkout' };
		} catch (err) {
			return fail(400, { error: grpcMessage(err) });
		}
	},

	hold: async ({ params, locals }) => {
		if (!locals.user) throw redirect(303, `/auth/login?next=/browse/${params.id}`);
		try {
			await call(getCuriosClient(), 'PlaceHold', {
				curio_id: params.id,
				user_id: locals.user.userId,
				user_node_id: locals.nodeId
			});
			return { success: true, action: 'hold' };
		} catch (err) {
			return fail(400, { error: grpcMessage(err) });
		}
	}
};
