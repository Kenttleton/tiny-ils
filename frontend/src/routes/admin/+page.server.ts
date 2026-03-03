import { getCuriosClient, call } from '$lib/server/grpc/clients';
import type { CurioList } from '$lib/api';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async () => {
	// Load a small snapshot of recent curios as a dashboard summary.
	const data = await call<CurioList>(getCuriosClient(), 'ListCurios', {
		query: '',
		media_type: '',
		limit: 10,
		offset: 0
	}).catch(() => ({ curios: [], total: 0 } as CurioList));
	return { recentCurios: data.curios ?? [], total: data.total ?? 0 };
};
