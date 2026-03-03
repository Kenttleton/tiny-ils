import adapter from '@sveltejs/adapter-node';

/** @type {import('@sveltejs/kit').Config} */
const config = {
	kit: {
		adapter: adapter(),
		// Disable the built-in CSRF check; we implement our own in hooks.server.ts
		// using the public_url stored in the database during first-run setup.
		csrf: { checkOrigin: false }
	}
};

export default config;
