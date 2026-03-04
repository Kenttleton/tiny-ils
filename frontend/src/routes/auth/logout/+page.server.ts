import { redirect } from '@sveltejs/kit';
import { clearAuthCookie } from '$lib/server/auth';
import type { Actions } from './$types';

export const actions: Actions = {
	default: async ({ cookies, locals }) => {
		clearAuthCookie(cookies, locals.nodeId);
		throw redirect(303, '/auth/login');
	}
};
