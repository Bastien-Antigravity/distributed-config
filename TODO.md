# TODO: distributed-config

## 🏗️ Architecture & Refactoring

## 🧪 Testing & CI/CD
- [x] Add integration tests for jittered exponential backoff.
- [x] Implement CGO bridge stress tests.
- [x] Implement network resilience and reconnection tests.
- [x] Add FFI validation for Python and Rust.

## ✅ Completed
- [x] **FFI Safety Audit**: Verified all language bridge implementations follow **[[ADR-002-Static-Loading-Law|ADR-002: The Static Loading Law]]** (strictly no `dlclose`).
- [x] **Standardize error codes**: Implemented uniform error reporting across Go, CGO, Python, and Rust bridges (FEAT-007).
- [x] **Sync Bug**: Fix `GetAddress` logic in `capabilities.go`. It currently caches the initial YAML state and ignores `LiveConfig` updates (FEAT-004). (Validated)
- [x] **Handle Race Condition**: Fix mutex unlock timing in `cgo_bridge/config.go` to prevent potential use-after-free during concurrent Close/Get (FEAT-006). (Validated with stress test)
- [x] Initial BDD Spec migration to Obsidian Brain.
- [x] Implement Jittered Exponential Backoff for Network Client.
- [x] Refactor CGO bridge with RWMutex for thread safety.
