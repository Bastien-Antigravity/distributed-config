# TODO: distributed-config

## 🚨 High Priority (Governance Gaps) 
- [ ] **Strategy Merger (Purger Rule)**: Merge `Production` and `Staging` strategies into a single `CloudStrategy` to reduce code duplication (FEAT-004). (Approval Required)
- [ ] **Sync Bug**: Fix `GetAddress` logic in `capabilities.go`. It currently caches the initial YAML state and ignores `LiveConfig` updates (FEAT-004). (Approval Required)
- [x] **Handle Race Condition**: Fix mutex unlock timing in `cgo_bridge/config.go` to prevent potential use-after-free during concurrent Close/Get (FEAT-006). (Validated with stress test)
- [ ] **FFI Safety Audit**: Verify all language bridge implementations follow **[[ADR-002-Static-Loading-Law|ADR-002: The Static Loading Law]]** (strictly no `dlclose`).

## 🏗️ Architecture & Refactoring
- [ ] Decouple private config management to microservice-toolbox.
- [ ] Standardize error codes across all language bridges.

## 🧪 Testing & CI/CD
- [x] Add integration tests for jittered exponential backoff.
- [x] Implement CGO bridge stress tests.
- [x] Implement network resilience and reconnection tests.
- [x] Add FFI validation for Python and Rust.

## ✅ Completed
- [x] Initial BDD Spec migration to Obsidian Brain.
- [x] Implement Jittered Exponential Backoff for Network Client.
- [x] Refactor CGO bridge with RWMutex for thread safety.
