# syntax=docker/dockerfile:1
FROM golang:1.23-alpine AS builder
RUN apk update && apk add --no-cache \
    protoc \
    make gcc musl-dev

# install protoc requirements based on https://grpc.io/docs/languages/go/quickstart/
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.34.2
RUN go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.5.1
ENV PATH "$PATH:$(go env GOPATH)/bin"

WORKDIR /app
# pre-copy/cache go.mod for pre-downloading dependencies and only redownloading them in subsequent builds if they change
COPY go.mod ./
COPY go.sum ./
RUN go mod download && go mod verify

COPY . ./

RUN make deps

RUN make build

FROM golang:1.23-alpine
WORKDIR /
COPY --from=builder /app/build/api-devices /api-devices
COPY --from=builder /app/.env_template /.env

# add grpc health probe to support readiness and liveness probes
RUN GRPC_HEALTH_PROBE_VERSION=v0.4.13 && \
    wget -qO/bin/grpc_health_probe https://github.com/grpc-ecosystem/grpc-health-probe/releases/download/${GRPC_HEALTH_PROBE_VERSION}/grpc_health_probe-linux-amd64 && \
    chmod +x /bin/grpc_health_probe

ENTRYPOINT ["/api-devices"]
