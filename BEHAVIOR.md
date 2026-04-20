# Configuration Behavior & Priority

This document explains how `distributed-config` discovers, loads, and prioritizes configuration data across different environments and profiles.

## 1. The Strategy Profiles

The library uses **Strategies** to determine how to bootstrap and synchronize data.

| Profile | Purpose | Data Source | Sync Mode | File Discovery |
| :--- | :--- | :--- | :--- | :--- |
| **Standalone** | Local Development | Local YAML only - No server connection | Disabled | `standalone.yaml` |
| **Test** | CI/CD / Testing | Server (Full Sync) default 127.0.0.2 + Local YAML | GET & PUT | `test.yaml` |
| **Preprod** | Staging / QA | Prod Server (Read-only) + Local YAML | GET Only | `config_preprod.yaml` |
| **Production** | Live Deployment | Prod Server (Full Sync) + Local YAML | GET & PUT | `config.yaml` |

---

## 2. Parameter Integrity & Update Map

This table shows exactly which configuration sections are modified during each phase of the initialization.

| Phase | Source | **Common** | **Capabilities** | **MemConfig** |
| :--- | :--- | :---: | :---: | :---: |
| **1. Bootstrap** | Code Defaults | [x] | [x] | [ ] |
| **2. Environment** | `os.Getenv` | [x] | [x] | [ ] |
| **3. Handshake** | Server `GET_SYNC` | [x] | [ ] | [x] |
| **4. Local File** | authoritative YAML | [x] | [x] | [x] |
| **5. Runtime** | Server Broadcasts | [ ] | [ ] | [x] |

---

## 3. Detailed Strategy Flow

The search and initialization order differs slightly between strategies to balance automation vs. strictness.

### Standalone (Dev/Offline)
1. **Resolve**: `standalone.yaml`
2. **Bootstrap**: code defaults used only if file missing.
3. **Load**: reads file once.
4. **Expansion**: `os.Expand` resolves `${}` from env.

### Production (Authoritative)
1. **Env Load**: NAME, RESET, CF_IP, CF_PORT pulled from OS.
2. **Server Sync**: Handshake connects to CF_IP:CF_PORT and merges `Common.Name`.
3. **File Sync**: `config/config.yaml` is loaded. It **MUST** exist (skeleton generated if missing). **Always overwrites Step 2.**
4. **Runtime**: Background goroutine listens for `BROADCAST_SYNC` to update `MemConfig`.

### Test (CI/Automation)
1. **Resolve**: `test.yaml`
2. **Bootstrap**: code defaults (127.0.0.2) used if file missing.
3. **Env Override**: NAME, RESET, etc. override defaults.
4. **Server Sync**: Handshake merges `Common.Name`.
5. **Re-Load**: **CRITICAL** - The local file is re-read to ensure manual `test.yaml` settings override server data.
6. **Safety**: Integrity check fails if any IP != 127.0.0.2.

---

## 4. Static vs. Dynamic Parameters (Technical Breakdown)

There is a fundamental architectural difference between parameters that are fixed at startup and those that can change while the application is running.

### Static Parameters (`Common` & `Capabilities`)
- **Location**: Defined in the `common:` and `capabilities:` root sections of the YAML.
- **Behavior**: Loaded **once** during application startup (`New()`).
- **Authority**: The **Local File** is the absolute authority.
- **Updates**: These sections **cannot** be updated at runtime via the network. A restart is required to apply changes from the server or local file.

### Dynamic Parameters (`MemConfig`)
- **Location**: All other keys in the YAML file (outside of common/capabilities) or data pushed purely via the network.
- **Behavior**: Initially loaded from the file, but then kept in sync via a persistent socket connection.
- **Authority**: The **Config Server** is the authority at runtime.
- **Updates**: Updated **live** via `BROADCAST_SYNC` messages. Applications can register callbacks (`OnMemConfUpdate`) to react to these changes instantly without restarting.

---

## 4. Environment Variable Reference

### Reserved Variables
The following environment variables are explicitly checked during the `LoadCommonFromEnv` phase.

| Variable | Target Config Field | Description |
| :--- | :--- | :--- |
| `NAME` | `Common.Name` | Overrides the application name (used for Safe-Socket identity). |
| `RESET` | `Common.Reset` | If "true", signals the strategy to potentially reset local state. |
| `CF_IP` | `Capabilities["config_server"]["ip"]` | The IP address of the remote Config Server. |
| `CF_PORT` | `Capabilities["config_server"]["port"]` | The Port of the remote Config Server. |

### YAML Variable Expansion
You can use any environment variable inside your YAML files using the `${VAR:default}` syntax. These are expanded **after** the file is read but **before** it is decoded into the Go struct.

```yaml
capabilities:
  log_server:
    ip: "${LS_IP:127.0.0.1}"
    port: "${LS_PORT:9020}"
```

---

## 5. File Resolution (Search Order)

When looking for a configuration file for a profile (e.g., "production"), the `PathResolver` searches in this order:

1.  `config/config.yaml` (CWD)
2.  `config/config.yaml` (Executable Directory)
3.  `config/default.yaml` (CWD)
4.  `config/default.yaml` (Executable Directory)
5.  `config.yaml` (CWD)
6.  `config.yaml` (Executable Directory)
7.  `[ExecutableName].yaml` (Fallback)

---

## 6. Initialization Lifecycle

1.  **Instantiation**: `distributed_config.New("my-app", "production")` is called.
2.  **Bootstrap**: `core.NewDefaultConfig()` populates initial hardcoded defaults.
3.  **Environment Check**: `NAME`, `RESET`, etc., are pulled from the OS environment.
4.  **Handshake**: The client connects to the server and performs a `GET_SYNC`.
5.  **Local Override**: The resolver finds the best YAML file and decodes it, resolving `${}` variables.
6.  **Startup**: The application starts with the final merged state.
7.  **Sync**: A background goroutine maintains the socket connection for dynamic `MemConfig` updates.
