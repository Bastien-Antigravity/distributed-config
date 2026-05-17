---
microservice: distributed-config
type: documentation
status: active
tags:
- '#ai/ignore'
- '#zone/3-fleet'
- '#type/documentation'
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
To ensure 100% lock-free reads, the core configuration uses an Atomic Pointer Swap mechanism.
-   **Implementation**: `src/core/config.go` uses `atomic.Pointer[LiveConfig]`.
-   **Benefit**: High-concurrency applications can read configuration without any lock contention, even during background updates.

## Major Components

### Facade (`src/facade`)
The primary entry point for users. It coordinates the lifecycle of a `Config` instance, managing the active strategy and background sync loops.

### Network (`src/network`)
Provides a resilient client built on **safe-socket v0.0.1**. It handles:
-   **Stability**: Uses a 30s deadline to prevent connection loops on local interfaces.
-   **Watch**: Persistent connection for hot-reloads.
-   **Resilience**: Automatic reconnection with jittered exponential backoff.

### Security (`src/secret`)
Integrates RSA decryption. Values prefixed with `ENC(...)` are automatically decrypted at runtime if a valid private key is provided.

## Polyglot Support (FFI Bridge)
The project exports its Go-based logic to other languages through a C-compatible ABI.
- **CGO Bridge**: `src/cgo_bridge/` manages session handles and memory safety between Go and C.
- **Shared Library**: `cmd/libdistconf/main.go` exports the C symbols (e.g., `DistConf_New`, `DistConf_Get`).
- **Bindings**: Located in `distconf/`, including:
    - **C++**: `distconf/cpp/`
    - **Python**: `distconf/python/`
    - **Rust**: `distconf/rust/`
    - **VBA**: `distconf/vba/`
