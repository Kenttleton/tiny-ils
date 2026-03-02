import type { Handle } from '@sveltejs/kit';

const BFF = process.env.BFF_URL ?? 'http://localhost:3001';

// Cache the node ID — it never changes at runtime.
let cachedNodeId: string | null = null;
async function getNodeId(): Promise<string> {
	if (cachedNodeId !== null) return cachedNodeId;
	try {
		const res = await fetch(`${BFF}/node-info`);
		const data = res.ok ? await res.json() : { nodeId: '' };
		cachedNodeId = data.nodeId ?? '';
	} catch {
		cachedNodeId = '';
	}
	return cachedNodeId ?? '';
}

export const handle: Handle = async ({ event, resolve }) => {
	const cookie = event.request.headers.get('cookie') ?? '';

	const [nodeId, meRes] = await Promise.all([
		getNodeId(),
		fetch(`${BFF}/auth/me`, { headers: { cookie } }).catch(() => null)
	]);

	event.locals.nodeId = nodeId;
	event.locals.user = meRes?.ok ? await meRes.json().catch(() => null) : null;

	return resolve(event);
};
