---
microservice: distributed-config
type: repository
status: active
language: go
tags:
- '#service/distributed-config'
- '#domain/configuration'
- '#domain/networking'
- '#zone/3-fleet'
- '#type/repository'
- '#state/active'
---

# Distributed Config

A robust, strategy-based configuration management library for distributed systems in Go.

## Overview

`distributed-config` provides a unified interface for loading, validating, and synchronizing configuration across different environments. It supports local files, environment variables, and remote configuration servers.

## Configuration Loading Process

The library uses a layered approach to build the final configuration:

1.  **Code Defaults**: The application initializes with a hardcoded set of "safe" defaults (defined in `src/core/defaults.go`). This ensures the application can always start, even without a config file.
2.  **Auto-Skeleton Generation**: If the target configuration file (e.g., `standalone.yaml`) is missing, the library **automatically recreates it** using the internal master template. This ensures a functional "Zero-Config" bootstrap out of the box.
3.  **Configuration Discovery**: The library automatically searches for a YAML file with strict priority:
    *   **Priority 0 — Explicit Environment Overrides (Highest Precedence)**:
        *   `CONFIG_PATH`: Directly specifies the absolute or relative path to the YAML configuration file (e.g., `docker-deployment/shared-config/native.yaml`).
        *   `SHARED_CONFIG_PATH`: Alternative environment override pointing to the configuration file or root shared-config directory.
    *   **Priority 1 — Profile-Based Search** (Target provided, e.g. `standalone`):
        1. `config/[profile].yaml` (Current Working Directory)
        2. `config/[profile].yaml` (Executable Directory)
        3. `[profile].yaml` (Current Working Directory)
    *   **Priority 2 — Default Fallbacks** (In order):
        4. `config/[executable_name].yaml` (Current Working Directory)
        5. `config/[executable_name].yaml` (Executable Directory)
        6. `[executable_name].yaml` (Current Working Directory)
        7. `[executable_name].yaml` (Executable Directory)
    
    The binary name acts as the final global fallback.
4.  **Environment Variables**: Values in the YAML file can use `${VAR_NAME:default}` syntax. These are expanded using the system's environment variables at runtime.
5.  **Remote Sync**: Finally, if the selected profile supports it (like `production`), the library connects to the Config Server to fetch the latest "LiveConfig" updates.

## Features

- **Multi-Profile Strategy**: Built-in support for `production`, `staging`, `test`, and `standalone` environments.
- **Remote Synchronization**: Automatically fetches and updates configuration from a central server using [safe-socket](https://github.com/Bastien-Antigravity/safe-socket).
- **Secrets Management**: Native support for environment variable expansion (e.g., `${TS_PASSWORD}`).
- **Robust Expansion Engine**: Uses a dedicated Regex-based expansion engine that supports the `${VAR:default}` syntax, ensuring reliable fallbacks even when environment variables are missing.
- **Environment-First Flexibility**: Supports "Pure-Environment" deployments where a local config file is optional. If missing, the system uses `CF_IP`/`CF_PORT` to connect to the central server and hydrate required capabilities.
- **Fail-Safe & Strict**: Enforces "Mandatory Service Validation" (Fail-Fast logic) to ensure critical infrastructure like `log_server` is correctly configured (via any source) before boot.
- **Live Updates**: Support for dynamic configuration updates via callbacks. Manually calling `Set()` (bulk) or `SetSingle()` (convenience) on the facade now correctly triggers local observers and automatically synchronizes with the fleet.
- **High Performance & Lock-Free**: Uses an **Atomic Pointer Swap (RCU)** architecture. Configuration reads (`Get`) are 100% lock-free and non-blocking, ensuring zero-latency configuration access for high-frequency microservices. Even during bulk updates, readers always see a consistent snapshot.
- **Polyglot Ecosystem (v0.0.1)**: Native support for **Python, Rust, C/C++, and VBA** via a centralized CGO-based shared library (`libdistconf`). Achieve 100% architectural parity across your entire microservice fleet with raw error transparency.
 
## 🛡️ Feature Specs & Governance (BDD)
The behavior of this microservice is governed by strict specifications in the ****:
- **Core Strategy**: , , 
- **Live Sync**: , 
- **Security**: 
- **Polyglot & FFI**: , , 
- **Resilience**: , 
- **Advanced Injection**: 

## Installation

```bash
go get github.com/Bastien-Antigravity/distributed-config
```

### **Shared Library (Polyglot Support)**
For Python, Rust, or C++ integration, build the centralized shared library:
```bash
make build-lib
```
This generates `release/libdistconf.so` (or `.dylib` on macOS), which is used by the `microservice-toolbox` facades.

## Usage

Import the library and initialize it with your desired profile:

```go
package main

import (
	"fmt"

	distributed_config "github.com/Bastien-Antigravity/distributed-config"
)

func main() {
	// Initialize configuration for "production" environment
	// Options: "production", "staging", "test", "standalone"
	cfg := distributed_config.New("production")

	// Access static configuration values using typed unmarshaling helpers
	fmt.Printf("Service Name: %s\n", cfg.Common.Name)
	
	type GenericServer struct {
		IP   string `json:"ip"`
		Port string `json:"port"`
	}
	var tsDb GenericServer
	if err := cfg.GetCapability("timescaledb", &tsDb); err == nil && tsDb.IP != "" {
		fmt.Printf("DB Host: %s\n", tsDb.IP)
	}

    // Access dynamic (Live) configuration
    // Updates are automatically synchronized and use Atomic Pointer Swaps.
    // For iterating over all values, load the current snapshot:
    if snapshot := cfg.LiveConfig.Load(); snapshot != nil {
        for section, kv := range *snapshot {
            fmt.Printf("LiveConfig section %s has %d keys\n", section, len(kv))
        }
    }

    // Register a callback for when remote configuration is updated
    cfg.OnLiveConfUpdate(func(updates map[string]map[string]string) {
        fmt.Println("Live Configuration updated remotely!")
    })

    // Register a callback for when the Service Registry shifts (nodes join/leave)
    cfg.OnRegistryUpdate(func(registry map[string][]string) {
        fmt.Printf("Active nodes tracking %d services\n", len(registry["active_services"]))
    })

    // Update configuration (Bulk)
    cfg.Set(map[string]map[string]string{
        "custom_section": {"status": "active"},
    })

    // Update configuration (Single - Helper)
    cfg.SetSingle("custom_section", "mode", "optimized")
}
```

## Configuration Profiles

| Profile      | Description |
|---|---|
| **standalone** | Loads from local YAML file only. No network connection. |
| **test**       | Uses hardcoded "safe" defaults (127.0.0.2). Connects to server to mimic production. |
| **staging**    | CloudStrategy (Read-Only). Remote sync disabled. Mandatory local configuration file. |
| **production** | CloudStrategy (Full Sync). Enforces strict safety checks (e.g. No 127.0.0.2). |

## Secrets

You can use environment variables in your YAML configuration files. They will be expanded at runtime:

```yaml
capabilities:
  tele_remote:
    token: "${TR_TOKEN}"
    chat_id: "${TR_CHATID}"
```

## Security & Encryption (v0.0.1)

`distributed-config` supports native RSA encryption for sensitive fields. If a string is wrapped in `ENC(...)`, it will be automatically decrypted at runtime using a private key.

### **Key Distribution Policy**
- **Public Key (`public.pem`)**: Non-sensitive. Used by developers to encrypt secrets. Distribute via secure internal channels; **DO NOT** commit to Git.
- **Private Key (`private.pem`)**: Critical secret. **MUST NOT** be committed to Git. In production, mount it at `/etc/bastien/private.pem`.

### **Utilities**
The unified **`config-tool`** is provided in the `cmd/` directory:

1.  **Generate Keys**:
    ```bash
    go run ./cmd/config-tool keygen --dir .
    ```

2.  **Encrypt a Secret**:
    ```bash
    go run ./cmd/config-tool encrypt --key public.pem --token "your-secret-here"
    ```

## **Polyglot Usage (Python/Rust/C++/VBA)**

The `distributed-config` core is exposed via a stable C ABI. 

### **Key Bridge API:** 
- `DistConf_New(profile)`: Initialize a new session.
- `DistConf_Get(handle, section, key)`: Retrieve a value.
- `DistConf_Set(handle, section, key, val)`: Update a value locally (triggers callbacks).
- `DistConf_OnLiveConfUpdate(handle, callback)`: Register a live update listener.
- `DistConf_Sync(handle)`: Force a manual refresh from the Config Server.
- `DistConf_ShareConfig(handle, json)`: Broadcast state (flat or nested map) to the ecosystem.
- `DistConf_ApplyFileOverride(handle, filename)`: Load a local YAML override. Returns a JSON string of the 'local' section.
- `DistConf_ValidateMandatoryServices(handle)`: Ensure the environment satisfies mandatory services.
- `DistConf_Decrypt(handle, ciphertext)`: Decrypt a secret.
- `DistConf_GetLastError()`: Retrieve the last raw engine-level error message.

For high-level usage, refer to the **`distconf/`** directory or the **`microservice-toolbox`** implementations.

---

## Testing

Run the full test suite (including the new RSA round-trip tests):
```bash
go test -v ./...
```
