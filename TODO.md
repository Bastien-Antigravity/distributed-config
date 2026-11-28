---
microservice: distributed-config
type: note
status: active
tags:
- '#service/distributed-config'
- '#type/note'
- '#state/active'
- '#zone/3-fleet'
---

# TODO: distributed-config

## 📌 Reminders: Pending Ecosystem Verification & Design Review
- [ ] **Retain `CommonFilePath` temporarily in `CommonConfig` (`src/core/config.go`)**:
  - *Context*: `config-server` currently references `s.AppConfig.Common.CommonFilePath` in `src/server/controller.go:135`.
  - *Action*: Keep field in `CommonConfig` for ecosystem compatibility. Schedule deprecation and removal once `config-server` has been migrated to modern configuration resolution.
- [ ] **Re-introduce Empty Skeleton Generation in a Clean Manner**:
  - *Context*: Auto-creating skeleton YAML files during runtime configuration loading was removed to enforce fail-fast discovery. However, skeleton templates are useful for developer onboarding and project initialization.
  - *Action*: Re-introduce empty skeleton generation via an explicit CLI command (e.g., `config-tool init` or `config-cli init --profile <name>`) or an opt-in bootstrap utility rather than implicit runtime file creation on missing configuration.

## 🏗️ Architecture & Reliability (Audit Action Items)
- [ ] **Upgrade `ConfigMsg.Version` from `int32` to `uint64` across `distributed-config` & `config-server`**: Currently guarded with `math.MaxInt32` overflow check in `HandleOutgoingWithVersion` until synchronized across protobuf schemas.

## ✅ Completed
- [x] **Native `core.Config.Local` Decoding & Single-Pass `ApplyFileOverride` (`src/core/config.go`, `src/loader/loader.go`, `src/facade/config_facade.go`)**: Added `Local` field and deep-merge support to `core.Config`/`LoadYAML`, eliminating duplicate YAML parsing in `ApplyFileOverride`.
- [x] **Quoted Boolean String Preservation (`src/loader/loader.go`)**: Preserved explicitly quoted `"true"`/`'false'` scalars as `!!str` in `ProcessNode()` while keeping unquoted booleans typed as `!!bool`.
- [x] **Environment-First `LS_IP`/`LS_PORT` & Multi-Endpoint IP Validation (`src/loader/env_loader.go`, `src/loader/validator.go`)**: Added `LS_IP`/`LS_PORT` bootstrap support and extended `CheckTestIPs`/`CheckProductionIPs` to validate `ip`, `grpc_ip`, and `rest_ip`.
- [x] **IPv6 Endpoint Formatting & Private Key `0600` Permissions (`src/core/config.go`, `cmd/config-tool/main.go`)**: Switched `getAddr()` to `net.JoinHostPort` (`[::]:port`) and hardened `config-tool keygen` to write `private.pem` with `0600` permissions.
- [x] **Strategy Consolidation & Standalone Mode Parity (`src/strategies/standalone.go`, `cloud.go`, `test.go`)**: Synced scalar capabilities to `LiveConfig` on `StandaloneStrategy.Load`, added OCC version checks/increments to `StandaloneStrategy`, eliminated double YAML disk reads in `CloudStrategy.Load`, and delegated `TestStrategy` to `CloudStrategy`.
- [x] **Self-Broadcast `ACK` Version Fix & Explicit Disconnect Errors (`src/network/proto_handler.go`, `client.go`)**: Tracked `lastAckedVersion` so `BROADCAST_SYNC` following `ACK` still fires `onLiveConfUpdate` once, and replaced silent mock-success branches in `network.Client` with explicit disconnected errors.
- [x] **CGO Profile Aliases, Rust Library Memoization & Zone 1 Isolation (`src/cgo_bridge/`, `cmd/libdistconf/main.go`, `distconf/rust/src/lib.rs`)**: Accepted all `factory.NewStrategy` aliases in `cgo_bridge.New`, validated handles in `DistConf_Decrypt`, added `LIB_CACHE` and parity methods to Rust `DistConfig`, and isolated Go unit tests from repo symlinks.
- [x] **`CheckTestIPs()` Conditional Re-Integration (`src/loader/validator.go` / `src/strategies/test.go`)**: Enforced behind `CHECK_TEST_IPS=true` so Docker test pipelines remain unaffected while local test IP checks can be opted into.
- [x] **Thread-Local & Fallback Fixed-Buffer Error Storage in CGO Bridge (`src/cgo_bridge/errors.c`, `helpers.h`, `cmd/libdistconf/main.go`)**: Replaced `strdup`/`free` heap allocations with fixed `512`-byte thread-local + mutex-guarded global buffers (`get_last_error()`, `get_last_error_code()`), eliminating leaks and Go thread-migration `NULL` errors.
- [x] **Centralized `Config.ResolveCapability()` & `Config.GetVersion()` (`src/core/config.go`, `src/facade/config_facade.go`)**: Consolidated capability merge logic in `ResolveCapability()` and synchronized `atomic.Uint64` version state across `ConfigProtoHandler` and `core.Config`.
- [x] **Explicit `CONFIG_PATH` Fail-Fast (`src/loader/path_resolver.go`, `src/loader/loader.go`)**: Removed silent fallback when explicit `CONFIG_PATH` is missing so invalid explicit paths fail fast.
- [x] **Private Key Precedence & Production Hardening (`src/secret/crypto.go`)**: Prioritized `/etc/bastien/private.pem` over `./private.pem` and disabled `./private.pem` fallback when `APP_ENV=production`.
- [x] **Dead Code Removal & Strategy Consolidation**: Removed unused `src/core/skeleton.go` (`NewSkeletonConfig`), `Config.Apply()`, and `loader.ValidateCommonConfig()`; consolidated `ConfigStrategy` interface methods, fixed double-delete in `facade.Config.DeleteKeys`/`DeleteKeysWithOCC`, deduplicated `hydrateCapabilitiesFromLiveConfig()`, and added `GetRESTAddress` / Bastien Python import parity across polyglot wrappers.
- [x] **RCU Writer Lock & Atomic Swap Sequence (`src/core/config.go`, `src/network/proto_handler.go`)**: Added `writeMu` to `core.Config` to serialize `Set()`, `Delete()`, `Apply()`, and `ReplaceLive()` mutations while keeping `Get()`, `GetCapability()`, and `GetAddress()` 100% lock-free.
- [x] **Dynamic Capability Resolution (`src/core/config.go:getAddr()`)**: Enabled `GetAddress()`, `GetGRPCAddress()`, and `GetRESTAddress()` to resolve capabilities added dynamically via `LiveConfig` even when absent from static `c.Capabilities`.
- [x] **Revision Monotonicity & Stale Snapshot Protection (`src/network/proto_handler.go`)**: Rejected stale/duplicate `BROADCAST_SYNC` (`incoming <= lastSeenVersion`) and delayed `GET_SYNC`/`FULL_REFRESH` snapshots (`incoming < lastSeenVersion`).
- [x] **Acknowledged Delta Commits & Reconnect Race Fix (`src/strategies/cloud.go`, `src/network/client.go`)**: Changed `CloudStrategy.Set()` and `DeleteKeys()` to push delta-only updates via `UpdateConfigMapWithOCC(updates, 0)` / `DeleteKeysWithOCC(deletions, 0)` and apply locally only after server `ACK`; guarded `Client.connect()` against concurrent `Close()`.
- [x] **Fail-Fast Secret Decryption & Key Memoization (`src/secret/crypto.go`)**: Added thread-safe RSA private key caching (keyed by `BASTIEN_PRIVATE_KEY` SHA-256 hash or file path + mtime + size with `ResetKeyCache()`) and enforced fail-fast error returns in `ProcessConfigSecrets()` when `ENC(...)` tokens are present.
- [x] **Directory-Aware `SHARED_CONFIG_PATH` (`src/loader/path_resolver.go`)**: Resolved `<SHARED_CONFIG_PATH>/<targetName>.yaml` (and `standalone.yaml` fallback) when `SHARED_CONFIG_PATH` points to a directory, and cleaned up ineffectual assignment in `path_resolver_test.go`.
- [x] **Remove Dead Code `GenerateRSAKeypair()` (`src/secret/crypto.go`)**: Removed unused helper; key generation is handled directly in `cmd/config-tool/main.go`.
- [x] **Remove Dead Code `EnsureFileExists()` (`src/loader/loader.go`)**: Removed obsolete skeleton file creation from runtime loader and cleaned up unused `path/filepath` import.
- [x] **Fix Socket Resource Leak on Connection Failure**: Added `client.Close()` on `client.Open()` failure in `src/network/client.go:connect()`.
- [x] **Profile Parity (Server Capability Hydration)**: Aligned `TestStrategy.Load()` with `CloudStrategy.Load()` to copy server-provided capability sections into the `cfg.Capabilities` baseline.
- [x] **Deduplicate Schema Types**: Removed duplicate local `type ConfigServerCap` declarations in `src/strategies/cloud.go` and `src/strategies/test.go`, standardizing on `core.ConfigServerCap`.
- [x] **Resolve Static Analysis Warning in CGO Bridge**: Fixed unchecked error on `Set()` in `src/cgo_bridge/bridge_test.go:137`.
- [x] **Optimize CGO Stress Test**: Isolated handle lifecycle in `src/cgo_bridge/stress_test.go` using `standalone` profile, reducing test runtime from ~158s to ~13s while preserving race detection coverage.
- [x] **Align Strategy Docs and README**: Updated `standalone.go`, `test.go`, and `README.md` to reflect fail-fast configuration discovery (no runtime skeleton generation).
- [x] **FFI Safety Audit**: Verified all language bridge implementations follow **[[ADR-002-Static-Loading-Law|ADR-002: The Static Loading Law]]** (strictly no `dlclose`).
- [x] **Standardize error codes**: Implemented uniform error reporting across Go, CGO, Python, and Rust bridges (FEAT-007).
- [x] **Sync Bug**: Fix `GetAddress` logic in `capabilities.go`. It currently caches the initial YAML state and ignores `LiveConfig` updates (FEAT-004). (Validated)
- [x] **Handle Race Condition**: Fix mutex unlock timing in `cgo_bridge/config.go` to prevent potential use-after-free during concurrent Close/Get (FEAT-006). (Validated with stress test)
- [x] **Align CGO dynamic capability reading**: Verified `cgo_bridge.GetCapability` delegates to `session.Config.GetCapability` with LiveConfig merge and test coverage.
- [x] **Guard socket receive loop**: Implemented `syncMu` and single-reader pattern for `requestSync` in `network/client.go`.
- [x] **Initial BDD Spec migration to Obsidian Brain**.
- [x] **Implement Jittered Exponential Backoff for Network Client**.
- [x] **Refactor CGO bridge with RWMutex for thread safety**.
- [x] **Purge tracked 10.5MB `.dylib` binaries and configure `.gitignore`**.
- [x] **Realign `AGENTS.md` and `README.md` documentation and BDD spec links**.
