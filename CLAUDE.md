# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**api-devices** is a gRPC microservice in the home-anthill home automation platform. It manages device registration and control, using MongoDB for persistence and MQTT for device communication. Written in Go 1.26.2.

## Build & Development Commands

**First time setup:**
```bash
make deps           # Install dev tools (shadow, staticcheck, air, go-cover-treemap) and update modules
```

**Common commands:**
```bash
make build          # Generate protobuf, vet, lint, then compile to ./build/api-devices
make run            # Generate protobuf, vet, lint, then start with air (hot-reload)
make test           # Generate protobuf, vet, lint, run tests with race detection and coverage (sets ENV=testing)
make proto          # Regenerate Go code from .proto files
make vet            # Run go vet + shadow analysis
make lint           # Run staticcheck
make check          # Run govulncheck to find vulnerabilities
```

**Run a single test:**
```bash
ENV=testing go test -v -count=1 -run "TestName" ./...
```

**Test framework:** Ginkgo v2 (BDD) + Gomega for integration tests; standard Go tests are used where simpler unit coverage is sufficient. Integration tests require a running MongoDB replica set and use `controllers-test` when `ENV=testing`. Coverage is generated to `./coverage/` (HTML + SVG treemap).

## Architecture

### Startup Sequence

1. **Initialization** (`initialization/start.go`): `Start()` orchestrates:
   - Logger initialization (Zap with file rotation to `LOG_FOLDER`)
   - Environment variable loading from `.env`
   - Database connection to MongoDB
   - gRPC server setup with TLS (if enabled) and health check service
   - Returns `(logger, server, listener, mongoClient)` for cleanup in `main.go`

2. **Entry point** (`main.go`):
   - Calls `initialization.Start()`
   - Initializes MQTT client with TLS support (if configured)
   - Starts gRPC server in a goroutine
   - Blocks on signal channel (SIGINT/SIGTERM) for graceful shutdown
   - On shutdown: stops gRPC server, disconnects MongoDB with 10s timeout

### gRPC Services

Defined in `api/` with protobuf specs in `api/{device,register}/`:

- **Device** (`devicesGrpc.go`): Implements two RPCs:
  - `GetValue`: Reads device feature value from MongoDB, returns with timestamps
  - `SetValues`: Loads authorized controller documents, signs requested feature values, publishes them to MQTT topic `devices/{uuid}/values`, then updates MongoDB only after the publish token reports success. Capped at 100 values per request to prevent resource exhaustion.
  - Both RPCs reject malformed device UUIDs with `codes.InvalidArgument`.

- **Registration** (`registerGrpc.go`): Implements one RPC:
  - `Register`: Upserts device/controller documents in MongoDB with profile, device info, features, and status
  - Rejects missing `Feature`, malformed device UUIDs, and invalid `profileOwnerId` ObjectIDs with `codes.InvalidArgument`

- **Health** (`grpc_health_v1`): Standard gRPC health check probe. Registered in `server.go`

**Handler patterns:** All gRPC handlers follow constructor-based DI:
- `NewDevicesGrpc(logger, mongoClient)` and `NewRegisterGrpc(logger, mongoClient)` inject MongoDB collection and logger
- Handlers use per-request `ctx` parameter (not cached) to respect deadlines and cancellation
- Validation failures return `codes.InvalidArgument` or `codes.NotFound` as appropriate

### Data Layer

**MongoDB** (`db/`): v2 driver.
- Database: `controllers` (prod) or `controllers-test` (ENV=testing)
- Collection: `controllers` (single collection for all device documents)
- Replica set required in CI for transaction support

### Models

`models/` defines:
- **Controller**: Device document with profile info (`apiTokenHash`, `apiTokenEncrypted`, `profileOwnerId`), device info (UUID, MAC), feature details, and status (value with created/modified timestamps)
- **MqttFeatureValue**: MQTT payload structure for publishing device value changes

### MQTT Client

`mqttclient/` implements a global client with:
- TLS support (mTLS with CA cert, client cert/key)
- Publishes to topic `devices/{uuid}/values` at QoS 0
- `SetValues` treats MongoDB status as the last successfully published command state; failed MQTT publishes must not mark a controller status as applied.
- **Security:** Device UUIDs are validated with `github.com/google/uuid` before interpolation to prevent MQTT topic injection
- `InitMqtt()` propagates TLS configuration errors instead of panicking
- Client access is expected to use Mosquitto ACL-oriented publisher credentials (`api_devices_pub` by default)
- QoS 0 (at-most-once delivery) chosen for real-time sensor data where occasional loss is acceptable

## Docker

Multi-stage Dockerfile:
- **Builder stage**: Go 1.26.2 Alpine container, installs protoc tooling, verifies modules, and compiles the binary
- **Runtime stage**: Hardened Alpine base image (`dhi.io/alpine-base:3.23`)
  - Runs as non-root (UID 65534/nobody)
  - Pre-creates `/logs` directory owned by `nobody`
  - Sets `ENV=prod` and `LOG_FOLDER=/logs/` as defaults
  - Published to Docker Hub as `ks89/api-devices:{tag}`

## Testing

**Setup**: MongoDB replica set must be running for integration tests that exercise MongoDB behavior.

**Test structure:**
- **Repository tests** run with `make test`: generates protobuf, runs vet, shadow, staticcheck, then `ENV=testing go test -v -race -count=1 -coverpkg ./... -coverprofile ./coverage/profile.cov ./...`
- **Integration tests** in `integration_tests/`: Connect to real MongoDB (replica set required for transaction-capable CI setups)
  - Use `controllers-test` database (auto-created on first run)
  - Ginkgo v2 (BDD) test framework with Gomega matchers
  - Test suites in `*_test.go` files; main suite in `tests_suite_test.go`
- **MQTT unit tests** in `mqttclient/`: Validate UUID rejection/acceptance with the mock MQTT client
- **MQTT mocking** via `testutils/mqtt_client_mock.go`: Fully mocked, no broker required
- **DB helpers** in `testutils/db_utils.go`: Drop collections, find documents, insert test data

**Coverage**: `make test` generates:
- HTML report: `./coverage/cover.html`
- SVG treemap: `./coverage/out.svg`
- Coverage profile: `./coverage/profile.cov`

## Environment Configuration

Copy `.env_template` to `.env` (gitignored). Key variables:

| Variable | Purpose | Example |
|----------|---------|---------|
| `ENV` | Execution mode | `development`, `testing`, `prod` |
| `LOG_FOLDER` | Log output directory | `./logs/` or `/logs/` (Docker) |
| `MONGODB_URL` | Database connection | `mongodb://localhost:27017/` |
| `GRPC_URL` | gRPC listen address | `:50051` |
| `GRPC_TLS` | Enable gRPC TLS | `true`/`false` |
| `CERT_FOLDER_PATH` | TLS certificate folder (if GRPC_TLS=true) | `cert/` |
| `MQTT_URL` | MQTT broker address | `tcp://localhost` |
| `MQTT_PORT` | MQTT broker port | `1883` |
| `MQTT_TLS` | Enable MQTT TLS | `true`/`false` |
| `MQTT_CA_FILE` | MQTT CA certificate (if MQTT_TLS=true) | `cert/ca.crt` |
| `MQTT_CERT_FILE` | MQTT client certificate (if MQTT_TLS=true) | `cert/client.crt` |
| `MQTT_KEY_FILE` | MQTT client key (if MQTT_TLS=true) | `cert/client.key` |
| `MQTT_CLIENT_ID` | MQTT client identifier | `apiDevices` |
| `MQTT_AUTH` | Enable MQTT authentication | `true`/`false` |
| `MQTT_USER` | MQTT username (if MQTT_AUTH=true) | `api_devices_pub` |
| `MQTT_PASSWORD` | MQTT password (if MQTT_AUTH=true) | `ApiDevicesPassword1!` |
| `API_TOKEN_HASH_SECRET` | Mandatory HMAC secret/pepper for controller API token lookup hashes | `change-me` |
| `API_TOKEN_ENCRYPTION_KEY` | Mandatory 32-byte raw or base64/base64url AES-GCM key for encrypted API token storage | `0123456789abcdef0123456789abcdef` |

## Code Patterns and Conventions

**Go style** (per parent CLAUDE.md):
- Use tabs for indentation
- Never discard errors with `_`
- Constructor-based DI: `NewDevicesGrpc(logger, client)` and `NewRegisterGrpc(logger, client)`
- Zap logging with structured fields via `logger.Infof()`, `logger.Errorf()`, etc.
- Use per-request context in gRPC handlers (never cache `context.Background()`)

**Security patterns:**
- **UUID validation**: Device UUIDs are parsed with `github.com/google/uuid` in gRPC handlers and MQTT publishing before topic interpolation
- **Request validation**: gRPC handlers validate apiToken, deviceUuid, mac fields with FindOne queries (encodes auth check in DB lookup)
- **API token storage**: Raw profile API tokens are not persisted in controller documents. `apiTokenHash` is used for lookup, and `apiTokenEncrypted` is decrypted only when signing MQTT commands for a controller.
- **Credential masking**: MONGODB_URL, MQTT_USER, MQTT_PASSWORD are masked in startup logs
- **TLS error propagation**: TLS misconfiguration returns errors instead of panicking

**Concurrency:**
- `SetValues` processes requested feature values sequentially so authorization/signing errors stop the command batch before MQTT publish
- Max 100 feature values per request to prevent resource exhaustion
- First validation, signing, MQTT, or MongoDB error is returned to the caller

**Protobuf conventions:**
- Proto files: `api/device/device.proto` and `api/register/register.proto`
- Generated files: `*.pb.go` and `*_grpc.pb.go` — **do not edit these directly**
- Regenerate with `make proto` after modifying `.proto` files
