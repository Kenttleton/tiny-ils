FROM golang:1.22-alpine AS builder
WORKDIR /src

# Copy workspace and module files first for layer caching
COPY go.work go.work.sum ./
COPY shared/go.mod shared/go.sum ./shared/
COPY gen/go.mod gen/go.sum ./gen/
COPY curios-manager/go.mod curios-manager/go.sum ./curios-manager/

RUN go work sync

# Copy source
COPY shared/ ./shared/
COPY gen/ ./gen/
COPY curios-manager/ ./curios-manager/

RUN CGO_ENABLED=0 go build -o /curios-manager ./curios-manager

FROM alpine:3.19
RUN apk add --no-cache ca-certificates
COPY --from=builder /curios-manager /curios-manager
EXPOSE 50051
ENTRYPOINT ["/curios-manager"]
