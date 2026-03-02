import { redirect } from '@sveltejs/kit';
import type { RequestHandler } from './$types';

// PUBLIC_BFF_URL must be the browser-visible BFF address (e.g. http://localhost:3001)
const BFF_PUBLIC = process.env.PUBLIC_BFF_URL ?? 'http://localhost:3001';

export const GET: RequestHandler = async () => {
	throw redirect(302, `${BFF_PUBLIC}/auth/google`);
};
