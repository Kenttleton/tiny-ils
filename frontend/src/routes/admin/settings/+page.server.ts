import { fail } from '@sveltejs/kit';
import { getUsersClient, call, grpcMessage } from '$lib/server/grpc/clients';
import { getPublicUrl, getAllowLocalhost, applyConfigUpdate } from '$lib/server/config';
import type { Actions, PageServerLoad } from './$types';

export const load: PageServerLoad = async () => {
	return {
		publicUrl: getPublicUrl(),
		allowLocalhost: getAllowLocalhost()
	};
};

export const actions: Actions = {
	default: async ({ request }) => {
		const data = await request.formData();
		const publicUrl = String(data.get('publicUrl') ?? '').trim().replace(/\/$/, '');
		const allowLocalhost = data.get('allowLocalhost') === 'true';

		if (!publicUrl) {
			return fail(400, { error: 'Public URL is required.', allowLocalhost });
		}

		try {
			await call(getUsersClient(), 'UpdateAppConfig', {
				public_url: publicUrl,
				allow_localhost: allowLocalhost
			});
			// Apply immediately to the in-process cache so the change takes effect
			// without requiring a server restart.
			applyConfigUpdate(publicUrl, allowLocalhost);
			return { success: true, publicUrl, allowLocalhost };
		} catch (err) {
			return fail(400, { error: grpcMessage(err), allowLocalhost });
		}
	}
};
