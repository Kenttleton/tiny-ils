# tiny-ils

A lightweight, federated Integrated Library System for home and community libraries.

Each node is a self-contained library instance that can peer with other nodes to share catalogs and support cross-node borrowing — similar in spirit to cross-chain crypto networks.

## What is a "curio"?

From the root word *curiosity*: a curio is an object of interest. The term is used to generically classify every item in your library regardless of type — books, games, videos, physical tools, digital assets, or anything else.

## Architecture

```text
Browser → SvelteKit frontend (3000)
               ↓ REST
         Node.js BFF (3001)  ← OAuth / SSO / session management
          ↙       ↓       ↘
 curios-manager  users-manager  network-manager
    (50051)        (50052)         (50053)
       ↓               ↓               ↓
                 PostgreSQL
```

All Go microservices expose **gRPC only**. The Node.js BFF translates REST ↔ gRPC and handles authentication. Nodes communicate with each other via the `network-manager` gRPC service.

## Quick start

### Prerequisites

- Docker + Docker Compose v2
- [just](https://github.com/casey/just) (`brew install just` / `cargo install just`)

### 1. Clone and configure

```bash
git clone https://github.com/your-org/tiny-ils.git
cd tiny-ils
cp .env.example .env
# Edit .env — set BOOTSTRAP_MANAGER_EMAIL at minimum
```

### 2. Start everything

```bash
just dev
```

This builds all services and starts them. On first run:

- PostgreSQL runs migrations automatically via the `migrate` service.
- `users-manager` checks `BOOTSTRAP_MANAGER_EMAIL` — if set, it creates that user (or promotes an existing one) as a **MANAGER** for this node.

### 3. Open the app

- Frontend: <http://localhost:3000>
- BFF health: <http://localhost:3001/health>

Log in with the bootstrap manager email to access the admin panel.

---

## Environment variables

Copy `.env.example` to `.env` and adjust:

| Variable                   | Required  | Description                                            |
| -------------------------- | --------- | ------------------------------------------------------ |
| `BOOTSTRAP_MANAGER_EMAIL`  | First run | Email to promote to MANAGER on startup                 |
| `SESSION_SECRET`           | Yes       | Secret for signing session cookies (change in prod)    |
| `GOOGLE_CLIENT_ID`         | Optional  | Google OAuth client ID for SSO                         |
| `GOOGLE_CLIENT_SECRET`     | Optional  | Google OAuth client secret                             |
| `TMDB_API_KEY`             | Optional  | TMDB key for video metadata enrichment                 |
| `IGDB_CLIENT_ID`           | Optional  | Twitch/IGDB client ID for game metadata                |
| `IGDB_CLIENT_SECRET`       | Optional  | Twitch/IGDB client secret                              |

---

## Justfile targets

```sh
just dev             # docker compose up --build (foreground)
just dev-bg          # docker compose up --build (background)
just down            # stop containers
just down-clean      # stop and remove volumes (wipes database)
just build           # compile all Go services locally
just test            # run Go tests
just db-migrate      # run migrations against local DB
just proto           # regenerate protobuf stubs
just logs            # follow container logs
just db              # open psql shell in postgres container
```

---

## Federation / peering

Each node generates an **Ed25519 keypair** on first start (stored in the `node_identity` Docker volume). The public key fingerprint is the node's network identity.

To connect two nodes:

1. Exchange public keys and addresses out-of-band.
2. In each node's admin UI, go to **Admin → Peers** and register the other node's `nodeId`, `publicKey`, and `address` (`host:50053`).

Once peered, users can search across nodes via **Browse** and initiate cross-node borrow requests. User JWTs signed by their home node are verified by the remote node using the exchanged public key.

---

## Role model

| Role    | Capabilities                                                                                    |
| ------- | ----------------------------------------------------------------------------------------------- |
| USER    | Browse catalog, check out copies, place holds                                                   |
| MANAGER | All user capabilities + admin panel (curio CRUD, loan management, peer registration, claims)    |

Claims are **node-scoped**: a person can be MANAGER on Node A and USER on Node B. Managers need given the MANAGER role for each node they are allowed to manage.

---

## Metadata enrichment

When creating a curio, the admin form can auto-populate title, description, tags, and authors from free/open APIs:

| Media type | API           | API key required                              |
| ---------- | ------------- | --------------------------------------------- |
| BOOK       | Open Library  | No                                            |
| AUDIO      | MusicBrainz   | No                                            |
| VIDEO      | TMDB          | Yes (`TMDB_API_KEY`)                          |
| GAME       | IGDB (Twitch) | Yes (`IGDB_CLIENT_ID` / `IGDB_CLIENT_SECRET`) |
| THING      | Manual only   | —                                             |

---

## Services

| Service           | Port       | Description                                                    |
| ----------------- | ---------- | -------------------------------------------------------------- |
| `curios-manager`  | 50051 gRPC | Catalog CRUD, physical copies, loans, holds, digital leases    |
| `users-manager`   | 50052 gRPC | Auth, JWT issuance, RBAC claims, SSO upsert                    |
| `network-manager` | 50053 gRPC | Peer registry, federated search, cross-node borrow             |
| `frontend-bff`    | 3001 REST  | BFF: translates REST → gRPC, manages sessions, OAuth flows     |
| `frontend`        | 3000 HTTP  | SvelteKit UI                                                   |
| `postgres`        | 5432       | Shared database (can be split per-service)                     |

---

## Digital leasing (stub)

The data model and gRPC RPCs for digital leasing are in place (`IssueLease`, `RevokeLease`, `GetDigitalAsset`). Access token delivery is intentionally stubbed — communities can plug in their own file server or DRM system. A future module will provide a reference implementation.

---

## Technologies

- **Go** microservices with gRPC + protobuf
- **Node.js / Express** BFF with `openid-client` (OIDC/OAuth2), `express-session`
- **SvelteKit** frontend with `@sveltejs/adapter-node`
- **PostgreSQL** (pgx driver)
- **Docker Compose** + **Justfile** for local dev
