FROM golang:1.22-alpine AS builder
WORKDIR /src

COPY go.work go.work.sum ./
COPY shared/go.mod shared/go.sum ./shared/
COPY gen/go.mod gen/go.sum ./gen/
COPY users-manager/go.mod users-manager/go.sum ./users-manager/

RUN go work sync

COPY shared/ ./shared/
COPY gen/ ./gen/
COPY users-manager/ ./users-manager/

RUN CGO_ENABLED=0 go build -o /users-manager ./users-manager

FROM alpine:3.19
RUN apk add --no-cache ca-certificates
COPY --from=builder /users-manager /users-manager
EXPOSE 50052
ENTRYPOINT ["/users-manager"]
