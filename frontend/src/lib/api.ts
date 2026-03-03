// Typed fetch wrappers for the SvelteKit API routes.
// All calls are same-origin relative paths — no separate BFF service.

async function api<T>(method: string, path: string, body?: unknown): Promise<T> {
  const res = await fetch(path, {
    method,
    headers: body ? { 'Content-Type': 'application/json' } : {},
    body: body ? JSON.stringify(body) : undefined,
  });
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: res.statusText }));
    throw new Error(err.error ?? res.statusText);
  }
  if (res.status === 204) return undefined as T;
  return res.json();
}

// ─── Auth ─────────────────────────────────────────────────────────────────────

export const auth = {
  googleLogin: () => { window.location.href = '/auth/google'; },
};

// ─── Curios ───────────────────────────────────────────────────────────────────

export const curios = {
  list: (params?: { q?: string; mediaType?: string; limit?: number; offset?: number }) =>
    api<CurioList>('GET', '/api/curios?' + new URLSearchParams((params ?? {}) as Record<string, string>).toString()),
  get: (id: string) => api<Curio>('GET', `/api/curios/${id}`),
  create: (data: CreateCurioInput) => api<Curio>('POST', '/api/curios', data),
  update: (id: string, data: Partial<CreateCurioInput>) => api<Curio>('PUT', `/api/curios/${id}`, data),
  delete: (id: string) => api<void>('DELETE', `/api/curios/${id}`),
  enrich: (mediaType: string, identifier: string) =>
    api<CurioMetadata>('POST', '/api/curios/enrich', { mediaType, identifier }),
  listCopies: (id: string) => api<CopyList>('GET', `/api/curios/${id}/copies`),
  placeHold: (id: string) => api<Hold>('POST', `/api/curios/${id}/hold`, {}),
};

// ─── Loans ────────────────────────────────────────────────────────────────────

export const loans = {
  checkout: (copyId: string) => api<PhysicalLoan>('POST', `/api/copies/${copyId}/checkout`, {}),
  return: (copyId: string) => api<PhysicalLoan>('POST', `/api/copies/${copyId}/return`, {}),
  cancelHold: (holdId: string) => api<void>('DELETE', `/api/holds/${holdId}`),
};

// ─── Search ───────────────────────────────────────────────────────────────────

export const search = {
  local: (q: string, mediaType?: string) =>
    api<CurioList>('GET', `/api/search/local?q=${encodeURIComponent(q)}&mediaType=${mediaType ?? ''}`),
  network: (q: string, mediaType?: string) =>
    api<{ results: NetworkSearchResult[] }>('GET', `/api/search/network?q=${encodeURIComponent(q)}&mediaType=${mediaType ?? ''}`),
};

// ─── Admin ────────────────────────────────────────────────────────────────────

export const peers = {
  list: () => api<PeerList>('GET', '/api/peers'),
  register: (data: { nodeId: string; publicKey: string; address: string; displayName?: string }) =>
    api<object>('POST', '/api/peers', data),
};

export const claims = {
  list: () => api<ClaimList>('GET', '/api/claims'),
  grant: (userId: string, role: string, nodeId?: string) =>
    api<void>('POST', '/api/claims/grant', { userId, role, nodeId }),
  revoke: (userId: string, nodeId?: string) =>
    api<void>('DELETE', '/api/claims/revoke', { userId, nodeId }),
};

export const transfers = {
  list: (params?: { status?: string; nodeId?: string; transferType?: string }) =>
    api<TransferList>('GET', '/api/transfers?' + new URLSearchParams((params ?? {}) as Record<string, string>).toString()),
  get: (id: string) => api<CopyTransfer>('GET', `/api/transfers/${id}`),
  request: (data: RequestTransferInput) => api<CopyTransfer>('POST', '/api/transfers', data),
  approve: (id: string, notes?: string) => api<CopyTransfer>('POST', `/api/transfers/${id}/approve`, { notes }),
  reject: (id: string, notes?: string) => api<CopyTransfer>('POST', `/api/transfers/${id}/reject`, { notes }),
  ship: (id: string, notes?: string) => api<CopyTransfer>('POST', `/api/transfers/${id}/ship`, { notes }),
  receive: (id: string, notes?: string) => api<CopyTransfer>('POST', `/api/transfers/${id}/receive`, { notes }),
  cancel: (id: string, notes?: string) => api<CopyTransfer>('POST', `/api/transfers/${id}/cancel`, { notes }),
};

// ─── Types ────────────────────────────────────────────────────────────────────

export interface Claim { node: string; role: string; }
export interface User { id: string; email: string; displayName: string; ssoProvider?: string; }
export interface Curio {
  id: string; title: string; description: string;
  mediaType: string; formatType: string; tags: string[];
  barcode?: string; qrCode?: string; createdAt: number;
}
export interface CurioList { curios: Curio[]; total: number; }
export interface CreateCurioInput { title: string; description?: string; mediaType: string; formatType: string; tags?: string[]; barcode?: string; }
export interface CurioMetadata { title: string; description: string; authors: string[]; coverUrl: string; tags: string[]; source: string; }
export interface PhysicalCopy { id: string; curioId: string; condition: string; location: string; status: string; }
export interface CopyList { copies: PhysicalCopy[]; }
export interface PhysicalLoan { id: string; copyId: string; userId: string; checkedOut: number; dueDate: number; returnedAt?: number; }
export interface Hold { id: string; curioId: string; userId: string; placedAt: number; }
export interface NetworkSearchResult { nodeId: string; nodeName: string; curios: Curio[]; error?: string; }
export interface PeerList { peers: { nodeId: string; publicKey: string; address: string; displayName: string }[]; }
export interface ClaimList { claims: { userId: string; nodeId: string; role: string; }[]; }
export interface CopyTransfer {
  id: string; copyId: string; transferType: string;
  sourceNode: string; destNode: string;
  initiatedBy: string; approvedBy?: string;
  status: string; notes?: string;
  requestedAt: number; approvedAt?: number; shippedAt?: number; receivedAt?: number;
}
export interface TransferList { transfers: CopyTransfer[]; }
export interface RequestTransferInput {
  copyId: string; transferType: string;
  sourceNode: string; destNode: string; notes?: string;
}
