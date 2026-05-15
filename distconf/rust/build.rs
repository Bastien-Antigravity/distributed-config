use std::process::Command;
use std::path::Path;
use std::env;

fn main() {
    let root_dir = env::current_dir().unwrap().parent().unwrap().parent().unwrap().to_path_buf();
    let lib_dir = root_dir.join("distconf").join("libdistconf");
    
    let ext = if cfg!(target_os = "macos") {
        "dylib"
    } else if cfg!(target_os = "windows") {
        "dll"
    } else {
        "so"
    };
    
    let lib_name = format!("libdistconf.{}", ext);
    let lib_path = lib_dir.join(&lib_name);

    if !lib_path.exists() {
        println!("cargo:warning=Shared library missing, attempting to build via root Makefile...");
        let status = Command::new("make")
            .arg("-C")
            .arg(&root_dir)
            .arg("build-lib")
            .status()
            .expect("Failed to execute make command");

        if !status.success() {
            panic!("Failed to build libdistconf via root Makefile");
        }
    }

    // Tell cargo to rerun if the library changes
    println!("cargo:rerun-if-changed={}", lib_path.display());
}
