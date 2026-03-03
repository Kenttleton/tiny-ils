import { redirect } from '@sveltejs/kit';
import type { LayoutServerLoad } from './$types';

export const load: LayoutServerLoad = async ({ locals, url, parent }) => {
	await parent();
	if (!locals.user) throw redirect(303, `/auth/login?next=${encodeURIComponent(url.pathname)}`);
	return {};
};
