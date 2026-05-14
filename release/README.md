# Distributed Config - Build Artifacts

This directory contains the compiled binaries and shared libraries for the `distributed-config` ecosystem.

## 🛠 Included Artifacts

### 1. Config Tool (`config-tool`)
A centralized CLI utility used for:
- Generating RSA key pairs for secret management.
- Encrypting sensitive configuration strings (tokens, passwords).
- Validating YAML configuration files.

**Binaries included:**
- `config-tool-darwin-amd64` (macOS Intel)
- `config-tool-darwin-arm64` (macOS Apple Silicon)
- `config-tool-linux-amd64` (Linux)
- `config-tool-windows-amd64.exe` (Windows)

### 2. Polyglot Shared Library (`libdistconf`)
*Note: This is typically generated via `make build-lib`.*

The C-ABI shared library enables architectural parity across:
- **Python** (via `ctypes`)
- **Rust** (via `bindgen`)
- **C/C++**
- **VBA**

## 🏗 Build Instructions

To regenerate these artifacts from source, use the following commands from the project root:

```bash
# Build all config-tool binaries
make build-tools

# Build the polyglot shared library
make build-lib
```

## 🛡️ Distribution Policy
These artifacts are generated for deployment and should be handled according to the environment's security policies. **Never** include private keys (`private.pem`) in this directory or any public distribution.

---
*Bastien-Antigravity Fleet Management*
