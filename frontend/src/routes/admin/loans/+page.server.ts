// Admin loans view — a dedicated list-all-loans RPC is not yet exposed by the BFF.
// Stub until GET /loans is wired to curios-manager's loan store.
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async () => {
	return { loans: [] };
};
