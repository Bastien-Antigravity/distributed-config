# Architecture: libdistconf Bridge

The `libdistconf` bridge acts as the "Point of Truth" for the entire polyglot microservice ecosystem.

## Layered Design

1.  **Go Core Engine**: The same robust logic used in Go services.
2.  **Facade Store**: A thread-safe registry in Go that tracks multiple `Config` sessions using opaque `uintptr` handles.
3.  **CGO Export Layer**: Converts Go types (slices, maps, structs) into C-compatible types (`char*`, `int`).
4.  **Shared Library**: A self-contained binary artifact that can be loaded via standard dynamic linking (FFI).

## Thread Safety

All bridge methods are protected by a global mutex in the Facade Store. However, individual `Config` sessions are designed to be thread-safe for reading, while updates trigger atomic swaps of the internal configuration state.

## Memory Safety

Because Go uses a Garbage Collector and C does not, any string returned from Go to C must be explicitly allocated and then explicitly freed.
*   **The Go Side**: Allocates strings using `C.CString`.
*   **The C Side**: Must call `DistConf_FreeString` once it has finished copying the data.
*   **The SDK Side**: All provided SDKs (Python, Rust, C++) handle this automatically via RAII or explicit cleanup helpers.

## JSON as the "Universal Glue"

To avoid complex struct mapping in FFI (which is brittle), the bridge uses JSON strings for all complex data structures (Capabilities, Full Config). This allows the Go core to evolve its internal data structures without breaking the binary interface (ABI) of the shared library.
