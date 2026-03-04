import type { Handle } from '@sveltejs/kit';
import { cookieName, decodeJWTPayload } from '$lib/server/auth';
import { getUsersClient, getDirectoryClient, call } from '$lib/server/grpc/clients';
import { ensureConfig, getPublicUrl, getAllowLocalhost } from '$lib/server/config';
import type { Claim } from '$lib/api';

// Discovered from the gRPC service on first request; cached for the process lifetime.
let _nodeId = '';
let _announced = false;

const SAFE_METHODS = new Set(['GET', 'HEAD', 'OPTIONS', 'TRACE']);

function isLocalhost(origin: string): boolean {
	try {
		const { hostname } = new URL(origin);
		// Named loopback
		if (hostname === 'localhost') return true;
		// IPv6 loopback
		if (hostname === '::1') return true;
		// All-interfaces bind address used by some dev servers
		if (hostname === '0.0.0.0') return true;
		// Full 127.0.0.0/8 loopback range (127.x.x.x)
		if (/^127\.\d{1,3}\.\d{1,3}\.\d{1,3}$/.test(hostname)) return true;
		return false;
	} catch {
		return false;
	}
}

export const handle: Handle = async ({ event, resolve }) => {
	// Announce UI to network-manager's LocalDirectory once per process lifetime.
	if (!_announced) {
		try {
			await call(getDirectoryClient(), 'Announce', { name: 'ui', address: '' });
			_announced = true;
		} catch {
			// network-manager not yet reachable — will retry on the next request
		}
	}

	if (!_nodeId) {
		try {
			const res = await call<{ node_id: string }>(getUsersClient(), 'GetNodeID', {});
			_nodeId = res.node_id ?? '';
		} catch {
			// Service not yet reachable — will retry on the next request
		}
	}
	event.locals.nodeId = _nodeId;

	await ensureConfig();

	// Custom CSRF guard.
	// - Before setup (publicUrl empty): exempt — no sessions exist yet.
	// - After setup: allow requests from the configured public URL.
	// - Also allow localhost origins when allow_localhost is true (default).
	const publicUrl = getPublicUrl();
	if (publicUrl && !SAFE_METHODS.has(event.request.method)) {
		const origin = event.request.headers.get('origin');
		if (origin && origin !== publicUrl) {
			if (!getAllowLocalhost() || !isLocalhost(origin)) {
				return new Response('Cross-site request forbidden', { status: 403 });
			}
		}
	}

	const token = event.cookies.get(cookieName(event.locals.nodeId));
	if (token) {
		try {
			const payload = decodeJWTPayload(token);
			event.locals.user = {
				userId: payload.uid as string,
				claims: (payload.claims ?? []) as Claim[],
				homeNode: payload.home_node
			};
		} catch {
			event.locals.user = null;
		}
	} else {
		event.locals.user = null;
	}

	return resolve(event);
};
