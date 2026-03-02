import { serverFetch } from '$lib/server/bff';
import type { CurioList } from '$lib/api';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ request }) => {
	const cookie = request.headers.get('cookie') ?? '';
	// Load a small snapshot of recent curios as a dashboard summary.
	const { data } = await serverFetch<CurioList>('GET', '/curios?limit=10&offset=0', cookie).catch(
		() => ({ data: { curios: [], total: 0 } as CurioList })
	);
	return { recentCurios: data.curios ?? [], total: data.total ?? 0 };
};
