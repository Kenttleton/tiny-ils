import path from 'path';
import * as grpc from '@grpc/grpc-js';
import * as protoLoader from '@grpc/proto-loader';

// At runtime the protos are at <app root>/proto/ — process.cwd() is the build output dir.
const PROTO_DIR = path.resolve(process.cwd(), 'proto');

function loadProto(filename: string) {
	return protoLoader.loadSync(path.join(PROTO_DIR, filename), {
		keepCase: true,
		longs: String,
		enums: String,
		defaults: true,
		oneofs: true
	});
}

const creds = grpc.credentials.createInsecure();

// ─── Curios Manager ──────────────────────────────────────────────────────────

const curiosDef = loadProto('curios.proto');
// eslint-disable-next-line @typescript-eslint/no-explicit-any
const curiosPkg = (grpc.loadPackageDefinition(curiosDef) as any).curios;

export function getCuriosClient() {
	const addr = process.env.CURIOS_GRPC ?? 'localhost:50151';
	return new curiosPkg.CuriosManager(addr, creds);
}

// ─── Users Manager ───────────────────────────────────────────────────────────

const usersDef = loadProto('users.proto');
// eslint-disable-next-line @typescript-eslint/no-explicit-any
const usersPkg = (grpc.loadPackageDefinition(usersDef) as any).users;

export function getUsersClient() {
	const addr = process.env.USERS_GRPC ?? 'localhost:50152';
	return new usersPkg.UsersManager(addr, creds);
}

// ─── Network Manager ─────────────────────────────────────────────────────────

const networkDef = loadProto('network.proto');
// eslint-disable-next-line @typescript-eslint/no-explicit-any
const networkPkg = (grpc.loadPackageDefinition(networkDef) as any).network;

export function getNetworkClient() {
	const addr = process.env.NETWORK_GRPC ?? 'localhost:50153';
	return new networkPkg.NetworkManager(addr, creds);
}

// ─── Promisify helper ────────────────────────────────────────────────────────

export function call<T>(client: grpc.Client, method: string, req: object): Promise<T> {
	return new Promise((resolve, reject) => {
		// eslint-disable-next-line @typescript-eslint/no-explicit-any
		(client as any)[method](req, (err: grpc.ServiceError | null, res: T) => {
			if (err) reject(err);
			else resolve(res);
		});
	});
}

/** Extract a human-readable message from a gRPC error. */
export function grpcMessage(err: unknown): string {
	if (err && typeof err === 'object' && 'details' in err) return (err as { details: string }).details;
	return String(err);
}
