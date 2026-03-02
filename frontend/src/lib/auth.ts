import type { Claim } from './api';

export function isManager(claims: Claim[], nodeId: string): boolean {
	return claims.some((c) => c.node === nodeId && c.role === 'MANAGER');
}
