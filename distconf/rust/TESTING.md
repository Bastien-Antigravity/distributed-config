# Testing: Rust SDK

## Prerequisites

1. Build the Go library:
   ```bash
   make build-lib
   ```
2. Install Rust toolchain (cargo).

## Running Tests

From the `distconf/rust` directory:

```bash
cargo test
```

## Coverage

- **Initialization**: Verifies library loading and session creation.
- **Data Access**: Tests the Get/Set bridge.
- **Drop Logic**: Ensures cleanup happens without panics.
