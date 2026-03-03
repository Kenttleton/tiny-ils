import { json, error } from '@sveltejs/kit';
import { getNetworkClient, grpcMessage } from '$lib/server/grpc/clients';
import type { RequestHandler } from './$types';
import type { NetworkSearchResult } from '$lib/api';

export const GET: RequestHandler = async ({ url }) => {
	const q = url.searchParams.get('q') ?? '';
	const mediaType = url.searchParams.get('mediaType') ?? '';

	return new Promise((resolve) => {
		const client = getNetworkClient();
		// eslint-disable-next-line @typescript-eslint/no-explicit-any
		const stream = (client as any).SearchNetwork({ query: q, media_type: mediaType });
		const results: NetworkSearchResult[] = [];

		stream.on('data', (r: NetworkSearchResult) => results.push(r));
		stream.on('end', () => resolve(json({ results })));
		stream.on('error', (err: unknown) =>
			resolve(error(500, grpcMessage(err)) as unknown as Response)
		);
	});
};
