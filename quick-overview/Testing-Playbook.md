---
tags:
- '#ai/ignore'
---
# Testing Playbook

Reliability is a core requirement for a configuration library. This project employs a multi-layered testing strategy to ensure correctness, performance, and resilience.

## Test Categories

### Unit Tests
Unit tests are co-located with the source code in `_test.go` files. They focus on individual component logic.
- **Example**: `src/loader/loader_test.go` validates YAML parsing and environment expansion.

### Stress Tests
Stress tests focus on identifying race conditions, particularly in the CGO bridge where Go and C memory interact.
- **Location**: `src/cgo_bridge/stress_test.go`
- **Focus**: Concurrent initialization, access, and destruction of configuration handles.

### Network Resilience Tests
These tests simulate network failures, latency, and server drops to ensure the client reconnects and maintains a consistent state.
- **Location**: `src/network/network_resilience_test.go`
- **Key Scenarios**: Socket timeout, connection refused, intermittent drops.

### FFI / Polyglot Validation
Each language binding has its own test suite to ensure the C ABI is correctly mapped and handled.
- **Python**: `distconf/python/tests/test_distconf.py`
- **Rust**: `distconf/rust/src/lib.rs` (integrated tests)
- **C++**: `distconf/cpp/tests/test_distconf.cpp`

## Running Tests

### All Go Tests
The simplest way to run all Go-based tests is via the Makefile:
```bash
make test
```

### Specific Package Tests
To run tests for a specific package with more detail:
```bash
go test -v ./src/core/...
```

### Race Detection
Always run tests with the race detector enabled when modifying core logic or the bridge:
```bash
go test -race ./src/...
```

### Language Binding Tests
Each binding has its own requirements (e.g., `pytest` for Python, `cargo test` for Rust). Refer to the `README.md` within each `distconf/` subdirectory for specific instructions.

## Test Data
Test-specific configuration files are typically located in the same directory as the tests, often named `*.test.yaml`. These are used to provide deterministic inputs for validation logic.
