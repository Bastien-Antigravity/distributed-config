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

## 🏗️ Architecture & Refactoring
- [ ] Align CGO bridge dynamic capability reading (`session.Config.GetCapability`) with Go core.
- [ ] Guard socket receive loop against concurrent sync calls in `network/client.go`.
- [ ] Memoize RSA private key in `src/secret/crypto.go` to eliminate disk I/O on repeated decryption.

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
- [x] Purge tracked 10.5MB `.dylib` binaries and configure `.gitignore`.
- [x] Realine `AGENTS.md` and `README.md` documentation and BDD spec links.
