# Architecture: Python SDK

The Python SDK acts as a thin FFI layer over `libdistconf`.

## Design Patterns

- **Handle-based Lifecycle**: The `DistConfig` class manages a `uintptr_t` handle. It ensures `DistConf_Close` is called via the `__del__` destructor (RAII-lite).
- **ctypes Mapping**: Standard `ctypes` mappings are used for all C exports.
- **String Management**: The wrapper handles the conversion between Python `str` and Go `c_char_p`. It explicitly calls `DistConf_FreeString` to prevent memory leaks from strings returned by Go.

## Data Flow

1. Python calls `DistConf_New`.
2. Go Core initializes the configuration session and returns an ID.
3. Python stores the ID and uses it for all subsequent calls.
4. Go methods return JSON strings for complex types, which Python parses using `json.loads`.
