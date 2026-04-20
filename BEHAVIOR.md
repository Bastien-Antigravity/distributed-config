# Configuration Behavior & Priority

This document explains how `distributed-config` discovers, loads, and prioritizes configuration data across different environments and profiles. It acts as the definitive guide to understanding system state behaviors during cold-starts, network synchronizations, and hot-swaps.

## 1. The Strategy Profiles

The library uses **Strategies** to determine how to bootstrap and synchronize data.

| Profile | Purpose | Data Source | Sync Mode | File Discovery |
| :--- | :--- | :--- | :--- | :--- |
| **Standalone** | Local Dev / Isolated Apps | Local YAML only - No server connection | Disabled | `standalone.yaml` |
| **Test** | CI/CD / Automated Testing | Network Server + Local YAML (`127.0.0.2` enforced) | GET & PUT | `test.yaml` |
| **Staging** | Staging / QA | Network Server (Read-only) + Local YAML | GET Only | `staging.yaml` |
| **Production** | Live Deployment | Network Server (Full Sync) + Local YAML | GET & PUT | `production.yaml` |

> [!IMPORTANT]
> **Missing File Behavior (Prod/Staging)**: The configuration file is **optional**. If missing, the system proceeds using Environment variables (`CF_IP`, `CF_PORT`) to discover the Config Server. It **DOES NOT** generate skeletons or new files for these profiles. However, final validation still enforces that all mandatory parameters are eventually satisfied via the network.

---

## 2. Parameter Integrity & Memory Map

This table denotes the state of specific internal nodes throughout the initialization sequence.

| Phase | Action | Source | `Common` | `Capabilities` | `LiveConfig` |
| :--- | :--- | :--- | :---: | :---: | :---: |
| **1. File Bootstrap**| Loading local properties via `PathResolver` | Local YAML | [x] | [x] | [x] |
| **2. Environment** | Dynamic process variables parsing | `os.Getenv` | [x] | [x] | [ ] |
| **3. Server Baseline**| Initial HTTP/TCP TCP-Hello payload merge | Server `GET_SYNC` | [x] | [ ] | [x] |
| **4. File Override** | Re-parsing to assert `File > Server` hierarchy | Local YAML | [x] | [x] | [ ] |
| **5. Runtime** | Ephemeral updates / Service Registry | `BROADCAST_SYNC`| [ ] | [ ] | [x] |

---

## 3. Strict Service Schemas & Validation

To ensure ecosystem robustness, `distributed-config` enforces **Strict Schema Validation** for all core services. The system will fail-fast and reject any configuration that lacks "Minimal Required Parameters" for foundational infrastructure.

### Mandatory Core Services

The following services are considered **Mandatory**. If they are missing from your `Capabilities` map or lack the required keys, the application will fail to start:

| Service | Required Parameters |
| :--- | :--- |
| **`log_server`** | `ip`, `port` |
| **`config_server`** | `ip`, `port` |

> [!IMPORTANT]
> The system strictly enforces these schemas. For instance, attempting to define a `log_server` without a `port` will trigger a startup panic: `log_server: ip and port are mandatory`.

### Extended Service Schemas

While not always mandatory for boot, the following schemas are pre-defined for consistent usage across the ecosystem:

- **`tele_remote`**: `token`, `chat_id`, `ip`, `port`
- **`timescale_db`**: `ip`, `port`, `db_name`, `user`, `password`
- **`file_system`**: `temp_path`, `data_path`
- **`web_interface`**, **`scheduler`**, **`jupyter`**: `ip`, `port`

---

## 4. Parameter Segregation & Permissions

Configuration arguments are partitioned into strict domains controlling how network layers are allowed to observe them.

### `Common` & `Capabilities` (Static Application State)
- **State**: Root maps (`common:`, `capabilities:`).
- **Behavior**: Single-pass loaded.
- **Authority**: Local file takes unconditional precedence over network values.
- **Updates**: Cannot be altered dynamically during runtime securely without specific `LiveConfig` hot-swap mappings!

### `LiveConfig` (Dynamic Remote State)
- **State**: Root keys outside reserved maps dynamically constructed.
- **Behavior**: Inherently transient. Live streaming capable.
- **Authority**: Pushed dynamically from `config-server`.
- **Updates**: Injects live parameters triggering arbitrary `OnLiveConfUpdate()` callbacks.

### Private & External Configurations
The `Config` system strictly enforces separation by detaching private structs off the core structural schemas.
Use `PrivateConfig` to define the **Identity and Local Requirements** of your specific microservice:

```yaml
name: "my-service"
private_file_path: "config/private.yaml"
private:
  local_buffer_size: "1024"
```

To load these isolation layers, use `loader.LoadYAML`:
```go
import "github.com/Bastien-Antigravity/distributed-config/src/loader"

var myLocal core.PrivateConfig
err := loader.LoadYAML("config/private.yaml", &myLocal)
```

---

## 5. Environment Variable Reference

### Reserved Variables
| Variable | Target Config Field | Description |
| :--- | :--- | :--- |
| `NAME` | `Common.Name` | Evaluates identically to the System Distributed Identity; effectively acts as network identity identifier inside safe-sockets. |
| `RESET` | `Common.Reset` | If "true", signals the strategy to potentially reset local state. |
| `CF_IP` | `Capabilities["config_server"]["ip"]` | The IP address of the remote Config Server fallback. |
| `CF_PORT` | `Capabilities["config_server"]["port"]` | The Port of the remote Config Server fallback. |

---

## 7. File Resolution Sequence

When searching for target config files (example target: "staging"), the PathResolver strictly falls back utilizing these paths ensuring no wildcard ambiguity:

1. `config/staging.yaml` (CWD)
2. `config/staging.yaml` (Executable Directory)
3. `config/[ExecutableName].yaml` (CWD)
4. `config/[ExecutableName].yaml` (Executable Directory)
5. `[ExecutableName].yaml` (CWD)
6. `[ExecutableName].yaml` (Executable Directory)

*Note: Steps 1 & 2 are skipped if no profile (targetName) is provided.*
