# Testing Guide

This document describes the testing strategy and procedures for the Distributed Config library.

## Getting Started

The project uses the standard Go testing toolchain. You can run all tests using the provided `Makefile`:

```bash
make test
```

Or manually using `go test`:

```bash
go test -v ./...
```

## Test Layers

The library is tested across four distinct layers to ensure reliability and safety.

### 1. Configuration Discovery (`src/loader/path_resolver_test.go`)
Ensures that the library can find configuration files in various environments (Development, Installed, Containerized).
*   **Testing Sandbox**: End-to-end integration across microservices should be tested locally inside the `testing-sandbox/` directory using native `docker-compose`.
*   **Priority Rules**: Verifies that `config/` subfolders are strictly prioritized for profile files (e.g., `config/staging.yaml`).
*   **Fallback Logic**: Tests that the system falls back to the executable-name based YAML (e.g., `config/[exe].yaml` then `[exe].yaml`) if profile files are missing.

### 2. Loader & Secret Management (`src/loader/loader_test.go`)
Tests the parsing logic and security features.
*   **YAML Mapping**: Ensures YAML structures correctly populate the `core.Config` Go structs.
*   **Environment Expansion**: Verifies that placeholders like `${DB_PASSWORD}` are correctly expanded from environment variables at runtime.
*   **Auto-Generation**: Confirms that missing configuration files are automatically managed: Test/Standalone generates **Full Defaults**, while Production/Staging **proceeds if ENVs are present** (no file creation).

### 3. Network & Dynamic Updates (`src/network/network_test.go`)
Tests the bidirectional communication with the Config Server.
*   **Protobuf Integrity**: Validates the serialization and deserialization of the underlying raw JSON mapping blobs inside Protobuf payload bytes.
*   **Callback System**: Ensures that registering a listener via `cfg.OnLiveConfUpdate(...)` or `cfg.OnRegistryUpdate(...)` triggers accurately upon config changes.
*   **Live Sync**: Verifies that incoming server data correctly updates the internal `LiveConfig` map.

### 4. Safety & Integrity (`src/loader/validator_test.go`)
Critical fail-safe tests to prevent configuration errors in production.
*   **IP Sanity**: Verifies that "Test" IPs (127.0.0.2) are strictly blocked in Production profiles.
*   **Required Fields**: Ensures that the library fails fast if mandatory services are not satisfied by any source (File or Env). Specifically checks `log_server` and `config_server`.

## Writing New Tests

When adding new features, please follow these guidelines:
*   **Isolation**: Use `t.TempDir()` and `t.Setenv()` to avoid polluting the host environment.
*   **Predictability**: Avoid relying on hardcoded absolute paths; instead, use relative path resolution.
*   **Race Conditions**: Always run tests with the `-race` flag for multi-threaded updates.

```bash
go test -race ./src/...
```
