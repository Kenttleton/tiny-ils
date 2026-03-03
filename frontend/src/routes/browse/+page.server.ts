import { getCuriosClient, call } from '$lib/server/grpc/clients';
import type { CurioList } from '$lib/api';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ url }) => {
	const q = url.searchParams.get('q') ?? '';
	const mediaType = url.searchParams.get('mediaType') ?? '';

	const data = await call<CurioList>(getCuriosClient(), 'ListCurios', {
		query: q,
		media_type: mediaType,
		limit: 50,
		offset: 0
	});
	return { curios: data.curios ?? [], total: data.total ?? 0, q, mediaType };
};
