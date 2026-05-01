# Testing: C++ SDK

## Prerequisites

1. Build the Go library:
   ```bash
   make build-lib
   ```
2. A C++11 compatible compiler (g++, clang).

## Running Tests

From the `distconf/cpp` directory:

```bash
g++ tests/test_distconf.cpp -I. -L../../release -ldistconf -o test_app
LD_LIBRARY_PATH=../../release ./test_app
```

## Coverage

- **RAII Compliance**: Ensures handle is released on scope exit.
- **Data Integrity**: Verifies Get/Set consistency.
- **Safety Checks**: Validates mandatory service detection.
