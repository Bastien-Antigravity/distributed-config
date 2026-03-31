# Distributed Config

A robust, strategy-based configuration management library for distributed systems in Go.

## Overview

`distributed-config` provides a unified interface for loading, validating, and synchronizing configuration across different environments. It supports local files, environment variables, and remote configuration servers.

## Configuration Loading Process

The library uses a layered approach to build the final configuration:

1.  **Code Defaults**: The application initializes with a hardcoded set of "safe" defaults (defined in `src/core/defaults.go`). This ensures the application can always start, even without a config file.
2.  **File Overrides**: It looks for a YAML file named after the executable (e.g., `config-cli.yaml` or `default.yaml` in specific modes). Values found in this file **overwrite** the defaults. This allows you to specify *only* the changes you need (e.g., just the DB password) rather than a full config file.
3.  **Environment Variables**: Values in the YAML file can use `${VAR_NAME}` syntax. These are expanded using the system's environment variables at runtime.
4.  **Remote Sync** (Profile Dependent): Finally, if the selected profile supports it (like `production`), the library connects to the Config Server to fetch the latest "MemConfig" updates.

## Features

- **Multi-Profile Strategy**: Built-in support for `production`, `preprod`, `test`, and `standalone` environments.
- **Remote Synchronization**: Automatically fetches and updates configuration from a central server using [safe-socket](https://github.com/Bastien-Antigravity/safe-socket).
- **Secrets Management**: Native support for environment variable expansion (e.g., `${MY_SECRET}`).
- **Fail-Safe Defaults**: Robust fallback mechanisms and skeleton generation for missing configurations.

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
	"log"

	"github.com/Bastien-Antigravity/distributed-config"
)

func main() {
	// Initialize configuration for "production" environment
	// Options: "production", "preprod", "test", "standalone"
	cfg := distributed_config.New("production")

	// Access configuration values
	fmt.Printf("Service Name: %s\n", cfg.Common.Name)
	
	// Access capability specific config
	if cfg.Capabilities.Database != nil {
		fmt.Printf("DB Host: %s\n", cfg.Capabilities.Database.IP)
	}
}
```

## Python Library

A Python wrapper is available in the `python/` directory. It uses a C-shared Go library to provide the same robust configuration logic to Python applications.

### Prerequisites

- **Go 1.25+**: Required to compile the underlying shared library.
- **Python 3.12+**: Tested and supported version.

### Building & Packaging

You can easily build the Go shared library and generate Python wheels using the provided `Makefile`:

```bash
# Compiles the Go shared library and builds the distribution wheels
make python-build
```

The compiled library will be placed in `python/distributed_config/` and the wheels will be available in `python/dist/`.

### Python Usage

```python
from distributed_config import DistributedConfig

# Initialize with a profile: "production", "preprod", "test", or "standalone"
with DistributedConfig("test") as dc:
    config = dc.get_config()
    
    # Access configuration values as a standard Python dictionary
    print(f"Service Name: {config['common']['name']}")
    
    if 'database' in config['capabilities']:
        print(f"DB Host: {config['capabilities']['database']['ip']}")
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
  telebot:
    token: "${TELEBOT_TOKEN}"
    chat_id: "${TELEBOT_CHAT_ID}"
```
