use libc::{c_char, uintptr_t, c_int};
use std::ffi::{CStr, CString};
use std::ptr;
use std::sync::Arc;
use libloading::{Library, Symbol};
use serde_json::Value;

pub type ConfigUpdateCb = extern "C" fn(json_data: *const c_char);

pub struct DistConfig {
    lib: Arc<Library>,
    handle: uintptr_t,
}

impl DistConfig {
    pub fn new(profile: &str, lib_path: &str) -> Result<Self, Box<dyn std::error::Error>> {
        let lib = unsafe { Arc::new(Library::new(lib_path)?) };
        
        let handle = unsafe {
            let func: Symbol<unsafe extern "C" fn(*const c_char) -> uintptr_t> = lib.get(b"DistConf_New")?;
            let profile_c = CString::new(profile)?;
            func(profile_c.as_ptr())
        };

        if handle == 0 {
            return Err("Failed to initialize DistConf".into());
        }

        Ok(DistConfig { lib, handle })
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

    pub fn set(&self, section: &str, key: &str, value: &str) {
        unsafe {
            let func: Symbol<unsafe extern "C" fn(uintptr_t, *const c_char, *const c_char, *const c_char)> = 
                self.lib.get(b"DistConf_Set").unwrap();
            let section_c = CString::new(section).unwrap();
            let key_c = CString::new(key).unwrap();
            let value_c = CString::new(value).unwrap();
            func(self.handle, section_c.as_ptr(), key_c.as_ptr(), value_c.as_ptr());
        }
    }

    pub fn sync(&self) -> bool {
        unsafe {
            let func: Symbol<unsafe extern "C" fn(uintptr_t) -> c_int> = self.lib.get(b"DistConf_Sync").unwrap();
            func(self.handle) == 1
        }
    }

    pub fn share_object(&self, section: &str, payload: &Value) -> bool {
        unsafe {
            let func: Symbol<unsafe extern "C" fn(uintptr_t, *const c_char, *const c_char) -> c_int> = 
                self.lib.get(b"DistConf_ShareObject").unwrap();
            let section_c = CString::new(section).unwrap();
            let json_data = CString::new(payload.to_string()).unwrap();
            func(self.handle, section_c.as_ptr(), json_data.as_ptr()) == 1
        }
    }

    pub fn validate_mandatory_services(&self) -> bool {
        unsafe {
            let func: Symbol<unsafe extern "C" fn(uintptr_t) -> c_int> = 
                self.lib.get(b"DistConf_ValidateMandatoryServices").unwrap();
            func(self.handle) == 1
        }
    }

    pub fn get_address(&self, capability: &str) -> String {
        unsafe {
            let func: Symbol<unsafe extern "C" fn(uintptr_t, *const c_char) -> *mut c_char> = 
                self.lib.get(b"DistConf_GetAddress").unwrap();
            let cap_c = CString::new(capability).unwrap();
            let res = func(self.handle, cap_c.as_ptr());
            if res.is_null() {
                return "".to_string();
            }
            let val = CStr::from_ptr(res).to_string_lossy().into_owned();
            self.free_string(res);
            val
        }
    }

    pub fn decrypt(&self, ciphertext: &str) -> String {
        unsafe {
            let func: Symbol<unsafe extern "C" fn(uintptr_t, *const c_char) -> *mut c_char> = 
                self.lib.get(b"DistConf_Decrypt").unwrap();
            let cipher_c = CString::new(ciphertext).unwrap();
            let res = func(self.handle, cipher_c.as_ptr());
            if res.is_null() {
                return ciphertext.to_string();
            }
            let val = CStr::from_ptr(res).to_string_lossy().into_owned();
            self.free_string(res);
            val
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
        let mut path = "../../release/libdistconf.so".to_string();
        if !Path::new(&path).exists() {
            path = path.replace(".so", ".dylib");
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
        let lib_path = get_lib_path();
        let cfg = DistConfig::new("standalone", &lib_path).unwrap();
        cfg.set("rust_test", "key", "val");
        assert_eq!(cfg.get("rust_test", "key"), "val");
    }
}
