// Placeholder — a dedicated "my loans" endpoint is not yet implemented on the BFF.
// This will be wired up when GET /loans/me is added to curios-manager.
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async () => {
	return { loans: [] };
};
