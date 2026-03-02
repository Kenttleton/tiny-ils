default: dev

# Start all services via Docker Compose
dev:
    docker compose up --build

# Start services in the background
dev-bg:
    docker compose up --build -d

# Stop all services
down:
    docker compose down

# Stop and remove volumes (full reset)
down-clean:
    docker compose down -v

# Run database migrations
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
        proto/network.proto

# Install BFF Node.js dependencies
bff-install:
    cd frontend-bff && npm install

# Install frontend Node.js dependencies
frontend-install:
    cd frontend && npm install

# Install all Node.js dependencies
install-js: bff-install frontend-install

# View logs for a specific service (e.g. just logs curios-manager)
logs service:
    docker compose logs -f {{service}}

# Open a psql shell against the dev database
db:
    docker compose exec postgres psql -U tils -d tils
