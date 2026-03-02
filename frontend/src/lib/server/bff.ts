// Server-side BFF helper — forwards cookies from the browser request to the BFF.
// Only import from .server.ts files or hooks.server.ts.

const BFF = process.env.BFF_URL ?? 'http://localhost:3001';

export interface BffResponse<T> {
	data: T;
	setCookie?: string;
}

export async function serverFetch<T>(
	method: string,
	path: string,
	cookieHeader: string,
	body?: unknown
): Promise<BffResponse<T>> {
	const res = await fetch(`${BFF}${path}`, {
		method,
		headers: {
			...(body ? { 'Content-Type': 'application/json' } : {}),
			cookie: cookieHeader
		},
		body: body ? JSON.stringify(body) : undefined
	});

	if (!res.ok) {
		const err = await res.json().catch(() => ({ error: res.statusText }));
		throw new Error(err.error ?? res.statusText);
	}

	const data = res.status === 204 ? (undefined as T) : ((await res.json()) as T);
	const setCookie = res.headers.get('set-cookie') ?? undefined;
	return { data, setCookie };
}

/** Parse the first Set-Cookie entry into name, value, and options for cookies.set(). */
export function parseSetCookie(header: string): {
	name: string;
	value: string;
	path: string;
	httpOnly: boolean;
	secure: boolean;
	sameSite: 'lax' | 'strict' | 'none';
	maxAge?: number;
} {
	const parts = header.split(';').map((p) => p.trim());
	const [nameValue, ...attrs] = parts;
	const eqIdx = nameValue.indexOf('=');
	const name = nameValue.slice(0, eqIdx);
	const value = nameValue.slice(eqIdx + 1);

	let path = '/';
	let httpOnly = false;
	let secure = false;
	let sameSite: 'lax' | 'strict' | 'none' = 'lax';
	let maxAge: number | undefined;

	for (const attr of attrs) {
		const lower = attr.toLowerCase();
		if (lower === 'httponly') httpOnly = true;
		else if (lower === 'secure') secure = true;
		else if (lower.startsWith('path=')) path = attr.slice(5);
		else if (lower.startsWith('samesite=')) {
			const val = attr.slice(9).toLowerCase();
			if (val === 'strict' || val === 'none') sameSite = val;
			else sameSite = 'lax';
		} else if (lower.startsWith('max-age=')) maxAge = parseInt(attr.slice(8));
	}

	return { name, value, path, httpOnly, secure, sameSite, maxAge };
}
