# Testing: Python SDK

## Prerequisites

1. Build the Go library:
   ```bash
   make build-lib
   ```
2. Install Python 3.

## Running Tests

From the `distconf/python` directory:

```bash
python3 -m unittest tests/test_distconf.py
```

## Coverage

- **Lifecycle**: Verifies initialization and handle disposal.
- **Data Access**: Tests synchronous Get/Set operations.
- **Validation**: Ensures mandatory service checks work through the bridge.
