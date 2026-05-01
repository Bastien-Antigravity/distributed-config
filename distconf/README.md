# DistConf Polyglot SDK

This directory contains the official language wrappers for the `distributed-config` CGO bridge. These wrappers provide a native, object-oriented interface for non-Go languages while maintaining 100% architectural parity with the Go core.

## Directory Structure

- **[cpp/](./cpp)**: Header-only C++11+ wrapper.
- **[python/](./python)**: Pure Python ctypes-based wrapper.
- **[rust/](./rust)**: Safe Rust wrapper using `libloading`.
- **[vba/](./vba)**: VBA module for Excel/Access integration.

## Core API Design

All wrappers follow the same naming convention and provide the following core methods:

| Method | Description |
| :--- | :--- |
| `New(profile)` | Initializes a new configuration session. |
| `Get(section, key)` | Retrieves a configuration value. |
| `Set(section, key, val)` | Updates a value locally (triggers callbacks). |
| `Sync()` | Forces a refresh from the Config Server. |
| `ShareConfig(payload)` | Broadcasts state to the entire ecosystem. |
| `ValidateMandatoryServices()` | Ensures connectivity environment is valid. |
| `Decrypt(ciphertext)` | Decrypts RSA-protected secrets. |
| `OnLiveConfUpdate(cb)` | Registers a listener for real-time config updates. |
| `OnRegistryUpdate(cb)` | Registers a listener for service registry updates. |

## Build Instructions

Before using any of these SDKs, you must build the shared library:

```bash
cd ..
make build-lib
```

This generates the necessary binaries in `distconf/libdistconf/libdistconf.so` (or `.dylib`).

## Testing

Each language has its own test suite located in its respective directory.
Refer to the `TESTING.md` in each subfolder for detailed instructions.
