// See https://svelte.dev/docs/kit/types#app.d.ts
// for information about these interfaces
declare global {
	namespace App {
		interface Locals {
			user: { userId: string; claims: import('$lib/api').Claim[] } | null;
			nodeId: string;
		}
		interface PageData {
			user: { userId: string; claims: import('$lib/api').Claim[] } | null;
			nodeId: string;
		}
	}
}

export {};
