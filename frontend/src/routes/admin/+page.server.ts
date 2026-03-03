import { getCuriosClient, getNetworkClient, call } from '$lib/server/grpc/clients';
import type { CurioList, PeerList } from '$lib/api';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async () => {
	const [curioData, peerData] = await Promise.all([
		call<CurioList>(getCuriosClient(), 'ListCurios', {
			query: '',
			media_type: '',
			limit: 10,
			offset: 0
		}).catch(() => ({ curios: [], total: 0 } as CurioList)),
		call<PeerList>(getNetworkClient(), 'ListPeers', {}).catch(() => ({ peers: [] } as PeerList))
	]);
	return {
		recentCurios: curioData.curios ?? [],
		total: curioData.total ?? 0,
		peers: peerData.peers ?? []
	};
};
