FROM golang:1.22-alpine AS builder
WORKDIR /src

COPY go.work go.work.sum ./
COPY shared/go.mod shared/go.sum ./shared/
COPY gen/go.mod gen/go.sum ./gen/
COPY network-manager/go.mod network-manager/go.sum ./network-manager/

RUN go work sync

COPY shared/ ./shared/
COPY gen/ ./gen/
COPY network-manager/ ./network-manager/

RUN CGO_ENABLED=0 go build -o /network-manager ./network-manager

FROM alpine:3.19
RUN apk add --no-cache ca-certificates
COPY --from=builder /network-manager /network-manager
EXPOSE 50053
ENTRYPOINT ["/network-manager"]
