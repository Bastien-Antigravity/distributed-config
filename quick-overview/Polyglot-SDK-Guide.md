---
tags:
- '#ai/ignore'
- '#zone/3-fleet'
---
# Polyglot SDK Guide

\`distributed-config\` provides a centralized Go core that is exposed to other languages through a C-compatible ABI (CGO Bridge). This ensures 100% architectural parity across the entire microservice fleet.

## 🏗️ The CGO Bridge

The core logic resides in \`src/cgo_bridge/\`, which manages:
- **Session Handles**: Thread-safe mapping of integer handles to Go \`Config\` instances.
- **Memory Safety**: Careful management of string pointers across the FFI boundary.
- **Async Callbacks**: Bridging Go channels/callbacks to native language function pointers.

## 📚 Language Bindings

| Language | Directory | Implementation |
| :--- | :--- | :--- |
| **Go** | \`src/\` | Native Core |
| **Python** | \`distconf/python/\` | \`ctypes\` based |
| **Rust** | \`distconf/rust/\` | \`libloading\` with Safe Wrappers |
| **C++** | \`distconf/cpp/\` | Header-only OOP Wrapper |
| **VBA** | \`distconf/vba/\` | Win32 API declarations for Excel/Access |

## 🛠️ Core API Methods

Every SDK provides the following standardized interface:

- \`New(profile)\`: Initialize a new session.
- \`Get(section, key)\`: Retrieve a value.
- \`Set(section, key, val)\`: Update a value locally and sync.
- \`OnLiveConfUpdate(callback)\`: Listen for real-time remote updates.
- \`ShareConfig(json_payload)\`: Broadcast local state to the ecosystem.
- \`Decrypt(ciphertext)\`: Decrypt RSA secrets.

## 🚀 Building the Bridge

Before using non-Go SDKs, you must compile the shared library:

\`\`\`bash
make build-lib
\`\`\`

This generates \`libdistconf.so\` (Linux), \`libdistconf.dylib\` (macOS), or \`libdistconf.dll\` (Windows).

## 🧪 Validation

Run the FFI validation suite to ensure your environment is correctly configured:
\`\`\`bash
python3 distconf/python/ffi_validation.py
cargo run --example ffi_validation
\`\`\`
