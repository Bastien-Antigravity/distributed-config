# Project TODOs

### 1. Hot swapping of system components (capabilities): [IN PROGRESS]
Now that the **Atomic Pointer Swap** architecture is in place, implementing hot-swapping of capabilities is significantly easier. We can now swap entire capability maps without blocking readers. Next step: Implement the "Capability Reload" trigger.

### 2. Release Artifact Management: [IN PROGRESS]
Standardize the distribution of `config-tool` binaries and update `.gitignore` to allow tracking of stable release artifacts in the `release/` directory.

### 3. Dynamic Key Discovery: [TODO]
Implement a "smart" context-aware key discovery mechanism. The goal is to support multi-tenancy on the same host (multiple distributed systems) by dynamically determining the private/public key paths based on the `common.name` in the configuration or other unique identifiers, without adding excessive complexity to the decryption engine.

### 4. Create service_schemas capabillities regarding config type (standalone, test....: [TODO]

