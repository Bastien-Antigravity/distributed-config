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
  - *Context*: `config-server` currently references `s.AppConfig.Common.CommonFilePath` in `src/server/controller.go:120`.
  - *Action*: Keep field in `CommonConfig` and placeholder in `NewSkeletonConfig()` for ecosystem compatibility. Schedule deprecation and removal once `config-server` has been migrated to modern configuration resolution.
- [ ] **Re-introduce Empty Skeleton Generation in a Clean Manner**:
  - *Context*: Auto-creating skeleton YAML files during runtime configuration loading was removed to enforce fail-fast discovery. However, skeleton templates are useful for developer onboarding and project initialization.
  - *Action*: Re-introduce empty skeleton generation via an explicit CLI command (e.g., `config-tool init` or `config-cli init --profile <name>`) or an opt-in bootstrap utility rather than implicit runtime file creation on missing configuration.
- [ ] **Dynamic Capability Resolution (`src/core/config.go:getAddr()`)**:
  - *Context*: When `Capabilities` map is non-nil but does not contain a requested service, `getAddr()` currently returns an error before checking dynamic overrides in `LiveConfig`.
  - *Action*: Verify ecosystem-wide (across all microservices and capability conventions) whether dynamic overrides (`capabilities.<svc>.*` or `<svc>.*`) should serve as fallback before rejecting unknown services, or if capabilities must always be pre-registered in the static baseline.
- [ ] **RSA Private Key Memoization (`src/secret/crypto.go`)**:
  - *Context*: Currently, decrypting `ENC(...)` tokens parses and reads the RSA private key from disk on every invocation.
  - *Action*: Evaluate architecture/security trade-offs of in-memory key memoization (`sync.RWMutex` + cache invalidation via mtime or explicit reload) vs ephemeral disk reads across long-running daemons and short-lived CLI invocations.
- [ ] **Ineffectual Assignment in Test (`src/loader/path_resolver_test.go:41`)**:
  - *Context*: `golangci-lint` flags `ineffectual assignment to exeName`.
  - *Action*: Confirm test expectations at ecosystem level and clean up the unused reassignment.

## 🏗️ Architecture & Reliability (Audit Action Items)
- [ ] **RCU Writer Lock & Atomic Swap Sequence (Deferred)**: Protect `Set()`, `Apply()`, and the `PreviewSet -> UpdateConfigMap -> Apply` sequence with a writer mutex or CAS retry loop to prevent concurrent writers (e.g., caller `Set()` vs background `Watch()` incoming frames) from overwriting each other's live configuration.
- [ ] **`CheckTestIPs()` Conditional Re-Integration (`src/loader/validator.go` / `src/strategies/test.go`)**:
  - *Context*: `CheckTestIPs()` strictly enforces that all capability IPs equal `127.0.0.2`. It was commented out in `test.go` because Docker test pipelines use container IPs (`172.x.x.x`) or hostnames, and falling back to `standalone.yaml` uses `127.0.0.1`.
  - *Action*: Re-enable conditionally behind an environment flag (e.g., `CHECK_TEST_IPS=true`) or when an explicit `test.yaml` is loaded, avoiding false test failures in Docker environments while preserving local test integrity checks.
- [ ] **Thread-Local Storage / Error Reporting Sanitization in CGO Bridge**: Address OS thread affinity in `errors.c` when accessed via Go goroutines, and guard against `strdup` memory leaks across foreign language calls.

## ✅ Completed
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
