# Configuration Behavior & Priority

This document explains how `distributed-config` discovers, loads, and prioritizes configuration data across different environments and profiles.

## 1. Strategy Profiles

The library uses **Strategies** to determine how to bootstrap and synchronize data.

| Profile | Purpose | Data Source | Sync Mode | File Discovery |
| :--- | :--- | :--- | :--- | :--- |
| **Standalone** | Local Dev / Isolated Apps | Local YAML only - No server connection | Disabled | `standalone.yaml` |
| **Test** | CI/CD / Automated Testing | Network Server + Local YAML (`127.0.0.2` enforced) | GET & PUT | `test.yaml` |
| **Staging** | Staging / QA | Network Server (Read-only) + Local YAML | GET Only | `staging.yaml` |
| **Production** | Live Deployment | Network Server (Full Sync) + Local YAML | GET & PUT | `production.yaml` |

> [!IMPORTANT]
> **Missing File Behavior (Prod/Staging)**: The configuration file is **optional**. If missing, the system proceeds using Environment variables (`CF_IP`, `CF_PORT`) to discover the Config Server.

---

## 2. Resolution Hierarchy (The 6-Step Path)

When searching for target config files (e.g., "staging"), the `PathResolver` follows this strict priority:

1. `config/staging.yaml` (Current Working Directory)
2. `config/staging.yaml` (Executable Directory)
3. `config/[ExecutableName].yaml` (Current Working Directory)
4. `config/[ExecutableName].yaml` (Executable Directory)
5. `[ExecutableName].yaml` (Current Working Directory)
6. `[ExecutableName].yaml` (Executable Directory)

*Note: Steps 1 & 2 are skipped if no profile is provided.*

---

## 3. Parameter Integrity & Merge Logic

| Phase | Action | Authority |
| :--- | :--- | :--- |
| **1. File Bootstrap**| Loading local properties | Local YAML |
| **2. Environment** | Dynamic process variables parsing | `os.Getenv` |
| **3. Server Baseline**| Initial HTTP/TCP payload merge | Server `GET_SYNC` |
| **4. File Override** | Re-asserting hierarchy | **File > Env > Server** |
| **5. Runtime** | Live Updates / Service Registry | `BROADCAST_SYNC` |

---

## 4. Strict Service Schemas

To ensure ecosystem robustness, the system will **fail-fast** if these mandatory services are missing from the configuration:

- **`log_server`**: Requires `ip` and `port`.
- **`config_server`**: Requires `ip` and `port`.

---

## 5. Environment Variables

The library supports dynamic expansion using `${VAR_NAME}` or `${VAR_NAME:default}`.

### Reserved System Variables
- `NAME`: System Distributed Identity.
- `CF_IP` / `CF_PORT`: Fallback connection parameters for the Config Server.
- `RESET`: If "true", triggers a local state reset.

---

## 6. Thread Safety (RCU)

All configuration reads are **100% lock-free** using an Atomic Pointer Swap (Read-Copy-Update) mechanism. This ensures zero latency for high-frequency microservices, even while background synchronization is updating the configuration snapshot.
