# tiny-ils

A lightweight, federated Integrated Library System for home and community libraries.

Each node is a self-contained library instance that can peer with other nodes to share catalogs and support cross-node borrowing — similar in spirit to cross-chain crypto networks.

## What is a "curio"?

From the root word *curiosity*: a curio is an object of interest. The term is used to generically classify every item in your library regardless of type — books, games, videos, physical tools, digital assets, or anything else.

---

## Architecture

```text
Browser
   │
   ▼
SvelteKit (port 3000)
   │  server-side routes handle sessions, OAuth, and gRPC fan-out
   ├──────────────────┬──────────────────┐
   ▼                  ▼                  ▼
curios-manager    users-manager    network-manager
  (50151)           (50152)           (50153)
   │                  │                  │
   └──────────────────┴──────────────────┘
                       │
                  PostgreSQL (5432)
```

All Go microservices expose **gRPC only**. The SvelteKit frontend acts as the backend-for-frontend (BFF): its server-side routes call the appropriate gRPC services and manage HTTP sessions. There is no separate REST API layer.

Nodes communicate with each other directly via each node's `network-manager` gRPC service.

---

## Services

| Service           | Port         | Description                                                       |
| ----------------- | ------------ | ----------------------------------------------------------------- |
| `frontend`        | 3000         | SvelteKit UI + BFF (sessions, OAuth, gRPC fan-out)                |
| `curios-manager`  | 50151        | Catalog CRUD, physical copies, loans, holds, digital leases       |
| `users-manager`   | 50152        | Auth, JWT issuance, RBAC claims, SSO upsert                       |
| `network-manager` | 50153 (mTLS) | Peer-to-peer gRPC; both remote peers and the local BFF use mTLS   |
| `postgres`        | 5432         | Shared database (all services connect to the same instance)       |

---

## Quick start

### Prerequisites

- Docker + Docker Compose v2
- [just](https://github.com/casey/just) (`brew install just` / `cargo install just`)

### 1. Clone and start

```bash
git clone https://github.com/your-org/tiny-ils.git
cd tiny-ils
just dev
```

This builds all images and starts all services. PostgreSQL migrations run automatically on first start.

### 2. First-run setup

Navigate to <http://localhost:3000/setup> to create the first manager account.

### 3. Optional: Google SSO

Set `GOOGLE_CLIENT_ID`, `GOOGLE_CLIENT_SECRET`, and `GOOGLE_REDIRECT_URI` in your environment (or a `.env` file) before starting.

---

## Environment variables

| Variable                  | Required | Description                                                    |
| ------------------------- | -------- | -------------------------------------------------------------- |
| `GOOGLE_CLIENT_ID`        | Optional | Google OAuth client ID for SSO login                           |
| `GOOGLE_CLIENT_SECRET`    | Optional | Google OAuth client secret                                     |
| `GOOGLE_REDIRECT_URI`     | Optional | OAuth redirect URI for Google SSO                              |
| `LCP_SERVER_URL`          | Optional | Readium LCP server URL for digital rights management           |
| `LSD_SERVER_URL`          | Optional | Readium LSD server URL                                         |
| `LSD_PUBLIC_URL`          | Optional | Public-facing LSD URL for license status documents             |

---

## Justfile targets

```sh
just dev             # docker compose up --build (foreground)
just dev-bg          # docker compose up --build (background)
just down            # stop containers
just down-clean      # stop and remove volumes (wipes database and node identity)
just build           # compile all Go services locally
just test            # run Go tests
just db-migrate      # run migrations against local DB
just proto           # regenerate protobuf stubs from proto/
just logs            # follow container logs
just db              # open psql shell in postgres container
```

---

## Authentication

tiny-ils uses two distinct authentication layers: one for users within a node, and one for node-to-node trust.

### User authentication

On login, `users-manager` signs a short-lived **Ed25519 JWT** using the node's private key. The JWT contains:

- `sub` / `uid` — the user's UUID
- `iss` — the node's fingerprint (SHA-256 of the public key, base64url-encoded)
- `claims` — the user's RBAC roles on this node (e.g. `MANAGER`)

The SvelteKit frontend stores this token in a server-side session. All privileged API actions require a valid session.

### Node identity

Each node generates an **Ed25519 keypair** on first start. The keypair is stored in the `node_identity` Docker volume at `/data/node.key` (private) and `/data/node.pub` (public). The public key fingerprint (first 16 bytes of its SHA-256 hash, base64url-encoded) is the node's permanent **Library ID**.

> **Note:** Running `docker compose down -v` destroys the `node_identity` volume and generates a new keypair on next start. Peer nodes will need to re-register with the new key.

### Node-to-node authentication

Inter-node gRPC uses **mutual TLS (mTLS)** with self-signed Ed25519 certificates. No certificate authority is involved — trust is established through public-key pinning via the peers table.

#### Transport

Each node's Ed25519 key pair is used to generate an in-memory self-signed x.509 certificate on startup. The `network-manager` listens on a **single mTLS port (50153)**. Both remote peer nodes and the local SvelteKit BFF connect on this port — the BFF authenticates using the same `node_identity` key pair, making it a first-class node on the network (similar to how a crypto wallet is also a node).

The generated certificate is written to `/data/node.crt` in the shared `node_identity` volume so the BFF can load it. The BFF presents the node's own cert when connecting; the interceptor recognizes the matching public key and grants it full `TrustConnected` access without a peers-table lookup.

#### Trust levels

Access to inter-node RPCs is tiered by the caller's certificate and registration status:

| Trust level      | Condition                                                   | Allowed RPCs                                                                                      |
| ---------------- | ----------------------------------------------------------- | ------------------------------------------------------------------------------------------------- |
| `TrustNone`      | No mTLS cert presented                                      | `GetNodeInfo`                                                                                     |
| `TrustCert`      | Valid cert, peer not yet CONNECTED (or PENDING)             | `RegisterPeer`, `SearchNetwork`, `ShareCatalog`                                                   |
| `TrustConnected` | Cert matches own node key (BFF), or peer is CONNECTED in DB | All other RPCs: borrow, return, transfers, cross-node auth, digital leases, loan fan-out, admin   |

#### Peer status lifecycle

Each entry in the `peers` table has a `status`:

- **`PENDING`** — the peer has called `RegisterPeer` (or has been seen before approval). They can search the catalog but cannot borrow.
- **`CONNECTED`** — the peer is fully trusted. Either the local admin called `ConnectPeer` (admin-initiated outbound), or the admin approved an inbound `PENDING` peer via `ApprovePeer`.

#### Connection flow

**Admin-initiated (outbound):** Admin enters the remote node's Library ID, public key, and address in **Admin → Network → Connect a partner library**. The local network-manager stores the peer as `CONNECTED` and calls the remote node's `RegisterPeer` so the remote knows about this node (the remote stores this node as `PENDING` until their admin approves).

**Inbound (peer-initiated):** A remote node calls `RegisterPeer` on this node. If this node has no record of them, they are stored as `PENDING`. If this admin had pre-registered them (via `ConnectPeer`), they are upgraded to `CONNECTED` immediately.

**Approving inbound peers:** Admin goes to **Admin → Network** and clicks **Approve** next to any `PENDING` partner library.

#### User JWT verification (for user-scoped operations)

For borrow, return, and transfer operations, the request includes a `user_jwt` signed by the requesting node. The receiving node verifies this JWT against the registered peer's public key — proving the user's home library actually authenticated them.

---

## Federated user identity

tiny-ils treats user identity like a crypto wallet: the user is the global key, not the account. A user's **home node** is their identity authority — it authenticates them and issues credentials — but their loans and digital leases live on whatever node holds the copy, keyed by `(user_id, user_node_id)`.

### How it works

- **User ID** — a stable UUID assigned at registration. Globally unique across all nodes in the network.
- **Home node** — the node where the user first registered. It is the OIDC/JWT authority for that user. The node fingerprint is stored as `iss` in every JWT.
- **Cross-node login** — users can sign in at any partner library via **Sign in → Partner library sign-in**. The local node contacts the user's home node (which must be CONNECTED), gets a guest token, creates a thin local user record, and issues a local session JWT with a `home_node` claim.
- **Loans follow the user** — `GetUserLoans` is a streaming fan-out RPC that queries every CONNECTED peer (and the local node) for loans keyed by `(user_id, home_node_id)`. The **My Loans** page automatically aggregates results across all libraries for cross-node users.
- **Digital leases** — same model: `RequestDigitalLease` and `RevokeDigitalLease` require a home-node-issued JWT; the owning node verifies it against the peer's registered public key.
- **Node-scoped RBAC** — a user's role (USER / MANAGER) is determined independently by each node from its own `node_claims` table. The home-node JWT carries **no role claims**. `GrantDefault` auto-grants USER on first cross-node login without downgrading an existing MANAGER grant.

### Cross-node login flow

```text
User (at Node B's UI)
  → enters: home node address, home node Library ID, user UUID
  → Node B's network-manager calls IssueGuestToken on Node A (CONNECTED)
  → Node A issues an audience-scoped JWT (aud = Node B's fingerprint)
  → Node B calls UpsertGuestUser on its own users-manager
  → users-manager creates a thin local record, GrantDefault (USER), returns session JWT
  → User is logged in at Node B with their home UUID preserved
```

### Key RPCs for federated identity

| RPC                   | Direction       | Description                                                         |
| --------------------- | --------------- | ------------------------------------------------------------------- |
| `IssueGuestToken`     | Node B → Node A | Node A mints an audience-scoped guest JWT for one of its users      |
| `AuthenticateGuest`   | BFF → local nm  | Orchestrates cross-node login; returns local session JWT            |
| `GetUserLoans`        | BFF → local nm  | Streaming fan-out: collects physical + digital loans from all peers |
| `RequestDigitalLease` | Node B → Node A | Lease a digital item; home-node JWT proves user identity            |
| `RevokeDigitalLease`  | Node B → Node A | Revoke a lease; same JWT verification                               |

---

## Node capabilities

Every node advertises what it offers via the `NODE_CAPABILITIES` environment variable (default: `curios,users,ui`). Capabilities are exchanged during peer registration and stored in the `peers` table.

| Capability | Mode   | What it provides                                                    |
| ---------- | ------ | ------------------------------------------------------------------- |
| `curios`   | server | Catalog items, physical copies, loans, holds, digital leases        |
| `users`    | server | Identity, authentication, JWT issuance, RBAC claims                 |
| `network`  | server | Peer management, trust enforcement, routing (always present)        |
| `ui`       | client | Human interface — initiates requests, accepts none from the network |

**Capability-aware routing:** Fan-out operations only target peers that can handle them. `SearchNetwork` and `GetUserLoans` only reach `curios`-capable peers; `IssueGuestToken` only calls `users`-capable peers. Peers with an empty capabilities list (nodes that predate this protocol) are treated as full-stack for backwards compatibility.

**Example deployments:**

```text
Full home library:       curios,users,ui     — complete stack
Headless branch:         curios              — holds inventory, trusts a parent node's users
Digital consortium:      curios,users        — no UI; member libraries connect their own patrons
Pure UI node:            ui                  — custom patron portal, no local data
```

The `NODE_CAPABILITIES` env var is comma-separated. The `network` capability is always present regardless of the setting.

---

## Federation / peering

To connect two tiny-ils nodes:

1. In each node's admin UI, go to **Admin → Network**.
2. Share your **Library ID** and **Public key** with the other library's administrator.
3. Enter their Library ID, public key, and gRPC address (`host:50153`) and click **Connect library**.
4. Your node stores them as `CONNECTED` and notifies their node, which stores you as `PENDING`.
5. Their admin approves your node in their **Admin → Network** panel → **Approve**.

Once both sides have approved, users can search across nodes via **Browse** and initiate cross-node borrow requests.

> **One-sided search:** A node that receives a `RegisterPeer` call immediately allows the calling node to search its catalog (status `PENDING`). Full borrowing access requires explicit admin approval on both sides.

---

## Role model

| Role    | Capabilities                                                                                 |
| ------- | -------------------------------------------------------------------------------------------- |
| USER    | Browse catalog, check out copies, place holds, cross-node borrowing                          |
| MANAGER | All user capabilities + admin panel (catalog CRUD, loan management, peer registry, claims)   |

Claims are **node-scoped**: a person can be MANAGER on Node A and USER on Node B. The MANAGER role must be granted separately on each node.

---

## Metadata enrichment

When creating a curio, the admin form can auto-populate title, description, tags, and authors from external APIs:

| Media type | API           | API key required                              |
| ---------- | ------------- | --------------------------------------------- |
| BOOK       | Open Library  | No                                            |
| AUDIO      | MusicBrainz   | No                                            |
| VIDEO      | TMDB          | Yes (`TMDB_API_KEY`)                          |
| GAME       | IGDB (Twitch) | Yes (`IGDB_CLIENT_ID` / `IGDB_CLIENT_SECRET`) |
| THING      | Manual only   | —                                             |

---

## Digital leasing (stub)

The data model and gRPC RPCs for digital leasing are in place (`IssueLease`, `RevokeLease`, `GetDigitalAsset`). Readium LCP/LSD integration is supported via optional Docker Compose profiles. Access token delivery for non-LCP assets is intentionally left open — communities can plug in their own file server or DRM system.

---

## Technologies

- **Go** microservices with gRPC + protobuf (`google.golang.org/grpc`, `google.golang.org/protobuf`)
- **SvelteKit** frontend with `@sveltejs/adapter-node` (also serves as the BFF)
- **Arctic** for OAuth 2.0 / OIDC flows (Google SSO)
- **Ed25519** for node identity, JWT signing, and node-to-node authentication
- **PostgreSQL** with `pgx` driver
- **Docker Compose** + **Justfile** for local development
- **Readium LCP/LSD** (optional) for digital rights management
