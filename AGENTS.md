# AGENTS.md: distributed-config

## Service Mission & Architecture Role
`distributed-config` is the foundational configuration engine of the Bastien-Antigravity ecosystem. It implements layered configuration merging (CLI flags > local YAML > config server > environment variables), secret decryption (`ENC(...)`), and provides a unified CGO bridge (`libdistconf`) for polyglot consumers (Python, Rust, C++).

- **Ecosystem Role**: Core engine wrapped by `microservice-toolbox`.
- **CGO Core**: `src/cgo_bridge` compiles into `libdistconf` for non-Go consumers.
- **Configuration Link**: `standalone.yaml -> ../docker-deployment/modes/local/config/native.yaml`

## Key Build & Test Commands
```bash
# Run unit tests
go test -v ./...

# Build CLI tool
go build -o bin/distconf ./cmd/distconf

# Build shared CGO engine
make shared-lib
```

## AI Development & Integration Guidelines
1. **Never Break CGO ABI**: When modifying `src/cgo_bridge/bridge.go`, ensure C export signatures remain backwards compatible with Python (`ctypes`) and Rust (`extern "C"`).
2. **Capability Access**: `GetCapability(name, &struct)` is the canonical way to fetch capabilities. `GetLocal()` is only for non-synced local block keys.
3. **Secret Encryption Standard**: `ENC(...)` tokens are decrypted using the ecosystem master key. Plaintext passwords must never be stored.
4. **Header Ritual**: All Go files MUST adhere to the Triple-Block header standard (`ESSENTIAL PROCESS`, `DATA FLOW`, `KEY PARAMETERS`).
5. **Section Dividers**: Use `// -----------------------------------------------------------------------------` between exported functions.
