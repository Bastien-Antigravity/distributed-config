---
microservice: distributed-config
type: architecture
status: active
tags:
  - domain/configuration
  - domain/networking
---

# Architecture

This document describes the internal design of the Distributed Config library.

## Data Flow

The following diagram illustrates how the library discovers, loads, and synchronizes configuration data across different environments.

```mermaid
flowchart TD
    %% Styles
    classDef core fill:#e3f2fd,stroke:#1565c0,stroke-width:2px,color:#0d47a1;
    classDef strategy fill:#e8f5e9,stroke:#2e7d32,stroke-width:2px,color:#1b5e20;
    classDef net fill:#fff8e1,stroke:#fbc02d,stroke-width:2px,color:#f57f17;
    classDef file fill:#f3e5f5,stroke:#8e24aa,stroke-width:2px,color:#4a148c;
    classDef loader fill:#fce4ec,stroke:#d81b60,stroke-width:2px,color:#880e4f;

    %% Application Node Styling
    style App fill:#37474f,stroke:#263238,stroke-width:3px,color:#ffffff

    %% Nodes
    App[Application] -->|"Init(profile)"| Facade(Facade):::core
    Facade -->|Request| Factory(Factory):::core
    Factory -->|Create| StratInterface[Strategy Interface]:::core

    subgraph Strategies [Configuration Strategies]
        direction TB
        StratInterface --> Prod[Production]:::strategy
        StratInterface --> Pre[Preprod]:::strategy
        StratInterface --> Test[Test]:::strategy
        StratInterface --> Stand[Standalone]:::strategy
    end
    style Strategies fill:#f1f8e9,stroke:#aed581,stroke-width:2px,color:#33691e

    subgraph Discovery [Loader & Discovery]
        direction TB
        Prod & Pre & Test & Stand --> Loader[Path Resolver]:::loader
        Loader -->|Checks| CWD[Current Working Dir]:::file
        Loader -->|Checks| BinDir[Binary Dir]:::file
        CWD & BinDir -->|Priority| ConfigFolder[config/ folder]:::file
        ConfigFolder --> YAML[YAML File]:::file
    end
    style Discovery fill:#fce4ec,stroke:#f06292,stroke-width:2px,color:#880e4f

    subgraph Sync [Dynamic Sync]
        direction TB
        Prod & Pre & Test -->|Propagate| Net[Network Config Server]:::net
        Net -.->|BROADCAST_SYNC| MemConfig[MemConfig Map]:::core
        Net -.->|BROADCAST_REGISTRY| Registry[Service Registry JSON]:::core
        MemConfig -.->|Trigger| Callback[User Callback]:::core
    end
    style Sync fill:#fffde7,stroke:#fff176,stroke-width:2px,color:#f57f17
```

## Key Components

### 1. Facade (`src/facade`)
The primary entry point (`distributed_config.New(profile)`). It acts as a wrapper around the core data, providing:
*   **Static Access**: Direct access to YAML-loaded fields (e.g., `cfg.Common.Name`).
*   **Dynamic Access**: Access to the `MemConfig` map for runtime updates.
*   **Callbacks**: Mechanism to register listeners (`OnMemConfUpdate`) for remote configuration changes.

### 2. Loader & Discovery (`src/loader`)
Handles the complex logic of finding and parsing configuration files.
*   **Path Resolver**: Automatically searches for YAML files in multiple locations (CWD, Binary Directory, and their respective `config/` subfolders).
*   **Precedence**: Explicitly prioritized to allow `config/default.yaml` or project-specific files to override generated skeletons.
*   **Env Expansion**: Processes `${VAR_NAME}` syntax during YAML parsing.

### 3. Strategies (`src/strategies`)
Implements the core logic for retrieving and synchronizing configuration based on the requested profile.
*   **Production**: Full bidirectional sync (GET/PUT) with the Config Server. authoritative local file.
*   **Preprod**: Read-only sync (GET) with the Config Server.
*   **Test**: Bootstraps with local defaults (e.g., `127.0.0.2`) then mimics Production behavior.
*   **Standalone**: Offline mode. Only uses local file discovery via the Loader.

### 4. Network & Protocol (`src/network`)
Manages communication with the remote Config Server using a slimmed-down Protobuf protocol wrapping unstructured JSON arrays/maps.
*   **Safe Socket**: High-performance TCP communication via `github.com/Bastien-Antigravity/safe-socket`.
*   **Proto Handler**: Parses generic `GET_SYNC`, `PUT_SYNC`, `BROADCAST_SYNC`, and `BROADCAST_REGISTRY` commands. Routes unstructured JSON blobs to `MemConfig` or Registry callbacks without needing rigidly coupled structs.

## Configuration Precedence

1.  **Code Defaults**: Hardcoded values in `NewDefaultConfig()`.
2.  **Remote Sync** (Dynamic): Merged into `MemConfig` at runtime.
3.  **Local YAML**: Discovered via the Loader. **Value in YAML always overrides Server/Default values.**
4.  **Environment Variables**: Overwrite corresponding YAML values via expansion.
