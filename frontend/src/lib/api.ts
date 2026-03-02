// Typed fetch wrappers for the BFF REST API.
// All calls go through the BFF — never directly to Go gRPC services.

const BFF = typeof window !== 'undefined'
  ? (import.meta.env.PUBLIC_BFF_URL ?? 'http://localhost:3001')
  : (process.env.BFF_URL ?? 'http://localhost:3001');

async function bff<T>(method: string, path: string, body?: unknown): Promise<T> {
  const res = await fetch(`${BFF}${path}`, {
    method,
    credentials: 'include',
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
  me: () => bff<{ userId: string; claims: Claim[] }>('GET', '/auth/me'),
  login: (email: string, password: string) =>
    bff<{ user: User }>('POST', '/auth/login', { email, password }),
  register: (email: string, password: string, displayName?: string) =>
    bff<{ user: User }>('POST', '/auth/register', { email, password, displayName }),
  logout: () => bff<void>('POST', '/auth/logout'),
  googleLogin: () => { window.location.href = `${BFF}/auth/google`; },
};

// ─── Curios ───────────────────────────────────────────────────────────────────

export const curios = {
  list: (params?: { q?: string; mediaType?: string; limit?: number; offset?: number }) =>
    bff<CurioList>('GET', '/curios?' + new URLSearchParams(params as Record<string, string> ?? {}).toString()),
  get: (id: string) => bff<Curio>('GET', `/curios/${id}`),
  create: (data: CreateCurioInput) => bff<Curio>('POST', '/curios', data),
  update: (id: string, data: Partial<CreateCurioInput>) => bff<Curio>('PUT', `/curios/${id}`, data),
  delete: (id: string) => bff<void>('DELETE', `/curios/${id}`),
  enrich: (mediaType: string, identifier: string) =>
    bff<CurioMetadata>('POST', '/curios/enrich', { mediaType, identifier }),
  listCopies: (id: string) => bff<CopyList>('GET', `/curios/${id}/copies`),
  placeHold: (id: string) => bff<Hold>('POST', `/curios/${id}/hold`, {}),
};

// ─── Loans ────────────────────────────────────────────────────────────────────

export const loans = {
  checkout: (copyId: string) => bff<PhysicalLoan>('POST', `/copies/${copyId}/checkout`, {}),
  return: (copyId: string) => bff<PhysicalLoan>('POST', `/copies/${copyId}/return`, {}),
  cancelHold: (holdId: string) => bff<void>('DELETE', `/holds/${holdId}`),
};

// ─── Search ───────────────────────────────────────────────────────────────────

export const search = {
  local: (q: string, mediaType?: string) =>
    bff<CurioList>('GET', `/search/local?q=${encodeURIComponent(q)}&mediaType=${mediaType ?? ''}`),
  network: (q: string, mediaType?: string) =>
    bff<{ results: NetworkSearchResult[] }>('GET', `/search/network?q=${encodeURIComponent(q)}&mediaType=${mediaType ?? ''}`),
};

// ─── Admin ────────────────────────────────────────────────────────────────────

export const peers = {
  list: () => bff<PeerList>('GET', '/peers'),
  register: (data: { nodeId: string; publicKey: string; address: string; displayName?: string }) =>
    bff<object>('POST', '/peers', data),
};

export const claims = {
  list: () => bff<ClaimList>('GET', '/claims'),
  grant: (userId: string, role: string, nodeId?: string) =>
    bff<void>('POST', '/claims/grant', { userId, role, nodeId }),
  revoke: (userId: string, nodeId?: string) =>
    bff<void>('DELETE', '/claims/revoke', { userId, nodeId }),
};

export const transfers = {
  list: (params?: { status?: string; nodeId?: string; transferType?: string }) =>
    bff<TransferList>('GET', '/transfers?' + new URLSearchParams((params ?? {}) as Record<string, string>).toString()),
  get: (id: string) => bff<CopyTransfer>('GET', `/transfers/${id}`),
  request: (data: RequestTransferInput) => bff<CopyTransfer>('POST', '/transfers', data),
  approve: (id: string, notes?: string) => bff<CopyTransfer>('POST', `/transfers/${id}/approve`, { notes }),
  reject: (id: string, notes?: string) => bff<CopyTransfer>('POST', `/transfers/${id}/reject`, { notes }),
  ship: (id: string, notes?: string) => bff<CopyTransfer>('POST', `/transfers/${id}/ship`, { notes }),
  receive: (id: string, notes?: string) => bff<CopyTransfer>('POST', `/transfers/${id}/receive`, { notes }),
  cancel: (id: string, notes?: string) => bff<CopyTransfer>('POST', `/transfers/${id}/cancel`, { notes }),
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
