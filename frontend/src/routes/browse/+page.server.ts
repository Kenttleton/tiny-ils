import { serverFetch } from '$lib/server/bff';
import type { CurioList } from '$lib/api';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ request, url }) => {
	const cookie = request.headers.get('cookie') ?? '';
	const q = url.searchParams.get('q') ?? '';
	const mediaType = url.searchParams.get('mediaType') ?? '';

	const params = new URLSearchParams({ q, mediaType, limit: '50' });
	const { data } = await serverFetch<CurioList>('GET', `/curios?${params}`, cookie);
	return { curios: data.curios ?? [], total: data.total ?? 0, q, mediaType };
};
