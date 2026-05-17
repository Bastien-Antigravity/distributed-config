use libc::{c_char, uintptr_t, c_int};
use std::ffi::{CStr, CString};
use libloading::{Library, Symbol};
use serde_json::Value;
use std::fmt;

pub type ConfigUpdateCb = extern "C" fn(handle: uintptr_t, json_data: *const c_char);

// Standardized Error Codes (must match helpers.h)
pub const DISTCONF_SUCCESS: i32 = 0;
pub const DISTCONF_ERR_GENERIC: i32 = 1;
pub const DISTCONF_ERR_INVALID_HANDLE: i32 = 2;
pub const DISTCONF_ERR_KEY_NOT_FOUND: i32 = 3;
pub const DISTCONF_ERR_VALIDATION_FAILED: i32 = 4;
pub const DISTCONF_ERR_NETWORK_FAILURE: i32 = 5;
pub const DISTCONF_ERR_DECRYPTION_FAILED: i32 = 6;
pub const DISTCONF_ERR_INVALID_INPUT: i32 = 7;

#[derive(Debug)]
pub struct DistConfError {
    pub message: String,
    pub code: i32,
}

impl fmt::Display for DistConfError {
    fn fmt(&self, f: &mut fmt::Formatter) -> fmt::Result {
        write!(f, "DistConf Error ({}): {}", self.code, self.message)
    }
}

impl std::error::Error for DistConfError {}

pub struct DistConfig {
    lib: &'static Library,
    handle: uintptr_t,
}

impl DistConfig {
    pub fn new(profile: &str, lib_path: &str) -> Result<Self, Box<dyn std::error::Error>> {
        // We use a static reference and leak the library because Go's runtime 
        // does not support being unloaded (dlclose) and will hang.
        let lib = Box::leak(Box::new(unsafe { Library::new(lib_path)? }));
        
        let handle = unsafe {
            let func: Symbol<unsafe extern "C" fn(*const c_char) -> uintptr_t> = lib.get(b"DistConf_New")?;
            let profile_c = CString::new(profile)?;
            func(profile_c.as_ptr())
        };

        if handle == 0 {
            return Err(Self::get_last_error_static(lib).into());
        }

        Ok(DistConfig { lib, handle })
    }

    fn get_last_error_static(lib: &Library) -> DistConfError {
        unsafe {
            let get_code: Symbol<unsafe extern "C" fn() -> c_int> = lib.get(b"DistConf_GetLastErrorCode").unwrap();
            let get_msg: Symbol<unsafe extern "C" fn() -> *const c_char> = lib.get(b"DistConf_GetLastError").unwrap();
            
            let code = get_code();
            let msg_ptr = get_msg();
            let message = if msg_ptr.is_null() {
                "Unknown error".to_string()
            } else {
                CStr::from_ptr(msg_ptr).to_string_lossy().into_owned()
            };
            
            DistConfError { message, code }
        }
    }

    fn raise_last_error(&self) -> DistConfError {
        Self::get_last_error_static(self.lib)
    }

    pub fn get(&self, section: &str, key: &str) -> String {
        unsafe {
            let func: Symbol<unsafe extern "C" fn(uintptr_t, *const c_char, *const c_char) -> *mut c_char> = 
                self.lib.get(b"DistConf_Get").unwrap();
            let section_c = CString::new(section).unwrap();
            let key_c = CString::new(key).unwrap();
            let res = func(self.handle, section_c.as_ptr(), key_c.as_ptr());
            if res.is_null() {
                return "".to_string();
            }
            let val = CStr::from_ptr(res).to_string_lossy().into_owned();
            self.free_string(res);
            val
        }
    }

    pub fn set(&self, section: &str, key: &str, value: &str) -> Result<(), DistConfError> {
        unsafe {
            let func: Symbol<unsafe extern "C" fn(uintptr_t, *const c_char, *const c_char, *const c_char) -> c_int> = 
                self.lib.get(b"DistConf_Set").unwrap();
            let section_c = CString::new(section).unwrap();
            let key_c = CString::new(key).unwrap();
            let value_c = CString::new(value).unwrap();
            if func(self.handle, section_c.as_ptr(), key_c.as_ptr(), value_c.as_ptr()) == 0 {
                return Err(self.raise_last_error());
            }
            Ok(())
        }
    }

    pub fn sync(&self) -> Result<(), DistConfError> {
        unsafe {
            let func: Symbol<unsafe extern "C" fn(uintptr_t) -> c_int> = self.lib.get(b"DistConf_Sync").unwrap();
            if func(self.handle) == 0 {
                return Err(self.raise_last_error());
            }
            Ok(())
        }
    }

    pub fn share_config(&self, payload: &Value) -> Result<(), DistConfError> {
        unsafe {
            let func: Symbol<unsafe extern "C" fn(uintptr_t, *const c_char) -> c_int> = 
                self.lib.get(b"DistConf_ShareConfig").unwrap();
            let json_data = CString::new(payload.to_string()).unwrap();
            if func(self.handle, json_data.as_ptr()) == 0 {
                return Err(self.raise_last_error());
            }
            Ok(())
        }
    }

    pub fn on_live_conf_update(&self, cb: ConfigUpdateCb) {
        unsafe {
            let func: Symbol<unsafe extern "C" fn(uintptr_t, ConfigUpdateCb)> = 
                self.lib.get(b"DistConf_OnLiveConfUpdate").unwrap();
            func(self.handle, cb);
        }
    }

    pub fn on_registry_update(&self, cb: ConfigUpdateCb) {
        unsafe {
            let func: Symbol<unsafe extern "C" fn(uintptr_t, ConfigUpdateCb)> = 
                self.lib.get(b"DistConf_OnRegistryUpdate").unwrap();
            func(self.handle, cb);
        }
    }

    pub fn validate_mandatory_services(&self) -> Result<(), DistConfError> {
        unsafe {
            let func: Symbol<unsafe extern "C" fn(uintptr_t) -> c_int> = 
                self.lib.get(b"DistConf_ValidateMandatoryServices").unwrap();
            if func(self.handle) == 0 {
                return Err(self.raise_last_error());
            }
            Ok(())
        }
    }

    pub fn get_address(&self, capability: &str) -> Result<String, DistConfError> {
        unsafe {
            let func: Symbol<unsafe extern "C" fn(uintptr_t, *const c_char) -> *mut c_char> = 
                self.lib.get(b"DistConf_GetAddress").unwrap();
            let cap_c = CString::new(capability).unwrap();
            let res = func(self.handle, cap_c.as_ptr());
            if res.is_null() {
                return Err(self.raise_last_error());
            }
            let val = CStr::from_ptr(res).to_string_lossy().into_owned();
            self.free_string(res);
            Ok(val)
        }
    }

    pub fn decrypt(&self, ciphertext: &str) -> Result<String, DistConfError> {
        unsafe {
            let func: Symbol<unsafe extern "C" fn(uintptr_t, *const c_char) -> *mut c_char> = 
                self.lib.get(b"DistConf_Decrypt").unwrap();
            let cipher_c = CString::new(ciphertext).unwrap();
            let res = func(self.handle, cipher_c.as_ptr());
            if res.is_null() {
                return Err(self.raise_last_error());
            }
            let val = CStr::from_ptr(res).to_string_lossy().into_owned();
            self.free_string(res);
            Ok(val)
        }
    }

    fn free_string(&self, ptr: *mut c_char) {
        unsafe {
            let func: Symbol<unsafe extern "C" fn(*mut c_char)> = self.lib.get(b"DistConf_FreeString").unwrap();
            func(ptr);
        }
    }
}

impl Drop for DistConfig {
    fn drop(&mut self) {
        unsafe {
            if let Ok(func) = self.lib.get::<unsafe extern "C" fn(uintptr_t)>(b"DistConf_Close") {
                func(self.handle);
            }
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use std::path::Path;

    fn get_lib_path() -> String {
        if let Ok(path) = std::env::var("LIBDISTCONF_PATH") {
            return path;
        }
        let mut path = "../libdistconf/libdistconf.so".to_string();
        if !Path::new(&path).exists() {
            if cfg!(target_os = "macos") {
                path = path.replace(".so", ".dylib");
            } else if cfg!(target_os = "windows") {
                path = path.replace(".so", ".dll");
            }
        }
        path
    }

    #[test]
    fn test_lifecycle() {
        let lib_path = get_lib_path();
        let cfg = DistConfig::new("standalone", &lib_path).unwrap();
        assert!(cfg.handle != 0);
    }

    #[test]
    fn test_get_set() {
        println!("Starting test_get_set");
        let lib_path = get_lib_path();
        println!("Loading DistConfig...");
        let cfg = DistConfig::new("standalone", &lib_path).unwrap();
        println!("DistConfig loaded. Calling set()...");
        cfg.set("rust_test", "key", "val").expect("set failed");
        println!("set() returned. Calling get()...");
        let val = cfg.get("rust_test", "key");
        println!("get() returned: {}", val);
        assert_eq!(val, "val");
    }
}
