import { redirect } from '@sveltejs/kit';
import { isManager } from '$lib/auth';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ locals }) => {
	if (!locals.user) throw redirect(303, '/auth/login');
	throw redirect(303, isManager(locals.user.claims, locals.nodeId) ? '/admin' : '/browse');
};
