# AGENTS.md: distributed-config

## Service Mission & Architecture Role
`distributed-config` is the foundational configuration engine of the Bastien-Antigravity ecosystem. It implements layered configuration merging (CLI flags > local YAML > config server > environment variables), secret decryption (`ENC(...)`), real-time bidirectional synchronization with `config-server`, and provides a unified CGO bridge (`libdistconf`) for polyglot consumers (Python, Rust, C++).

- **Ecosystem Role**: Core engine wrapped by `microservice-toolbox`.
- **CGO Core**: `src/cgo_bridge` compiles into `libdistconf` for non-Go consumers.
- **Remote Sync Protocol**: `safe-socket` client connecting to `config-server` (port `3306`) with monotonic sequence gap self-healing.
- **Configuration Link**: `standalone.yaml -> ../docker-deployment/modes/local/config/native.yaml`

## Key Build & Test Commands
```bash
# Run unit tests
go test -v ./...

# Build CLI tools
go build -o bin/config-tool ./cmd/config-tool
go build -o bin/config-cli ./cmd/config-cli

# Build shared CGO engine
make build-lib
```

## Internal Architecture & Subsystems
1. **Core State & RCU (`src/core`)**:
   - `atomic.Pointer[map[string]map[string]string]` (`LiveConfig`) provides lock-free Copy-On-Write reads (`Get()`).
   - `PreviewSet` handles delta merging and `"__DELETED__"` sentinel handling to prune deleted keys and empty sections.
   - `Delete(deletions)` removes specified keys or whole sections atomically.
2. **Network Client & Protocol Handler (`src/network`)**:
   - `Client`: Manages persistent `safe-socket` TCP connection, exponential backoff reconnection, and synchronized requests.
   - `ConfigProtoHandler`: Deserializes `ConfigMsg` envelopes. Tracks `lastSeenVersion uint64`. If a sequence gap is detected (`received > lastSeen + 1`), it automatically triggers an out-of-band `FullRefresh()` snapshot recovery.
3. **Consistency & OCC Subsystem**:
   - `GetVersion()`: Returns current known monotonic revision from `config-server`.
   - `SetWithOCC()` / `UpdateConfigMapWithOCC()`: Attaches `expectedVersion` to `PUT_SYNC`. If the server rejects due to concurrent edit, returns `ErrVersionConflict` (`errors.Is(err, ErrVersionConflict)`).
4. **Deletion Subsystem**:
   - `DeleteKey(sec, key)`, `DeleteKeys(map)`, `DeleteKeysWithOCC(map, ver)` transmit `DELETE_SYNC` to `config-server`.
   - Server elevates broadcasts to `FULL_REFRESH`, replacing `LiveConfig` entirely via atomic pointer swap (`LiveConfig.Store(&sections)`), guaranteeing zero residual tombstones.

## AI Development & Integration Guidelines
1. **Never Break CGO ABI**: When modifying `src/cgo_bridge/bridge.go`, ensure C export signatures remain backwards compatible with Python (`ctypes`) and Rust (`extern "C"`).
2. **Capability Access**: `GetCapability(name, &struct)` is the canonical way to fetch capabilities. `GetLocal()` is only for non-synced local block keys.
3. **Secret Encryption Standard**: `ENC(...)` tokens are decrypted using the ecosystem master key. Plaintext passwords must never be stored.
4. **Header Ritual**: All Go files MUST adhere to the Triple-Block header standard (`ESSENTIAL PROCESS`, `DATA FLOW`, `KEY PARAMETERS`).
5. **Section Dividers**: Use `// -----------------------------------------------------------------------------` between exported functions.
6. **OCC Usage Pattern**: When updating sensitive shared state, use `cfg.SetWithOCC(updates, expectedVer)`. In case of `ErrVersionConflict`, call `cfg.Sync()`, re-read the latest config and version via `cfg.GetVersion()`, and retry the mutation.
7. **Deletion Contract**: Do not invent custom tombstones. Use `cfg.DeleteKey` / `cfg.DeleteKeys` (or sentinel `"__DELETED__"` in maps) so deletions propagate via `DELETE_SYNC` and resolve with `FULL_REFRESH`.
