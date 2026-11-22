---
microservice: distributed-config
type: documentation
status: active
tags:
- '#ai/ignore'
- '#zone/3-fleet'
- '#type/documentation'
- '#service/distributed-config'
- '#state/active'
---
# Architecture Overview

The `distributed-config` project is a robust, strategy-based configuration management library designed for high-performance, polyglot environments. It emphasizes thread-safety, real-time updates (hot-reloading), and environment-specific behaviors.

## Core Design Principles

### Strategy-Based Configuration
The system uses a "Strategy" pattern to handle different execution environments.
-   **CloudStrategy**: Unified logic for `production` and `staging`.
-   **StandaloneStrategy**: File-only mode.
-   **TestStrategy**: Hardcoded `127.0.0.2` enforcement.
-   **Location**: `src/strategies/`

### Layered Configuration Loading
Configuration is resolved through a deterministic hierarchy:
1.  **Defaults**: Hardcoded in `src/core/defaults.go`.
2.  **File Discovery**: Local YAML files (e.g., `standalone.yaml`).
3.  **Environment Expansion**: Dynamic replacement of `${VAR}` tokens in YAML.
4.  **Remote Sync**: (CloudStrategy) Overlays from a remote Config Server.

### Read-Copy-Update (RCU) for Performance
To ensure 100% lock-free reads, the core configuration uses an Atomic Pointer Swap mechanism paired with a writer lock (`writeMu`).
-   **Implementation**: `src/core/config.go` uses `atomic.Pointer[LiveConfig]` for reads (`Get`, `GetCapability`, `GetAddress`) and `writeMu sync.Mutex` to serialize `Set()`, `Delete()`, `Apply()`, and `ReplaceLive()` mutations.
-   **Benefit**: High-concurrency applications can read configuration without any lock contention, while concurrent writers and background full refreshes never overwrite each other.
-   **Dynamic Capability Resolution**: `GetAddress()`, `GetGRPCAddress()`, and `GetRESTAddress()` resolve capabilities added dynamically via `LiveConfig` even when absent from static `Capabilities`.

### Optimistic Concurrency Control (OCC)
For sensitive read-modify-write workflows:
-   **Per-Request Opt-in**: Call `cfg.SetWithOCC(updates, expectedVersion)` or `client.UpdateConfigMapWithOCC()`.
-   **Conflict Handling**: Returns `ErrVersionConflict` if the server's monotonic version evolved.
-   **Version Discovery**: `cfg.GetVersion()` exposes the current known server revision.

### Clean Deletion & Convergence
Deletions are first-class operations:
-   **API**: `cfg.DeleteKey(section, key)`, `cfg.DeleteKeys(deletions)`, and `cfg.DeleteKeysWithOCC(deletions, expectedVersion)`.
-   **Sentinel Support**: Delta maps containing `"__DELETED__"` sentinel are cleaned up in `PreviewSet`.
-   **Convergence**: `DELETE_SYNC` triggers an elevated `FULL_REFRESH` from `config-server`, swapping the client's in-memory pointer to purge keys completely without tombstones.

## Major Components

### Facade (`src/facade`)
The primary entry point for users. It coordinates the lifecycle of a `Config` instance, managing the active strategy and background sync loops.

### Network (`src/network`)
Provides a resilient client built on **safe-socket v0.0.1**. It handles:
-   **Stability**: Uses a 30s deadline to prevent connection loops on local interfaces.
-   **Watch**: Persistent connection for hot-reloads.
-   **Resilience**: Automatic reconnection with jittered exponential backoff.
-   **Sequence Gap Self-Healing & Monotonicity**: Tracks `lastSeenVersion uint64`. Stale/duplicate broadcasts (`version <= lastSeenVersion`) and delayed full-refresh snapshots (`version < lastSeenVersion`) are rejected to prevent rollback. If a sequence gap is detected (`received > lastSeen + 1`), it automatically triggers an out-of-band `FullRefresh()` to guarantee parity.

### Security (`src/secret`)
Integrates RSA-OAEP decryption with in-memory private key memoization (keyed by `BASTIEN_PRIVATE_KEY` SHA-256 hash or key file path + modification time + size). Values prefixed with `ENC(...)` remain ciphertext at load time and are decrypted on-demand via `Decrypt()` / `DecryptSecret()`, while `ProcessConfigSecrets()` fails fast if any `ENC(...)` token cannot be decrypted.

## Polyglot Support (FFI Bridge)
The project exports its Go-based logic to other languages through a C-compatible ABI.
- **CGO Bridge**: `src/cgo_bridge/` manages session handles and memory safety between Go and C.
- **Shared Library**: `cmd/libdistconf/main.go` exports the C symbols (e.g., `DistConf_New`, `DistConf_Get`).
- **Bindings**: Located in `distconf/`, including:
    - **C++**: `distconf/cpp/`
    - **Python**: `distconf/python/`
    - **Rust**: `distconf/rust/`
    - **VBA**: `distconf/vba/`
