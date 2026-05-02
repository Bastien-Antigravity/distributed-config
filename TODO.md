# TODO: distributed-config

## 🚨 High Priority (Governance Gaps)
- [ ] **Strategy Merger (Purger Rule)**: Merge `Production` and `Staging` strategies into a single `CloudStrategy` to reduce code duplication (FEAT-004). (Approval Required)
- [ ] **Sync Bug**: Fix `GetAddress` logic in `capabilities.go`. It currently caches the initial YAML state and ignores `LiveConfig` updates (FEAT-004). (Approval Required)
- [ ] **Handle Race Condition**: Fix mutex unlock timing in `cgo_bridge/config.go` to prevent potential use-after-free during concurrent Close/Get (FEAT-006). (Approval Required)

## 🏗️ Architecture & Refactoring
- [ ] Decouple private config management to microservice-toolbox.
- [ ] Standardize error codes across all language bridges.

## 🧪 Testing & CI/CD
- [ ] Add integration tests for jittered exponential backoff.

## ✅ Completed
- [x] Initial BDD Spec migration to Obsidian Brain.
