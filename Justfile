default: dev

# Start all services via Docker Compose
dev:
    docker compose up --build

# Start services in the background
dev-bg:
    docker compose up --build -d

# Start all services including Readium LCP/LSD (requires LCP images — see .env.lcp.example)
dev-lcp:
    docker compose --profile lcp up --build

# Stop all services
down:
    docker compose down

# Stop and remove volumes (full reset)
down-clean:
    docker compose down -v

# Apply schema files against DATABASE_URL (production / external DB use).
# In development, schema is applied automatically via postgres's initdb mechanism.
db-migrate:
    go run ./tools/migrate

# Build all Go services
build:
    go build ./curios-manager/...
    go build ./users-manager/...
    go build ./network-manager/...

# Run all Go tests
test:
    go test ./...

# Generate Go code from proto files into gen/
# Requires: brew install protobuf protoc-gen-go protoc-gen-go-grpc
proto:
    protoc \
        --go_out=./gen --go_opt=module=tiny-ils/gen \
        --go-grpc_out=./gen --go-grpc_opt=module=tiny-ils/gen \
        proto/curios.proto \
        proto/users.proto \
        proto/network.proto \
        proto/internal.proto

# Install frontend Node.js dependencies
frontend-install:
    cd frontend && npm install

# View logs for a specific service (e.g. just logs curios-manager)
logs service:
    docker compose logs -f {{service}}

# Open a psql shell against the dev database
db:
    docker compose exec postgres psql -U tils -d tils

# ─── Two-node testing ─────────────────────────────────────────────────────────

# Start Node B peer cluster (Node A is the default `just dev`)
node-b:
    docker compose -f docker-compose.node-b.yml up --build

# Start Node B in the background
node-b-bg:
    docker compose -f docker-compose.node-b.yml up --build -d

# Stop Node B (simulate peer going offline)
node-b-down:
    docker compose -f docker-compose.node-b.yml down

# Full reset of Node B (remove volumes)
node-b-clean:
    docker compose -f docker-compose.node-b.yml down -v

# Seed Node A with test curios via gRPC
seed:
    CURIOS_GRPC=localhost:50151 go run ./tools/seed

# Seed Node B with test curios via gRPC
seed-b:
    CURIOS_GRPC=localhost:50161 go run ./tools/seed
