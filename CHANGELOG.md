# Changelog


## 4.1.0

### Features

- Added thermostat `mode` sensor feature
- Added `DeleteValue` gRPC cleanup for removing controller documents by device UUID and feature UUID.

### Tests

- Added device gRPC unit tests for command signing, missing controller API tokens, invalid encrypted tokens, invalid device UUIDs, missing hash secrets, and oversized `SetValues` batches.
- Added registration gRPC unit tests for missing features, invalid device UUIDs, invalid profile owner IDs, and missing API token secrets.
- Added API token crypto tests for hashing stability, encryption/decryption round trips, missing or invalid encryption keys, base64 key formats, and invalid ciphertext handling.
- Added initialization tests for required API token environment variables, logger creation, and log writer setup.
- Added database tests for production vs. testing database name selection and controller collection wiring.
- Added MQTT configuration tests for TLS setup failures, missing or invalid CA files, and non-TLS client initialization.


## 4.0.0

### Bug Fixes

- **Context ignored in gRPC handlers:** Handlers were using a cached `context.Background()` for DB operations, discarding deadlines and cancellation; replaced with the per-request `ctx`.
- **Unused context return value:** `Start()` returned a `context.Background()` that callers always discarded; removed from the signature.
- **Unreachable code after Fatalf:** Removed `panic()` calls following `logger.Fatalf()` (which exits the process).
- **Inconsistent timestamps:** `time.Now()` was called multiple times per operation; captured once and reused across all timestamp fields.
- **Success path returned non-nil error variable:** Handlers now explicitly return `nil` on success.
- **Silent MQTT publish timeout:** `WaitTimeout` returning `false` no longer yields a `nil` error; a descriptive timeout error is returned instead.
- **Invalid MQTT publish topic:** `mqttclient.SendValues()` now returns an error for invalid device UUIDs before publishing.
- **Typo:** Renamed variable `updatedStatue` → `updatedStatus`.

### Security

- **Controller state integrity:** `SetValues` now publishes the signed MQTT command batch before updating MongoDB status, so a failed publish no longer marks a controller state as applied.
- **MQTT publisher credentials:** Updated default MQTT credentials to Mosquitto ACL-oriented publisher credentials (`api_devices_pub` / `ApiDevicesPassword1!`).
- **MQTT topic injection:** Device UUID handling was hardened; the current implementation rejects values unless they parse as valid UUIDs in gRPC handlers and `mqttclient.SendValues()`.
- **Silent TLS failure:** CA file read errors are now propagated by `newTLSConfig()` instead of being swallowed.
- **Panic on bad TLS config:** Replaced `panic()` calls in MQTT TLS setup with proper error returns.
- **Credentials in logs:** `MONGODB_URL`, `MQTT_USER`, and `MQTT_PASSWORD` are masked (`****`) in startup logs.
- **Controller API token storage:** Controller documents now store `apiTokenHash` plus AES-GCM `apiTokenEncrypted` instead of plaintext `apiToken`; command signing decrypts only when needed and requires `API_TOKEN_HASH_SECRET` / `API_TOKEN_ENCRYPTION_KEY`.
- **Token secret startup validation:** Startup now fails if `API_TOKEN_HASH_SECRET` is missing or shorter than 32 characters, or if `API_TOKEN_ENCRYPTION_KEY` is missing.
- **Sensitive data exposure:** Removed a `fmt.Println` that dumped TLS certificate details and a `fmt.Printf` that logged full MQTT JSON payloads.
- **Unbounded input:** Added a cap of 100 feature values per `SetValues` request to prevent resource exhaustion.
- **Nil dereference in Register:** Added nil check for `in.Feature` before accessing its fields.
- **PII in logs:** Replaced full request struct logging with selective fields, and removed raw owner ID from validation error messages.
- **Invalid profile owner IDs:** `Register` rejects non-ObjectID `profileOwnerId` values with `codes.InvalidArgument`.

### Idiomatic Go & Code Quality

- Removed unused struct fields (`client`, `contextRef`, `ctx`) from `DevicesGrpc` and `RegisterGrpc`, and eliminated a package-level MongoDB client global.
- Replaced `interface{}` with `any` throughout (idiomatic since Go 1.18).
- Simplified redundant `if { return } else { return }` patterns to `if { return } return`.
- Removed stale commented-out code and replaced placeholder GoDoc comments with meaningful descriptions across all production source files.
- Added MQTT client unit tests for valid and invalid device UUID publishing behavior.

### Chores

- Upgraded to Go 1.26 and `go.mongodb.org/mongo-driver/v2`.
- Bumped the project and Docker builder image to Go 1.26.3.
- Updated dependencies including Ginkgo, MongoDB Go driver, gRPC, `golang.org/x/*`, and generated protobuf support libraries.
- Regenerated protobuf output for the device and registration services.
- Replaced plain Alpine runtime image with hardened `dhi.io/alpine-base:3.23`; service now runs as non-root (`nobody`).
- Log directory is configurable via `LOG_FOLDER` environment variable.
- Updated CI/GitHub Actions workflow for compatibility with current runner versions.
