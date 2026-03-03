import { getCuriosClient, call } from '$lib/server/grpc/clients';
import type { PageServerLoad } from './$types';

interface PhysicalLoan {
	id: string;
	copy_id: string;
	user_id: string;
	user_node_id: string;
	checked_out: string;
	due_date: string;
	returned_at: string;
	requesting_node: string;
	curio_id: string;
	curio_title: string;
}

interface LoanList {
	loans: PhysicalLoan[];
	total: number;
}

export const load: PageServerLoad = async ({ url }) => {
	const activeOnly = url.searchParams.get('active') !== 'false';
	const page = Math.max(1, parseInt(url.searchParams.get('page') ?? '1', 10));
	const limit = 50;
	const offset = (page - 1) * limit;

	try {
		const result = await call<LoanList>(getCuriosClient(), 'ListLoans', {
			active_only: activeOnly,
			limit,
			offset
		});
		return {
			loans: result.loans ?? [],
			total: result.total ?? 0,
			activeOnly,
			page,
			limit
		};
	} catch {
		return { loans: [], total: 0, activeOnly, page, limit };
	}
};
