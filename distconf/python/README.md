# DistConf Python SDK

A native Python wrapper for the `distributed-config` ecosystem.

## Installation

Ensure `libdistconf.so` is available and set the environment variable:
```bash
export LIBDISTCONF_PATH=/path/to/release/libdistconf.so
```

## Usage

```python
from distconf import DistConfig

# Initialize
cfg = DistConfig("standalone")

# Get/Set
val = cfg.get("common", "name")
success = cfg.set("local", "status", "ready")

# Event Callbacks
cfg.on_live_conf_update(lambda updates: print("Config changed:", updates))
cfg.on_registry_update(lambda registry: print("Registry changed:", registry))

# Broadcast state
cfg.share_config({"status": "healthy", "metrics": {"cpu": 45}})

# Sync and Validate
if cfg.validate_mandatory_services():
    print("Environment OK")

# Cleanup
cfg.close()
```

## Testing

Run tests using `pytest` or `unittest`:
```bash
python3 -m unittest tests/test_distconf.py
```
