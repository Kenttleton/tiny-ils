import { redirect } from '@sveltejs/kit';
import { isManager } from '$lib/auth';
import type { LayoutServerLoad } from './$types';

export const load: LayoutServerLoad = async ({ locals, parent }) => {
	await parent(); // ensures root layout auth guard runs first
	if (!locals.user || !isManager(locals.user.claims, locals.nodeId)) {
		throw redirect(303, '/browse');
	}
	return {};
};
