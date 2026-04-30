# Project TODOs

### 1. hot swapping of system components (capabilities):
find a way to hot swap system components (capabilities) without restarting the system.

### 2. Refactor PEM file loading: [DONE]
RSA keys now support configurable paths via the `BASTIEN_PRIVATE_KEY_PATH` environment variable, with standard fallbacks to `/etc/bastien/private.pem` and `./private.pem`. Ported to Go, Python, and Rust.

### 3. Release Artifact Management: [IN PROGRESS]
Standardize the distribution of `config-tool` binaries and update `.gitignore` to allow tracking of stable release artifacts in the `release/` directory.
