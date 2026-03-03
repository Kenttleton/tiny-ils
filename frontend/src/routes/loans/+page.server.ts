import { redirect } from '@sveltejs/kit';
import { getCuriosClient, getNetworkClient, call, callStream } from '$lib/server/grpc/clients';
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

// RemoteLoan from UserLoansResult — mirrors proto definition.
interface RemoteLoan {
	loan_id: string;
	curio_id: string;
	curio_title: string;
	is_digital: boolean;
	issued_at: string; // int64 serialised as string by grpc-js
	due_date: string;
	expires_at: string;
	closed: boolean;
	node_id: string;
	node_name: string;
}

interface UserLoansResult {
	node_id: string;
	node_name: string;
	loans: RemoteLoan[];
	error: string;
}

// Unified loan shape passed to the template.
export interface UnifiedLoan {
	id: string;
	curio_id: string;
	curio_title: string;
	is_digital: boolean;
	checked_out: string; // unix seconds as string
	due_date: string;
	closed: boolean;
	node_id: string;
	node_name: string;
}

export const load: PageServerLoad = async ({ locals, url }) => {
	if (!locals.user) throw redirect(303, `/auth/login?next=${encodeURIComponent(url.pathname)}`);

	const activeOnly = url.searchParams.get('active') !== 'false';
	const crossNode = !!locals.user.homeNode;

	if (crossNode) {
		// Cross-node user: fan out via network-manager streaming RPC.
		try {
			const results = await callStream<UserLoansResult>(getNetworkClient(), 'GetUserLoans', {
				user_id: locals.user.userId,
				user_node_id: locals.user.homeNode,
				active_only: activeOnly
			});

			const loans: UnifiedLoan[] = results.flatMap((r) =>
				(r.loans ?? []).map((l) => ({
					id: l.loan_id,
					curio_id: l.curio_id,
					curio_title: l.curio_title,
					is_digital: l.is_digital,
					checked_out: l.issued_at,
					due_date: l.due_date || l.expires_at,
					closed: l.closed,
					node_id: l.node_id || r.node_id,
					node_name: l.node_name || r.node_name
				}))
			);

			// Sort: open loans first, then by checked-out descending.
			loans.sort((a, b) => {
				if (a.closed !== b.closed) return a.closed ? 1 : -1;
				return parseInt(b.checked_out, 10) - parseInt(a.checked_out, 10);
			});

			return { loans, crossNode: true, activeOnly };
		} catch {
			return { loans: [] as UnifiedLoan[], crossNode: true, activeOnly };
		}
	}

	// Local user: direct query to curios-manager with pagination.
	const page = Math.max(1, parseInt(url.searchParams.get('page') ?? '1', 10));
	const limit = 25;
	const offset = (page - 1) * limit;

	try {
		const result = await call<LoanList>(getCuriosClient(), 'ListLoans', {
			active_only: activeOnly,
			user_id: locals.user.userId,
			user_node_id: locals.nodeId,
			limit,
			offset
		});

		const loans: UnifiedLoan[] = (result.loans ?? []).map((l) => ({
			id: l.id,
			curio_id: l.curio_id,
			curio_title: l.curio_title,
			is_digital: false,
			checked_out: l.checked_out,
			due_date: l.due_date,
			closed: !!parseInt(l.returned_at, 10),
			node_id: locals.nodeId,
			node_name: ''
		}));

		return {
			loans,
			total: result.total ?? 0,
			crossNode: false,
			activeOnly,
			page,
			limit
		};
	} catch {
		return { loans: [] as UnifiedLoan[], total: 0, crossNode: false, activeOnly, page: 1, limit };
	}
};
