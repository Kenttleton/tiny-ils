import { redirect } from '@sveltejs/kit';
import { getUsersClient, call } from '$lib/server/grpc/clients';
import type { LayoutServerLoad } from './$types';

const PUBLIC_PATHS = ['/auth/login', '/auth/register', '/setup'];
const PUBLIC_PREFIXES = ['/browse'];

// Cached once true — resets only on server restart (which is fine)
let _isSetup = false;

export const load: LayoutServerLoad = async ({ locals, url }) => {
	if (!_isSetup) {
		try {
			const status = await call<{ has_manager: boolean }>(getUsersClient(), 'HasSetup', {});
			_isSetup = status.has_manager;
		} catch {
			// gRPC unreachable — leave _isSetup false, user sees setup page
		}
	}

	if (!_isSetup && url.pathname !== '/setup') {
		throw redirect(303, '/setup');
	}
	if (_isSetup && url.pathname === '/setup') {
		throw redirect(303, '/');
	}
	const isPublic =
		PUBLIC_PATHS.includes(url.pathname) ||
		PUBLIC_PREFIXES.some((p) => url.pathname.startsWith(p));
	if (_isSetup && !locals.user && !isPublic) {
		throw redirect(303, `/auth/login?next=${encodeURIComponent(url.pathname)}`);
	}

	return { user: locals.user, nodeId: locals.nodeId };
};
