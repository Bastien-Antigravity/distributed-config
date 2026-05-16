---
tags:
- '#ai/ignore'
---
# General Project Information

This document provides a quick reference for the project structure, configuration files, and common developer tasks.

## Project Structure

- **`src/`**: Core Go source code.
    - `core/`: Data structures and RCU logic.
    - `facade/`: Main API layer.
    - `loader/`: YAML and Environment parsing.
    - `network/`: Resilient client-server communication.
    - `strategies/`: Environment-specific logic (Production, Staging, etc.).
    - `cgo_bridge/`: FFI layer for multi-language support.
- **`cmd/`**: Entry points for binaries and shared libraries.
    - `config-cli/`: CLI tool for managing configurations.
    - `libdistconf/`: The CGO-exported shared library (`.so`, `.dylib`, `.dll`).
- **`distconf/`**: Language-specific bindings and examples.
    - `cpp/`, `python/`, `rust/`, `vba/`.
- **`config/`**: Default configuration files used by the library.
- **`quick-overview/`**: This documentation directory.

## Key Configuration Files

- **`standalone.yaml`**: The primary local configuration file used when running in `Standalone` mode. It defines the "ground truth" for local development.
- **`go.mod` / `go.sum`**: Go module dependencies.
- **`Makefile`**: The central orchestration tool for building, testing, and cleaning the project.
- **`VERSION.txt`**: Tracks the current semantic version of the library.

## Developer Quick Start

### Building
To build all Go binaries and the shared library:
```bash
make build
```

### Dependency Management
The project uses Go modules. To update dependencies:
```bash
go mod tidy
```

### Versioning
Versions are tracked in `VERSION.txt`. When releasing, ensure this file is updated to reflect the new version, as it is often compiled into the binaries.

## Environment Variables
The library supports dynamic expansion of environment variables in YAML files. Use the syntax `${VAR_NAME}` or `${VAR_NAME:default_value}` within your configuration files.
