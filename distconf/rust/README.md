# DistConf Rust SDK

Safe Rust bindings for the `distributed-config` ecosystem.

## Usage

```rust
use distconf::{DistConfig, ConfigUpdateCb};
use serde_json::json;
use std::ffi::CStr;
use libc::{uintptr_t, c_char};

extern "C" fn on_update(handle: uintptr_t, json_data: *const c_char) {
    // Handle update
}

fn main() {
    let cfg = DistConfig::new("standalone", "../libdistconf/libdistconf.so").unwrap();
    
    cfg.on_live_conf_update(on_update);
    cfg.on_registry_update(on_update);

    let name = cfg.get("common", "name");
    cfg.set("local", "status", "active");

    cfg.share_config(&json!({"status": "healthy"}));
}
```

## Testing

```bash
cargo test
```
