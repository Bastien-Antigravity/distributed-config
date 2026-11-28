---
microservice: distributed-config
type: documentation
status: active
tags:
- '#ai/ignore'
- '#zone/3-fleet'
- '#type/documentation'
- '#domain/configuration'
- '#service/distributed-config'
- '#state/active'
---
# Features & Behavior

Detailed technical breakdown of the behaviors and logic patterns within `distributed-config`.

## 1. CloudStrategy (Production, Staging & Test)

The `CloudStrategy` (in `src/strategies/cloud.go`) is the primary mechanism for microservices connecting to the central `config-server`.

### Consolidation Logic
To adhere to the **Purger Rule**, `ProductionStrategy`, `StagingStrategy`, and `TestStrategy` all share `CloudStrategy` execution logic (`TestStrategy` delegates to `CloudStrategy` with `Profile: "test"` and `RequireFile: true`). It uses `ReadOnly`, `RequireFile`, and `Profile` to differentiate behavior.

### Loading Hierarchy
1.  **Local Discovery (Single Disk Read)**: Resolves the profile YAML (or `standalone.yaml` seed) and reads it from disk once, snapshotting file capabilities in memory.
2.  **Environment Expansion**: Injects `${VAR}` tokens and reserved bootstrap variables (`CF_IP`, `CF_PORT`, `LS_IP`, `LS_PORT`) for Environment-First bootstrapping.
3.  **Server Baseline**: Connects to the Config Server (`GET_SYNC`), starts the `Watch()` stream, and hydrates server capabilities.
4.  **Local Override (Authoritative)**: Re-applies the in-memory file snapshot over server data (unless `ResolveConfigPath` fell back to `standalone.yaml` as a seed while connected to `config-server`).

### Profile-Specific Integrity
-   **Production**: Enforces `CheckProductionIPs` across `ip`, `grpc_ip`, and `rest_ip` endpoints (preventing `127.0.0.2` outside `DOCKER_ENV=true`).
-   **Staging**: Read-Only mode (`ReadOnly: true`). Synchronization (`Sync`), `SetWithOCC`, and deletions are disabled to prevent staging environments from polluting the central config store.
-   **Test**: Requires a local YAML file (`RequireFile: true`) and enforces `CheckTestIPs` across `ip`, `grpc_ip`, and `rest_ip` when `CHECK_TEST_IPS=true`.

## 2. Lock-Free Live Config (RCU)

High-throughput trading microservices cannot afford lock contention on configuration lookups.

-   **Mechanism**: Uses `atomic.Pointer` to store the live configuration map.
-   **Read Pattern**: `cfg.Get()` loads the pointer and performs a map lookup. This is 100% non-blocking.
-   **Update Pattern**: `cfg.Set()` performs a **Read-Copy-Update (RCU)**. It copies the current map, applies updates, and atomically swaps the pointer.

## 3. Polyglot FFI Bridge

The library exports its Go core to C/C++, Python, Rust, and VBA.

-   **Memory Safety**: The bridge ensures that all strings passed to guest languages are allocated in C-memory.
-   **Sanitization**: `src/cgo_bridge/sanitizer.go` automatically trims whitespace and removes null bytes from incoming strings to prevent memory corruption in guest language runtimes.
-   **Handle Isolation**: Each guest language "object" maps to a unique `uintptr` handle in the Go `FacadeStore`, ensuring multiple independent config sessions can coexist.

## 4. Resilience & Networking

-   **SafeSocket Alignment**: Uses `safe-socket v0.0.1`.
-   **Persistent Connections**: The `Watch()` loop maintains a background stream.
-   **Watchdog Timeout**: Configured with a **30-second deadline** to account for high-load scheduling delays while maintaining connection health detection.
-   **Backoff**: Uses jittered exponential backoff for re-establishing server connections.
