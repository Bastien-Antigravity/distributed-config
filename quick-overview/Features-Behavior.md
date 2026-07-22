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

## 1. CloudStrategy (Production & Staging)

The `CloudStrategy` (in `src/strategies/cloud.go`) is the primary mechanism for microservices connecting to the central `config-server`.

### Consolidation Logic
To adhere to the **Purger Rule**, the once separate `ProductionStrategy` and `StagingStrategy` have been merged into this unified strategy. It uses a `ReadOnly` toggle and a `Profile` string to differentiate behavior.

### Loading Hierarchy
1.  **Local Discovery**: Tries to find `config/[profile].yaml`.
2.  **Environment Expansion**: Injects `${VAR}` tokens (e.g., `CF_IP`, `CF_PORT`).
3.  **Server Baseline**: Connects to the Config Server and fetches the global registry.
4.  **Local Override (Authoritative)**: Re-reads the local file. **The local file always wins** over server data to allow per-node fine-tuning.

### Profile-Specific Integrity
-   **Production**: Enforces `CheckProductionIPs`. It will return an error (preventing boot) if any capability IP is detected as `127.0.0.2` (the test IP).
-   **Staging**: Read-Only mode by default. Synchronization (`Sync`) and remote `Set` requests are disabled to prevent staging environments from polluting the central config store.

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
