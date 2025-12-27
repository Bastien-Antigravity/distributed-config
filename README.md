# Distributed Config

A robust, strategy-based configuration management library for distributed systems in Go.

## Overview

`distributed-config` provides a unified interface for loading, validating, and synchronizing configuration across different environments. It supports local files, environment variables, and remote configuration servers.

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
