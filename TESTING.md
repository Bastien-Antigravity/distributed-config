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
*   **Priority Rules**: Verifies that `config/` subfolders are prioritized over root folders.
*   **Fallback Logic**: Tests that `default.yaml` is used if specific configuration files are missing.

### 2. Loader & Secret Management (`src/loader/loader_test.go`)
Tests the parsing logic and security features.
*   **YAML Mapping**: Ensures YAML structures correctly populate the `core.Config` Go structs.
*   **Environment Expansion**: Verifies that placeholders like `${DB_PASSWORD}` are correctly expanded from environment variables at runtime.
*   **Auto-Generation**: Confirms that missing configuration files are automatically generated with safe defaults.

### 3. Network & Dynamic Updates (`src/network/network_test.go`)
Tests the bidirectional communication with the Config Server.
*   **Protobuf Integrity**: Validates the serialization and deserialization of configuration messages.
*   **Callback System**: Ensures that registering a listener via `cfg.OnMemConfUpdate(...)` correctly triggers when the server pushes new data.
*   **Memory Sync**: Verifies that incoming server data correctly updates the internal `MemConfig` map.

### 4. Safety & Integrity (`src/loader/validator_test.go`)
Critical fail-safe tests to prevent configuration errors in production.
*   **IP Sanity**: Verifies that "Test" IPs (127.0.0.2) are strictly blocked in Production profiles.
*   **Required Fields**: Ensures that the library fails fast if mandatory fields (Application Name, Server details) are missing.

## Writing New Tests

When adding new features, please follow these guidelines:
*   **Isolation**: Use `t.TempDir()` and `t.Setenv()` to avoid polluting the host environment.
*   **Predictability**: Avoid relying on hardcoded absolute paths; instead, use relative path resolution.
*   **Race Conditions**: Always run tests with the `-race` flag for multi-threaded updates.

```bash
go test -race ./src/...
```
