import fs from 'node:fs';
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

// mTLS credentials for the network-manager. The BFF authenticates as this node
// using the shared node_identity key pair — the same identity used by peers.
// Falls back to insecure if the key files are not yet available (e.g. during build).
function makeNetworkCreds(): grpc.ChannelCredentials {
	const keyPath = process.env.NODE_KEY_PATH ?? '/data/node.key';
	const certPath = process.env.NODE_CERT_PATH ?? '/data/node.crt';
	try {
		const key = fs.readFileSync(keyPath);
		const cert = fs.readFileSync(certPath);
		// Use own cert as root CA — the server presents the same cert, so signature
		// verification succeeds. Hostname verification is skipped because the cert
		// CN is the node Library ID, not the Docker hostname.
		return grpc.credentials.createSsl(cert, key, cert, {
			checkServerIdentity: () => undefined
		});
	} catch {
		return grpc.credentials.createInsecure();
	}
}

// Clients are lazy-initialised so proto files are only loaded at runtime,
// not during Vite's build-time SSR analysis.
// eslint-disable-next-line @typescript-eslint/no-explicit-any
let _curios: any, _users: any, _network: any;

// ─── Curios Manager ──────────────────────────────────────────────────────────

export function getCuriosClient(): grpc.Client {
	if (!_curios) {
		// eslint-disable-next-line @typescript-eslint/no-explicit-any
		const pkg = (grpc.loadPackageDefinition(loadProto('curios.proto')) as any).curios;
		_curios = new pkg.CuriosManager(process.env.CURIOS_GRPC ?? 'localhost:50151', creds);
	}
	return _curios as grpc.Client;
}

// ─── Users Manager ───────────────────────────────────────────────────────────

export function getUsersClient(): grpc.Client {
	if (!_users) {
		// eslint-disable-next-line @typescript-eslint/no-explicit-any
		const pkg = (grpc.loadPackageDefinition(loadProto('users.proto')) as any).users;
		_users = new pkg.UsersManager(process.env.USERS_GRPC ?? 'localhost:50152', creds);
	}
	return _users as grpc.Client;
}

// ─── Network Manager ─────────────────────────────────────────────────────────

export function getNetworkClient(): grpc.Client {
	if (!_network) {
		// eslint-disable-next-line @typescript-eslint/no-explicit-any
		const pkg = (grpc.loadPackageDefinition(loadProto('network.proto')) as any).network;
		_network = new pkg.NetworkManager(
			process.env.NETWORK_GRPC ?? 'localhost:50153',
			makeNetworkCreds()
		);
	}
	return _network as grpc.Client;
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

/** Collect all messages from a server-streaming RPC into an array. */
export function callStream<T>(client: grpc.Client, method: string, req: object): Promise<T[]> {
	return new Promise((resolve, reject) => {
		// eslint-disable-next-line @typescript-eslint/no-explicit-any
		const stream = (client as any)[method](req) as grpc.ClientReadableStream<T>;
		const results: T[] = [];
		stream.on('data', (msg: T) => results.push(msg));
		stream.on('end', () => resolve(results));
		stream.on('error', (err: grpc.ServiceError) => reject(err));
	});
}

/** Extract a human-readable message from a gRPC error. */
export function grpcMessage(err: unknown): string {
	if (err && typeof err === 'object' && 'details' in err) return (err as { details: string }).details;
	return String(err);
}
