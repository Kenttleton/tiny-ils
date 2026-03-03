import { getUsersClient, call } from './grpc/clients';

// Module-level cache — lives for the lifetime of the Node.js process.
// Call refresh() after any UpdateAppConfig RPC to apply changes immediately.
let _publicUrl = process.env.ORIGIN ?? '';
let _allowLocalhost = true; // default: allow local dev without explicit config
let _fetched = false;

export async function ensureConfig(): Promise<void> {
	if (_fetched) return;
	try {
		const res = await call<{ public_url: string; allow_localhost: boolean }>(
			getUsersClient(),
			'GetAppConfig',
			{}
		);
		if (res.public_url) _publicUrl = res.public_url;
		// allow_localhost defaults to true when absent from DB
		_allowLocalhost = res.allow_localhost !== false;
		_fetched = true;
	} catch {
		// gRPC not yet reachable — will retry on next request
	}
}

export function getPublicUrl(): string {
	return _publicUrl;
}

export function getAllowLocalhost(): boolean {
	return _allowLocalhost;
}

/** Called by the admin settings action to update the cache without a server restart. */
export function applyConfigUpdate(publicUrl: string, allowLocalhost: boolean): void {
	if (publicUrl) _publicUrl = publicUrl;
	_allowLocalhost = allowLocalhost;
	_fetched = true;
}
