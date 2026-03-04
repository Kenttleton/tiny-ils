FROM golang:1.25-alpine AS builder
WORKDIR /src

# go work sync requires every module listed in go.work to have its go.mod present
COPY go.work go.work.sum ./
COPY shared/go.mod shared/go.sum ./shared/
COPY gen/go.mod gen/go.sum ./gen/
COPY curios-manager/go.mod curios-manager/go.sum ./curios-manager/
COPY users-manager/go.mod users-manager/go.sum ./users-manager/
COPY network-manager/go.mod network-manager/go.sum ./network-manager/
COPY tools/migrate/go.mod tools/migrate/go.sum ./tools/migrate/

RUN go work sync

COPY shared/ ./shared/
COPY gen/ ./gen/
COPY network-manager/ ./network-manager/

RUN CGO_ENABLED=0 go build -o /network-manager ./network-manager
RUN CGO_ENABLED=0 GOWORK=off go install github.com/grpc-ecosystem/grpc-health-probe@latest

FROM alpine:3.19
RUN apk add --no-cache ca-certificates
COPY --from=builder /network-manager /network-manager
COPY --from=builder /go/bin/grpc-health-probe /grpc-health-probe
EXPOSE 50153
ENTRYPOINT ["/network-manager"]
