# Architecture: Rust SDK

Safe, idiomatic Rust bindings for `libdistconf`.

## Design Patterns

- **RAII (Drop Trait)**: The `DistConfig` struct implements `Drop`. When it goes out of scope, the bridge session is automatically closed.
- **Dynamic Loading**: Uses `libloading` to link with the shared library at runtime, allowing the same binary to work with different library versions.
- **Thread Safety**: The `DistConfig` struct can be wrapped in `Arc` for sharing across threads.
- **String Management**: Handles conversion between `&str` and `C-strings`. Automatically calls `DistConf_FreeString` after capturing Go data into a Rust `String`.

## Error Handling

Uses standard `Result` and `Option` types to handle initialization failures or missing configuration keys.
