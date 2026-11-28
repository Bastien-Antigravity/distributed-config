---
tags:
- '#ai/ignore'
- '#zone/3-fleet'
- '#service/distributed-config'
- '#type/overview'
- '#state/active'
microservice: distributed-config
type: overview
status: active
---
# Polyglot SDK Guide

`distributed-config` provides a centralized Go core that is exposed to other languages through a C-compatible ABI (CGO Bridge). This ensures 100% architectural parity across the entire microservice fleet.

## 🏗️ The CGO Bridge

The core logic resides in `src/cgo_bridge/`, which manages:
- **Session Handles**: Thread-safe mapping of integer handles to Go `Config` instances.
- **Memory Safety**: Careful management of string pointers across the FFI boundary.
- **Async Callbacks**: Bridging Go channels/callbacks to native language function pointers.

## 📚 Language Bindings

| Language | Directory | Implementation |
| :--- | :--- | :--- |
| **Go** | `src/` | Native Core |
| **Python** | `distconf/python/` | `ctypes` based |
| **Rust** | `distconf/rust/` | `libloading` with Safe Wrappers |
| **C++** | `distconf/cpp/` | Header-only OOP Wrapper |
| **VBA** | `distconf/vba/` | Win32 API declarations for Excel/Access |

## 🛠️ Core API Methods

Every SDK provides the following standardized interface:

- `New(profile)`: Initialize a new session.
- `Get(section, key)`: Retrieve a value.
- `Set(section, key, val)`: Update a value locally and sync.
- `OnLiveConfUpdate(callback)`: Listen for real-time remote updates.
- `OnRegistryUpdate(callback)`: Listen for real-time service registry updates.
- `ShareConfig(json_payload)`: Broadcast local state to the ecosystem.
- `ApplyFileOverride(filename)`: Load a local YAML override and return the 'local' section as JSON.
- `GetAddress(capability)`: Resolve primary `host:port` for a capability.
- `GetGRPCAddress(capability)`: Resolve gRPC `host:port` for a capability.
- `GetRESTAddress(capability)`: Resolve REST `host:port` for a capability.
- `GetCapability(capability)`: Retrieve merged capability metadata as JSON/dict.
- `GetFullConfig()`: Retrieve full `common`, `capabilities`, and `live` configuration state.
- `Decrypt(ciphertext)`: Decrypt RSA secrets.

## 🚀 Building the Bridge

Before using non-Go SDKs, you must compile the shared library:

```bash
make build-lib
```

This generates `libdistconf.so` (Linux), `libdistconf.dylib` (macOS), or `libdistconf.dll` (Windows).

## 🧪 Validation

Run the FFI validation suite to ensure your environment is correctly configured:
```bash
python3 distconf/python/ffi_validation.py
cargo run --example ffi_validation
```
