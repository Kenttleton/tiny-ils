import type { Cookies } from '@sveltejs/kit';
import type { Claim } from '$lib/api';

export function cookieName(nodeId: string): string {
	return nodeId ? `${nodeId}_token` : 'token';
}

export function setAuthCookie(cookies: Cookies, jwt: string, nodeId: string): void {
	cookies.set(cookieName(nodeId), jwt, {
		path: '/',
		httpOnly: true,
		sameSite: 'lax',
		secure: process.env.NODE_ENV === 'production',
		maxAge: 60 * 60 * 24 // 24 hours
	});
}

export function clearAuthCookie(cookies: Cookies, nodeId: string): void {
	cookies.delete(cookieName(nodeId), { path: '/' });
}

/** Decode a JWT payload without verifying the signature. Used only to read claims for UI. */
export function decodeJWTPayload(token: string): { uid?: string; claims?: Claim[]; home_node?: string } {
	const parts = token.split('.');
	if (parts.length !== 3) throw new Error('Invalid JWT');
	const payload = Buffer.from(parts[1], 'base64url').toString('utf8');
	return JSON.parse(payload);
}
