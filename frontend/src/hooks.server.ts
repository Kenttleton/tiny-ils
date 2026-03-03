import type { Handle } from '@sveltejs/kit';
import { COOKIE, decodeJWTPayload } from '$lib/server/auth';
import type { Claim } from '$lib/api';

export const handle: Handle = async ({ event, resolve }) => {
	event.locals.nodeId = process.env.NODE_ID ?? '';

	const token = event.cookies.get(COOKIE);
	if (token) {
		try {
			const payload = decodeJWTPayload(token);
			event.locals.user = {
				userId: payload.uid as string,
				claims: (payload.claims ?? []) as Claim[]
			};
		} catch {
			event.locals.user = null;
		}
	} else {
		event.locals.user = null;
	}

	return resolve(event);
};
