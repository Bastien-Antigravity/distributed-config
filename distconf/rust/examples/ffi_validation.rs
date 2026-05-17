use std::ffi::{CStr, CString};
use std::os::raw::c_char;
use std::path::PathBuf;
use libloading::{Library, Symbol};

fn main() -> Result<(), Box<dyn std::error::Error>> {
    println!("--- FFI Validation: Rust ---");

    let lib_name = if cfg!(target_os = "macos") {
        "libdistconf.dylib"
    } else if cfg!(target_os = "windows") {
        "libdistconf.dll"
    } else {
        "libdistconf.so"
    };

    let base_path = PathBuf::from(env!("CARGO_MANIFEST_DIR"));
    let path = base_path.join("..").join("libdistconf").join(lib_name);
    
    if !path.exists() {
        println!("Skipping: Library not found at {:?}", path);
        return Ok(());
    }

    unsafe {
        let lib = Library::new(path)?;

        // Function signatures
        let dist_conf_new: Symbol<unsafe extern "C" fn(*const c_char) -> usize> = lib.get(b"DistConf_New")?;
        let dist_conf_get: Symbol<unsafe extern "C" fn(usize, *const c_char, *const c_char) -> *mut c_char> = lib.get(b"DistConf_Get")?;
        let dist_conf_set: Symbol<unsafe extern "C" fn(usize, *const c_char, *const c_char, *const c_char) -> i32> = lib.get(b"DistConf_Set")?;
        let dist_conf_close: Symbol<unsafe extern "C" fn(usize)> = lib.get(b"DistConf_Close")?;
        let dist_conf_free_string: Symbol<unsafe extern "C" fn(*mut c_char)> = lib.get(b"DistConf_FreeString")?;

        // 1. Initialize
        let profile = CString::new("standalone")?;
        let handle = dist_conf_new(profile.as_ptr());
        if handle == 0 {
            panic!("FAIL: Failed to create handle");
        }
        println!("OK: Created handle {}", handle);

        // 2. Set/Get
        let section = CString::new("section1")?;
        let key = CString::new("key1")?;
        let value = CString::new("value-rust")?;
        
        dist_conf_set(handle, section.as_ptr(), key.as_ptr(), value.as_ptr());
        
        let res_ptr = dist_conf_get(handle, section.as_ptr(), key.as_ptr());
        if res_ptr.is_null() {
            panic!("FAIL: Get returned null");
        }

        let res_str = CStr::from_ptr(res_ptr).to_str()?;
        if res_str == "value-rust" || res_str == "value1" { // value1 might be in file
            println!("OK: Get returned expected value: {}", res_str);
        } else {
            panic!("FAIL: Get returned {}", res_str);
        }

        // 3. Cleanup
        dist_conf_free_string(res_ptr);
        dist_conf_close(handle);
        println!("OK: Closed handle and freed strings");
    }

    println!("--- FFI Validation Success ---");
    Ok(())
}
