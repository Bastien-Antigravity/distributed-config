# libdistconf

This directory contains the core binary artifacts for the `distributed-config` polyglot ecosystem.

## Artifacts

- **`libdistconf.h`**: The C header file containing all function declarations.
- **`libdistconf.so`**: Shared library for Linux/Unix systems.
- **`libdistconf.dylib`**: Shared library for macOS systems.

## API Mapping & Parity Reference

The following table maps the C ABI exports to their original Go counterparts in the `distributed-config` library.

| C ABI Function | Go Equivalent | Description |
| :--- | :--- | :--- |
| `DistConf_New(profile)` | `distributed_config.New(profile)` | Initializes a new configuration session. |
| `DistConf_Close(handle)` | N/A (GC managed) | Releases resources for a specific session. |
| `DistConf_Get(handle, sec, key)` | `cfg.Get(sec, key)` | Retrieves a configuration value. |
| `DistConf_Set(handle, sec, key, val)` | `cfg.Set(sec, key, val)` | Updates a local value (triggers callbacks). |
| `DistConf_Sync(handle)` | `cfg.Sync()` | Triggers manual synchronization with server. |
| `DistConf_OnLiveConfUpdate(h, cb)` | `cfg.OnLiveConfUpdate(f)` | Registers a dynamic update listener. |
| `DistConf_ShareObject(h, sec, json)` | `cfg.ShareObject(sec, obj)` | Broadcasts state to the ecosystem. |
| `DistConf_ApplyFileOverride(h, f)` | `cfg.ApplyFileOverride(f)` | Loads YAML override, returns 'local' as JSON. |
| `DistConf_ValidateMandatoryServices(h)` | `cfg.ValidateMandatoryServices()` | Validates required infrastructure. |
| `DistConf_GetAddress(h, cap)` | `cfg.GetAddress(cap)` | Resolves `host:port` for a capability. |
| `DistConf_GetGRPCAddress(h, cap)` | `cfg.GetGRPCAddress(cap)` | Resolves gRPC `host:port`. |
| `DistConf_GetCapability(h, cap)` | `cfg.GetCapability(cap, &t)` | Returns capability as JSON string. |
| `DistConf_GetFullConfig(h)` | `cfg` (Live/Common) | Returns the entire configuration as JSON. |
| `DistConf_Decrypt(h, ciphertext)` | `cfg.Decrypt(ciphertext)` | Decrypts an RSA secret. |
| `DistConf_FreeString(ptr)` | N/A (Internal) | **MANDATORY** helper to free Go-allocated strings. |

### Note on Manual Synchronization (`Sync`)
While `New()` automatically performs an initial synchronization, the `Sync()` method is exposed to allow polyglot services (Python/Rust) to manually trigger a refresh of the `LiveConfig` from the central Config Server at runtime. This is particularly useful for long-running services that need to force-update their state without a full restart.

### Note on High Performance & Concurrency
All reads via `DistConf_Get` and `DistConf_GetFullConfig` are backed by an **Atomic Pointer Swap (RCU)** architecture in the Go core. This ensures that non-Go clients (Python, Rust, C++) enjoy lock-free, zero-latency configuration access even during heavy background synchronization events.
