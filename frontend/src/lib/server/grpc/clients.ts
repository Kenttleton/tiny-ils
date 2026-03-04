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

// All internal connections use insecure credentials — Docker network isolation
// provides security for the Docker-private internal port.
const creds = grpc.credentials.createInsecure();

// ─── Directory client (network-manager internal port) ─────────────────────────

// eslint-disable-next-line @typescript-eslint/no-explicit-any
let _dir: any;
export function getDirectoryClient(): grpc.Client {
	if (!_dir) {
		// eslint-disable-next-line @typescript-eslint/no-explicit-any
		const pkg = (grpc.loadPackageDefinition(loadProto('internal.proto')) as any).internal;
		_dir = new pkg.LocalDirectory(process.env.NETWORK_GRPC ?? 'localhost:50154', creds);
	}
	return _dir as grpc.Client;
}

// ─── Service discovery ────────────────────────────────────────────────────────

// Discover a service address via Lookup, then verify with WhoAmI.
// Falls back to the env var override if set (useful for local dev/testing).
// Budget: 2 calls (Lookup + WhoAmI) per service — stays within the ≤5 call target.
async function discoverAddress(name: string, envVar: string): Promise<string> {
	const override = process.env[envVar];
	if (override) return override;

	// eslint-disable-next-line @typescript-eslint/no-explicit-any
	const res = await call<any>(getDirectoryClient(), 'Lookup', { name });
	if (!res.services?.length) throw new Error(`no '${name}' services registered with network-manager`);

	// Pick first; future work can fan-out to multiple instances.
	const addr: string = res.services[0].address;

	// Verify reachability via WhoAmI.
	// eslint-disable-next-line @typescript-eslint/no-explicit-any
	const capPkg = (grpc.loadPackageDefinition(loadProto('internal.proto')) as any).internal;
	const capClient = new capPkg.CapabilityService(addr, creds);
	await call(capClient, 'WhoAmI', {});

	return addr;
}

// Lazy-discovered clients — reset to null on error for automatic retry.
// eslint-disable-next-line @typescript-eslint/no-explicit-any
let _curios: any, _users: any, _network: any;

// ─── Curios Manager ───────────────────────────────────────────────────────────

export async function getCuriosClient(): Promise<grpc.Client> {
	if (!_curios) {
		const addr = await discoverAddress('curios', 'CURIOS_GRPC');
		// eslint-disable-next-line @typescript-eslint/no-explicit-any
		const pkg = (grpc.loadPackageDefinition(loadProto('curios.proto')) as any).curios;
		_curios = new pkg.CuriosManager(addr, creds);
	}
	return _curios as grpc.Client;
}

// ─── Users Manager ────────────────────────────────────────────────────────────

export async function getUsersClient(): Promise<grpc.Client> {
	if (!_users) {
		const addr = await discoverAddress('users', 'USERS_GRPC');
		// eslint-disable-next-line @typescript-eslint/no-explicit-any
		const pkg = (grpc.loadPackageDefinition(loadProto('users.proto')) as any).users;
		_users = new pkg.UsersManager(addr, creds);
	}
	return _users as grpc.Client;
}

// ─── Network Manager ──────────────────────────────────────────────────────────

export function getNetworkClient(): grpc.Client {
	if (!_network) {
		// eslint-disable-next-line @typescript-eslint/no-explicit-any
		const pkg = (grpc.loadPackageDefinition(loadProto('network.proto')) as any).network;
		_network = new pkg.NetworkManager(
			process.env.NETWORK_GRPC ?? 'localhost:50154',
			creds
		);
	}
	return _network as grpc.Client;
}

// ─── Promisify helpers ────────────────────────────────────────────────────────
// Both helpers accept either a resolved grpc.Client or a Promise<grpc.Client>
// so async discovery functions (getCuriosClient, getUsersClient) can be passed
// directly without awaiting at each call site.

export function call<T>(client: grpc.Client | Promise<grpc.Client>, method: string, req: object): Promise<T> {
	return Promise.resolve(client).then(
		(c) =>
			new Promise((resolve, reject) => {
				// eslint-disable-next-line @typescript-eslint/no-explicit-any
				(c as any)[method](req, (err: grpc.ServiceError | null, res: T) => {
					if (err) reject(err);
					else resolve(res);
				});
			})
	);
}

/** Collect all messages from a server-streaming RPC into an array. */
export function callStream<T>(client: grpc.Client | Promise<grpc.Client>, method: string, req: object): Promise<T[]> {
	return Promise.resolve(client).then(
		(c) =>
			new Promise((resolve, reject) => {
				// eslint-disable-next-line @typescript-eslint/no-explicit-any
				const stream = (c as any)[method](req) as grpc.ClientReadableStream<T>;
				const results: T[] = [];
				stream.on('data', (msg: T) => results.push(msg));
				stream.on('end', () => resolve(results));
				stream.on('error', (err: grpc.ServiceError) => reject(err));
			})
	);
}

/** Extract a human-readable message from a gRPC error. */
export function grpcMessage(err: unknown): string {
	if (err && typeof err === 'object' && 'details' in err) return (err as { details: string }).details;
	return String(err);
}
