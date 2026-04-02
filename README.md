# Distributed Config

A robust, strategy-based configuration management library for distributed systems in Go.

## Overview

`distributed-config` provides a unified interface for loading, validating, and synchronizing configuration across different environments. It supports local files, environment variables, and remote configuration servers.

## Configuration Loading Process

The library uses a layered approach to build the final configuration:

1.  **Code Defaults**: The application initializes with a hardcoded set of "safe" defaults (defined in `src/core/defaults.go`). This ensures the application can always start, even without a config file.
2.  **Configuration Discovery**: The library automatically searches for a YAML file in multiple locations. The order of priority is:
    *   **Current Working Directory** (for development):
        1. `config/[executable_name].yaml`
        2. `config/default.yaml`
        3. `[executable_name].yaml`
    *   **Binary Directory** (for installed environments):
        1. `config/[executable_name].yaml`
        2. `config/default.yaml`
        3. `[executable_name].yaml`
    
    This flexible approach allows you to place your configuration in a standardized `config/` folder or keep it alongside the executable.
3.  **Environment Variables**: Values in the YAML file can use `${VAR_NAME}` syntax. These are expanded using the system's environment variables at runtime.
4.  **Remote Sync**: Finally, if the selected profile supports it (like `production`), the library connects to the Config Server to fetch the latest "MemConfig" updates.

## Features

- **Multi-Profile Strategy**: Built-in support for `production`, `preprod`, `test`, and `standalone` environments.
- **Remote Synchronization**: Automatically fetches and updates configuration from a central server using [safe-socket](https://github.com/Bastien-Antigravity/safe-socket).
- **Secrets Management**: Native support for environment variable expansion (e.g., `${TS_PASSWORD}`).
- **Fail-Safe Defaults**: Robust fallback mechanisms and skeleton generation for missing configurations.
- **Dynamic Updates**: Support for memory configuration updates via callbacks.

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

	"github.com/Bastien-Antigravity/distributed-config"
)

func main() {
	// Initialize configuration for "production" environment
	// Options: "production", "preprod", "test", "standalone"
	cfg := distributed_config.New("production")

	// Access static configuration values
	fmt.Printf("Service Name: %s\n", cfg.Common.Name)
	
	if cfg.Capabilities.TimescaleDb != nil {
		fmt.Printf("DB Host: %s\n", cfg.Capabilities.TimescaleDb.IP)
	}

    // Access dynamic (Memory) configuration
    // Updates are automatically synchronized if the profile supports it.
    for key, value := range cfg.MemConfig {
        fmt.Printf("MemConfig section %s exists\n", key)
    }

    // Register a callback for when remote configuration is updated
    cfg.OnMemConfUpdate(func(updates map[string]map[string]string) {
        fmt.Println("Memory Configuration updated remotely!")
    })
}
```

## Configuration Profiles

| Profile      | Description |
|---|---|
| **standalone** | Loads from local YAML file only. No network connection. |
| **test**       | Uses hardcoded "safe" defaults (127.0.0.2). Connects to server to mimic production. |
| **preprod**    | Connects to Config Server (GET only). Local file is authoritative. |
| **production** | Full synchronization with Config Server (GET & PUT). Enforces safety checks. |

## Secrets

You can use environment variables in your YAML configuration files. They will be expanded at runtime:

```yaml
capabilities:
  tele_remote:
    token: "${TR_TOKEN}"
    chat_id: "${TR_CHATID}"
```
