use distconf::DistConfig;
use serde_json::json;

fn main() -> Result<(), Box<dyn std::error::Error>> {
    println!("--- DistConf Rust SDK Demo ---");

    // 1. Initialize
    // Search for library in the SDK folder
    let lib_path = if cfg!(target_os = "macos") {
        "../libdistconf/libdistconf.dylib"
    } else {
        "../libdistconf/libdistconf.so"
    };

    let cfg = DistConfig::new("standalone", lib_path)?;

    // 2. Get static values
    let app_name = cfg.get("common", "name");
    println!("App Name: {}", app_name);

    // 3. Set local state
    cfg.set("rust_demo", "status", "ready")?;
    println!("Local Status: {}", cfg.get("rust_demo", "status"));

    // 4. Share object (Using serde_json macro)
    let payload = json!({
        "node_info": {
            "service": "rust-node",
            "uptime": 3600,
            "features": ["async", "safe"]
        }
    });
    
    if cfg.share_config(&payload).is_ok() {
        println!("Successfully shared node info.");
    }

    // 5. Validate environment
    if cfg.validate_mandatory_services().is_ok() {
        println!("Mandatory services validated.");
    }

    println!("--- Demo Completed ---");
    Ok(())
}
