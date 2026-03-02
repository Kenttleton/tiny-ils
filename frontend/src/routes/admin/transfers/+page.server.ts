import { fail } from '@sveltejs/kit';
import type { PageServerLoad, Actions } from './$types';
import { serverFetch } from '$lib/server/bff';

export const load: PageServerLoad = async ({ locals, request, url }) => {
  const cookie = request.headers.get('cookie') ?? '';
  const nodeId = locals.nodeId;
  const tab = url.searchParams.get('tab') ?? 'incoming';

  const [incomingRes, outgoingRes, historyRes] = await Promise.all([
    serverFetch('GET', `/transfers?nodeId=${encodeURIComponent(nodeId)}&status=PENDING`, cookie)
      .catch(() => ({ transfers: [] })),
    serverFetch('GET', `/transfers?nodeId=${encodeURIComponent(nodeId)}&status=APPROVED`, cookie)
      .catch(() => ({ transfers: [] })),
    serverFetch('GET', `/transfers?nodeId=${encodeURIComponent(nodeId)}&status=IN_TRANSIT`, cookie)
      .catch(() => ({ transfers: [] })),
  ]);

  // Combine active transfers and split by direction
  const active = [
    ...((incomingRes as { transfers?: unknown[] }).transfers ?? []),
    ...((outgoingRes as { transfers?: unknown[] }).transfers ?? []),
    ...((historyRes as { transfers?: unknown[] }).transfers ?? []),
  ] as import('$lib/api').CopyTransfer[];

  const incoming = active.filter((t) => t.destNode === nodeId);
  const outgoing = active.filter((t) => t.sourceNode === nodeId && t.destNode !== nodeId);

  const histRes = await serverFetch(
    'GET',
    `/transfers?nodeId=${encodeURIComponent(nodeId)}&status=RECEIVED`,
    cookie,
  ).catch(() => ({ transfers: [] }));
  const rejRes = await serverFetch(
    'GET',
    `/transfers?nodeId=${encodeURIComponent(nodeId)}&status=REJECTED`,
    cookie,
  ).catch(() => ({ transfers: [] }));
  const canRes = await serverFetch(
    'GET',
    `/transfers?nodeId=${encodeURIComponent(nodeId)}&status=CANCELLED`,
    cookie,
  ).catch(() => ({ transfers: [] }));

  const history = [
    ...((histRes as { transfers?: import('$lib/api').CopyTransfer[] }).transfers ?? []),
    ...((rejRes as { transfers?: import('$lib/api').CopyTransfer[] }).transfers ?? []),
    ...((canRes as { transfers?: import('$lib/api').CopyTransfer[] }).transfers ?? []),
  ].sort((a, b) => (b.receivedAt ?? b.requestedAt) - (a.receivedAt ?? a.requestedAt));

  return { nodeId, incoming, outgoing, history, tab };
};

export const actions: Actions = {
  approve: async ({ request, locals }) => {
    const data = await request.formData();
    const id = String(data.get('id'));
    const cookie = request.headers.get('cookie') ?? '';
    try {
      await serverFetch('POST', `/transfers/${id}/approve`, cookie, {});
    } catch (e) {
      return fail(400, { error: String(e) });
    }
  },

  reject: async ({ request }) => {
    const data = await request.formData();
    const id = String(data.get('id'));
    const notes = String(data.get('notes') ?? '');
    const cookie = request.headers.get('cookie') ?? '';
    try {
      await serverFetch('POST', `/transfers/${id}/reject`, cookie, { notes });
    } catch (e) {
      return fail(400, { error: String(e) });
    }
  },

  ship: async ({ request }) => {
    const data = await request.formData();
    const id = String(data.get('id'));
    const cookie = request.headers.get('cookie') ?? '';
    try {
      await serverFetch('POST', `/transfers/${id}/ship`, cookie, {});
    } catch (e) {
      return fail(400, { error: String(e) });
    }
  },

  receive: async ({ request }) => {
    const data = await request.formData();
    const id = String(data.get('id'));
    const cookie = request.headers.get('cookie') ?? '';
    try {
      await serverFetch('POST', `/transfers/${id}/receive`, cookie, {});
    } catch (e) {
      return fail(400, { error: String(e) });
    }
  },

  cancel: async ({ request }) => {
    const data = await request.formData();
    const id = String(data.get('id'));
    const cookie = request.headers.get('cookie') ?? '';
    try {
      await serverFetch('POST', `/transfers/${id}/cancel`, cookie, {});
    } catch (e) {
      return fail(400, { error: String(e) });
    }
  },

  request: async ({ request, locals }) => {
    const data = await request.formData();
    const cookie = request.headers.get('cookie') ?? '';
    const body = {
      copyId: String(data.get('copyId')),
      transferType: String(data.get('transferType')),
      sourceNode: String(data.get('sourceNode') || locals.nodeId),
      destNode: String(data.get('destNode')),
      notes: String(data.get('notes') ?? ''),
    };
    try {
      await serverFetch('POST', '/transfers', cookie, body);
    } catch (e) {
      return fail(400, { error: String(e) });
    }
  },
};
