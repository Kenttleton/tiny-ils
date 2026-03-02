import path from 'path';
import * as grpc from '@grpc/grpc-js';
import * as protoLoader from '@grpc/proto-loader';

const PROTO_DIR = path.resolve(__dirname, '../../../proto');

function loadProto(filename: string) {
  return protoLoader.loadSync(path.join(PROTO_DIR, filename), {
    keepCase: true,
    longs: String,
    enums: String,
    defaults: true,
    oneofs: true,
  });
}

function makeChannel(address: string) {
  return grpc.credentials.createInsecure();
}

// ─── Curios Manager ──────────────────────────────────────────────────────────

const curiosDef = loadProto('curios.proto');
// eslint-disable-next-line @typescript-eslint/no-explicit-any
const curiosPkg = (grpc.loadPackageDefinition(curiosDef) as any).curios;

export function getCuriosClient() {
  const addr = process.env.CURIOS_GRPC ?? 'localhost:50051';
  return new curiosPkg.CuriosManager(addr, makeChannel(addr));
}

// ─── Users Manager ───────────────────────────────────────────────────────────

const usersDef = loadProto('users.proto');
// eslint-disable-next-line @typescript-eslint/no-explicit-any
const usersPkg = (grpc.loadPackageDefinition(usersDef) as any).users;

export function getUsersClient() {
  const addr = process.env.USERS_GRPC ?? 'localhost:50052';
  return new usersPkg.UsersManager(addr, makeChannel(addr));
}

// ─── Network Manager ─────────────────────────────────────────────────────────

const networkDef = loadProto('network.proto');
// eslint-disable-next-line @typescript-eslint/no-explicit-any
const networkPkg = (grpc.loadPackageDefinition(networkDef) as any).network;

export function getNetworkClient() {
  const addr = process.env.NETWORK_GRPC ?? 'localhost:50053';
  return new networkPkg.NetworkManager(addr, makeChannel(addr));
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
