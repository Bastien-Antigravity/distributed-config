---
microservice: distributed-config
type: repository
status: active
language: go
tags:
  - domain/configuration
  - domain/networking
---

# Distributed Config

A robust, strategy-based configuration management library for distributed systems in Go.

## Overview

`distributed-config` provides a unified interface for loading, validating, and synchronizing configuration across different environments. It supports local files, environment variables, and remote configuration servers.

## Configuration Loading Process

The library uses a layered approach to build the final configuration:

1.  **Code Defaults**: The application initializes with a hardcoded set of "safe" defaults (defined in `src/core/defaults.go`). This ensures the application can always start, even without a config file.
2.  **Configuration Discovery**: The library automatically searches for a YAML file in multiple locations with strict priority.
    *   **Profile-Based Search** (Target provided):
        1. `config/[profile].yaml` (Current Working Directory)
        2. `config/[profile].yaml` (Executable Directory)
    *   **Default Fallbacks** (In order):
        3. `config/[executable_name].yaml` (Current Working Directory)
        4. `config/[executable_name].yaml` (Executable Directory)
        5. `[executable_name].yaml` (Current Working Directory)
        6. `[executable_name].yaml` (Executable Directory)
    
    If a profile is specified, it **must** reside in a `config/` subdirectory. The binary name acts as the final global fallback.
3.  **Environment Variables**: Values in the YAML file can use `${VAR_NAME}` syntax. These are expanded using the system's environment variables at runtime.
4.  **Remote Sync**: Finally, if the selected profile supports it (like `production`), the library connects to the Config Server to fetch the latest "LiveConfig" updates.

## Features

- **Multi-Profile Strategy**: Built-in support for `production`, `staging`, `test`, and `standalone` environments.
- **Remote Synchronization**: Automatically fetches and updates configuration from a central server using [safe-socket](https://github.com/Bastien-Antigravity/safe-socket).
- **Secrets Management**: Native support for environment variable expansion (e.g., `${TS_PASSWORD}`).
- **Environment-First Flexibility**: Supports "Pure-Environment" deployments where a local config file is optional. If missing, the system uses `CF_IP`/`CF_PORT` to connect to the central server and hydrate required capabilities.
- **Fail-Safe & Strict**: Enforces "Mandatory Service Validation" (Fail-Fast logic) to ensure critical infrastructure like `log_server` is correctly configured (via any source) before boot.
- **Live Updates**: Support for dynamic configuration updates via callbacks. Manually calling `Set()` on the facade now correctly triggers local observers, ensuring system-wide synchronization even for local state changes.

## Installation

```bash
go get github.com/Bastien-Antigravity/distributed-config
```

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
    // Updates are automatically synchronized if the profile supports it.
    for key, value := range cfg.LiveConfig {
        fmt.Printf("LiveConfig section %s exists\n", key)
    }

    // Register a callback for when remote configuration is updated
    cfg.OnLiveConfUpdate(func(updates map[string]map[string]string) {
        fmt.Println("Live Configuration updated remotely!")
    })

    // Register a callback for when the Service Registry shifts (nodes join/leave)
    cfg.OnRegistryUpdate(func(registry map[string][]string) {
        fmt.Printf("Active nodes tracking %d services\n", len(registry["active_services"]))
    })
}
```

## Configuration Profiles

| Profile      | Description |
|---|---|
| **standalone** | Loads from local YAML file only. No network connection. |
| **test**       | Uses hardcoded "safe" defaults (127.0.0.2). Connects to server to mimic production. |
| **staging**    | Connects to Config Server (Read-only). Mandatory local configuration file. |
| **production** | Full synchronization with Config Server (GET & PUT). Enforces strict safety checks. |

## Secrets

You can use environment variables in your YAML configuration files. They will be expanded at runtime:

```yaml
capabilities:
  tele_remote:
    token: "${TR_TOKEN}"
    chat_id: "${TR_CHATID}"
```
