# Architecture: C++ SDK

The C++ SDK provides a high-level RAII wrapper for the C ABI.

## Design Patterns

- **RAII Lifecycle**: The `DistConfig` class manages the session handle. The destructor automatically calls `DistConf_Close`, preventing resource leaks.
- **Header-Only**: The implementation is contained entirely within `DistConf.hpp` for easy integration.
- **Exception Safety**: Methods throw `std::runtime_error` if initialization fails, following C++ best practices.
- **String Safety**: Uses `std::string` for all API boundaries. The internal bridge calls `DistConf_FreeString` after copying the data into the C++ string.

## Memory Management

Go-allocated strings are transferred to C++ strings and then immediately freed using the exported `DistConf_FreeString` method, ensuring no cross-language memory leaks.
