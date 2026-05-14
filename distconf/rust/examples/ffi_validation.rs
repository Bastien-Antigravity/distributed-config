use std::ffi::{CStr, CString};
use std::os::raw::{c_char, c_void};
use std::path::PathBuf;
use libloading::{Library, Symbol};

fn main() -> Result<(), Box<dyn std::error::Error>> {
    println!("--- FFI Validation: Rust ---");

    let lib_path = if cfg!(target_os = "macos") {
        "../../release/libdistconf.dylib"
    } else if cfg!(target_os = "windows") {
        "../../release/libdistconf.dll"
    } else {
        "../../release/libdistconf.so"
    };

    let path = PathBuf::from(lib_path);
    if !path.exists() {
        println!("Skipping: Library not found at {:?}", path);
        return Ok(());
    }

    unsafe {
        let lib = Library::new(path)?;

        type NewFn = unsafe extern "C" fn(*const c_char) -> *mut c_void;
        type GetFn = unsafe extern "C" fn(*mut c_void, *const c_char, *const c_char) -> *const c_char;
        type SetFn = unsafe extern "C" fn(*mut c_void, *const c_char, *const c_char, *const c_char) -> i32;
        type CloseFn = unsafe extern "C" fn(*mut c_void);

        let distconf_new: Symbol<NewFn> = lib.get(b"DistConf_New")?;
        let distconf_get: Symbol<GetFn> = lib.get(b"DistConf_Get")?;
        let distconf_set: Symbol<SetFn> = lib.get(b"DistConf_Set")?;
        let distconf_close: Symbol<CloseFn> = lib.get(b"DistConf_Close")?;

        // 1. Initialize
        let profile = CString::new("standalone")?;
        let handle = distconf_new(profile.as_ptr());
        if handle.is_null() {
            panic!("FAIL: Failed to create handle");
        }
        println!("OK: Created handle {:?}", handle);

        // 2. Set/Get
        let section = CString::new("section1")?;
        let key = CString::new("key1")?;
        let value = CString::new("value1")?;
        
        distconf_set(handle, section.as_ptr(), key.as_ptr(), value.as_ptr());
        
        let res_ptr = distconf_get(handle, section.as_ptr(), key.as_ptr());
        if res_ptr.is_null() {
            panic!("FAIL: Get returned null");
        }
        
        let res = CStr::from_ptr(res_ptr).to_str()?;
        if res == "value1" {
            println!("OK: Get returned expected value: {}", res);
        } else {
            panic!("FAIL: Get returned {}", res);
        }

        // 3. Close
        distconf_close(handle);
        println!("OK: Closed handle");
    }

    println!("--- FFI Validation Success ---");
    Ok(())
}
