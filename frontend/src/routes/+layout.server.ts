import { redirect } from '@sveltejs/kit';
import type { LayoutServerLoad } from './$types';

const PUBLIC_PATHS = ['/auth/login', '/auth/register'];

export const load: LayoutServerLoad = async ({ locals, url }) => {
	if (!locals.user && !PUBLIC_PATHS.includes(url.pathname)) {
		throw redirect(303, `/auth/login?next=${encodeURIComponent(url.pathname)}`);
	}
	return { user: locals.user, nodeId: locals.nodeId };
};
