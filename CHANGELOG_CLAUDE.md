# Changelog

## Recent Updates

- Bumped the project and Docker builder image to Go 1.26.2.
- Updated dependencies including Ginkgo, MongoDB Go driver, gRPC, `golang.org/x/*`, and generated protobuf support libraries.
- Regenerated protobuf output for the device and registration services.
- Updated default MQTT credentials to Mosquitto ACL-oriented publisher credentials (`api_devices_pub` / `ApiDevicesPassword1!`).

## Infrastructure & Dependencies

- Upgraded to Go 1.26 and `go.mongodb.org/mongo-driver/v2`.
- Replaced plain Alpine runtime image with hardened `dhi.io/alpine-base:3.23`; service now runs as non-root (`nobody`).
- Log directory is configurable via `LOG_FOLDER` environment variable.
- Updated CI/GitHub Actions workflow for compatibility with current runner versions.

## Security

- **MQTT topic injection:** Device UUID handling was hardened; the current implementation rejects values unless they parse as valid UUIDs in gRPC handlers and `mqttclient.SendValues()`.
- **Silent TLS failure:** CA file read errors are now propagated by `newTLSConfig()` instead of being swallowed.
- **Panic on bad TLS config:** Replaced `panic()` calls in MQTT TLS setup with proper error returns.
- **Credentials in logs:** `MONGODB_URL`, `MQTT_USER`, and `MQTT_PASSWORD` are masked (`****`) in startup logs.
- **Sensitive data exposure:** Removed a `fmt.Println` that dumped TLS certificate details and a `fmt.Printf` that logged full MQTT JSON payloads.
- **Unbounded input:** Added a cap of 100 feature values per `SetValues` request to prevent resource exhaustion.
- **Nil dereference in Register:** Added nil check for `in.Feature` before accessing its fields.
- **PII in logs:** Replaced full request struct logging with selective fields, and removed raw owner ID from validation error messages.
- **Invalid profile owner IDs:** `Register` rejects non-ObjectID `profileOwnerId` values with `codes.InvalidArgument`.

## Bug Fixes

- **Context ignored in gRPC handlers:** Handlers were using a cached `context.Background()` for DB operations, discarding deadlines and cancellation; replaced with the per-request `ctx`.
- **Unused context return value:** `Start()` returned a `context.Background()` that callers always discarded; removed from the signature.
- **Unreachable code after Fatalf:** Removed `panic()` calls following `logger.Fatalf()` (which exits the process).
- **Inconsistent timestamps:** `time.Now()` was called multiple times per operation; captured once and reused across all timestamp fields.
- **Success path returned non-nil error variable:** Handlers now explicitly return `nil` on success.
- **Silent MQTT publish timeout:** `WaitTimeout` returning `false` no longer yields a `nil` error; a descriptive timeout error is returned instead.
- **Invalid MQTT publish topic:** `mqttclient.SendValues()` now returns an error for invalid device UUIDs before publishing.
- **Typo:** Renamed variable `updatedStatue` → `updatedStatus`.

## Idiomatic Go & Code Quality

- Removed unused struct fields (`client`, `contextRef`, `ctx`) from `DevicesGrpc` and `RegisterGrpc`, and eliminated a package-level MongoDB client global.
- Replaced `interface{}` with `any` throughout (idiomatic since Go 1.18).
- Simplified redundant `if { return } else { return }` patterns to `if { return } return`.
- Removed stale commented-out code and replaced placeholder GoDoc comments with meaningful descriptions across all production source files.
- Added MQTT client unit tests for valid and invalid device UUID publishing behavior.
