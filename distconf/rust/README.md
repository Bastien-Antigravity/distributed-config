# DistConf Rust SDK

Safe Rust bindings for the `distributed-config` ecosystem.

## Usage

```rust
use distconf::DistConfig;

fn main() {
    let cfg = DistConfig::new("standalone", "libdistconf.so").unwrap();
    let name = cfg.get("common", "name");
    cfg.set("local", "status", "active");
}
```

## Testing

```bash
cargo test
```
