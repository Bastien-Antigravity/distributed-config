# Architecture

This document describes the internal design of the Distributed Config library.

## Data Flow

```mermaid
flowchart TD
    %% Styles
    classDef core fill:#e3f2fd,stroke:#1565c0,stroke-width:2px,color:#0d47a1;
    classDef strategy fill:#e8f5e9,stroke:#2e7d32,stroke-width:2px,color:#1b5e20;
    classDef net fill:#fff8e1,stroke:#fbc02d,stroke-width:2px,color:#f57f17;
    classDef file fill:#f3e5f5,stroke:#8e24aa,stroke-width:2px,color:#4a148c;

    %% Application Node Styling
    style App fill:#37474f,stroke:#263238,stroke-width:3px,color:#ffffff

    %% Nodes
    App[Application] -->|Init(profile)| Facade(Facade):::core
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

    subgraph Sources [Data Sources]
        direction TB
        Prod -->|TCP/Cap'n Proto| Net[Network Config Server]:::net
        Pre -->|TCP/Cap'n Proto| Net
        Test -->|TCP/Cap'n Proto| Net
        Stand -->|Read| Local[Local File]:::file
    end
    style Sources fill:#fffde7,stroke:#fff176,stroke-width:2px,color:#f57f17
```

## Key Components

### 1. Facade (`src/facade`)
The entry point (`distributed_config.New(profile)`). It validates the requested profile and delegates instantiation to the Factory.

### 2. Factory (`src/factory`)
Responsible for creating the appropriate strategy instance based on the profile string.

### 3. Strategies (`src/strategies`)
Implements the core logic for retrieving configuration data.
*   **Production**: Connects to the remote Config Server (GET & PUT) with full synchronization.
*   **Preprod**: Connects to the Config Server (GET only) without pushing updates.
*   **Test**: Uses local defaults (e.g., 127.0.0.2) but mimics the Production connection behavior.
*   **Standalone**: Operates offline, reading configuration from local YAML files.

### 4. Network (`src/network`)
Handles communication with the Config Server.
*   **Client**: Manages the connection lifecycle.
*   **Safe Socket**: Uses `github.com/Bastien-Antigravity/safe-socket` for reliable, length-prefixed TCP communication.
