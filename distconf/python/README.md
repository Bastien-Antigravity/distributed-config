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
cfg.set("local", "status", "ready")

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
