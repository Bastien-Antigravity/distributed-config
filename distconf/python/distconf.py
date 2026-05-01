import os
import ctypes
from ctypes import c_char_p, c_void_p, CFUNCTYPE, c_int
import json
from typing import Any, Callable, Dict, Optional

# Callback type: void (*config_update_cb)(uintptr_t handle, const char* json_data)
CALLBACK_TYPE = CFUNCTYPE(None, c_void_p, c_char_p)

class DistConfig:
    """
    Python wrapper for libdistconf.
    Provides native access to the distributed configuration ecosystem.
    """
    
    def __init__(self, profile: str, lib_path: Optional[str] = None):
        self._lib = self._load_lib(lib_path)
        if not self._lib:
            raise RuntimeError("Could not load libdistconf shared library")
            
        self._handle = self._lib.DistConf_New(profile.encode('utf-8'))
        if not self._handle:
            raise RuntimeError(f"Failed to initialize DistConf with profile: {profile}")
            
        self._callback_ref = None

    def _load_lib(self, lib_path: Optional[str]):
        if not lib_path:
            lib_path = os.getenv("LIBDISTCONF_PATH", "libdistconf.so")
            
        try:
            lib = ctypes.CDLL(lib_path)
            
            # Signatures
            lib.DistConf_New.argtypes = [c_char_p]
            lib.DistConf_New.restype = c_void_p
            
            lib.DistConf_Close.argtypes = [c_void_p]
            lib.DistConf_Close.restype = None
            
            lib.DistConf_Get.argtypes = [c_void_p, c_char_p, c_char_p]
            lib.DistConf_Get.restype = c_void_p
            
            lib.DistConf_Set.argtypes = [c_void_p, c_char_p, c_char_p, c_char_p]
            lib.DistConf_Set.restype = None
            
            lib.DistConf_Sync.argtypes = [c_void_p]
            lib.DistConf_Sync.restype = c_int
            
            lib.DistConf_OnLiveConfUpdate.argtypes = [c_void_p, CALLBACK_TYPE]
            lib.DistConf_OnLiveConfUpdate.restype = None

            lib.DistConf_OnRegistryUpdate.argtypes = [c_void_p, CALLBACK_TYPE]
            lib.DistConf_OnRegistryUpdate.restype = None
            
            lib.DistConf_ShareConfig.argtypes = [c_void_p, c_char_p]
            lib.DistConf_ShareConfig.restype = c_int
            
            lib.DistConf_ValidateMandatoryServices.argtypes = [c_void_p]
            lib.DistConf_ValidateMandatoryServices.restype = c_int
            
            lib.DistConf_GetAddress.argtypes = [c_void_p, c_char_p]
            lib.DistConf_GetAddress.restype = c_void_p
            
            lib.DistConf_GetGRPCAddress.argtypes = [c_void_p, c_char_p]
            lib.DistConf_GetGRPCAddress.restype = c_void_p
            
            lib.DistConf_GetCapability.argtypes = [c_void_p, c_char_p]
            lib.DistConf_GetCapability.restype = c_void_p
            
            lib.DistConf_GetFullConfig.argtypes = [c_void_p]
            lib.DistConf_GetFullConfig.restype = c_void_p
            
            lib.DistConf_Decrypt.argtypes = [c_void_p, c_char_p]
            lib.DistConf_Decrypt.restype = c_void_p
            
            lib.DistConf_FreeString.argtypes = [c_void_p]
            lib.DistConf_FreeString.restype = None
            
            return lib
        except Exception as e:
            print(f"Error loading libdistconf: {e}")
            return None

    def get(self, section: str, key: str) -> str:
        ptr = self._lib.DistConf_Get(self._handle, section.encode('utf-8'), key.encode('utf-8'))
        if not ptr:
            return ""
        val = ctypes.string_at(ptr).decode('utf-8')
        self._lib.DistConf_FreeString(ptr)
        return val

    def set(self, section: str, key: str, value: str):
        self._lib.DistConf_Set(self._handle, section.encode('utf-8'), key.encode('utf-8'), value.encode('utf-8'))

    def sync(self) -> bool:
        return self._lib.DistConf_Sync(self._handle) == 1

    def share_config(self, payload: Any) -> bool:
        json_data = json.dumps(payload)
        return self._lib.DistConf_ShareConfig(self._handle, json_data.encode('utf-8')) == 1

    def validate_mandatory_services(self) -> bool:
        return self._lib.DistConf_ValidateMandatoryServices(self._handle) == 1

    def get_address(self, capability: str) -> str:
        ptr = self._lib.DistConf_GetAddress(self._handle, capability.encode('utf-8'))
        if not ptr:
            return ""
        val = ctypes.string_at(ptr).decode('utf-8')
        self._lib.DistConf_FreeString(ptr)
        return val

    def get_grpc_address(self, capability: str) -> str:
        ptr = self._lib.DistConf_GetGRPCAddress(self._handle, capability.encode('utf-8'))
        if not ptr:
            return ""
        val = ctypes.string_at(ptr).decode('utf-8')
        self._lib.DistConf_FreeString(ptr)
        return val

    def get_capability(self, capability: str) -> Dict[str, Any]:
        ptr = self._lib.DistConf_GetCapability(self._handle, capability.encode('utf-8'))
        if not ptr:
            return {}
        val = ctypes.string_at(ptr).decode('utf-8')
        self._lib.DistConf_FreeString(ptr)
        return json.loads(val)

    def get_full_config(self) -> Dict[str, Any]:
        ptr = self._lib.DistConf_GetFullConfig(self._handle)
        if not ptr:
            return {}
        val = ctypes.string_at(ptr).decode('utf-8')
        self._lib.DistConf_FreeString(ptr)
        return json.loads(val)

    def decrypt(self, ciphertext: str) -> str:
        ptr = self._lib.DistConf_Decrypt(self._handle, ciphertext.encode('utf-8'))
        if not ptr:
            return ciphertext
        val = ctypes.string_at(ptr).decode('utf-8')
        self._lib.DistConf_FreeString(ptr)
        return val

    def on_live_conf_update(self, callback: Callable[[Dict[str, Any]], None]):
        def _wrapper(handle: int, json_data: bytes):
            data = json.loads(json_data.decode('utf-8'))
            callback(data)
            
        self._callback_ref = CALLBACK_TYPE(_wrapper)
        self._lib.DistConf_OnLiveConfUpdate(self._handle, self._callback_ref)

    def on_registry_update(self, callback: Callable[[Dict[str, list]], None]):
        def _wrapper(handle: int, json_data: bytes):
            data = json.loads(json_data.decode('utf-8'))
            callback(data)
            
        self._registry_callback_ref = CALLBACK_TYPE(_wrapper)
        self._lib.DistConf_OnRegistryUpdate(self._handle, self._registry_callback_ref)

    def close(self):
        if hasattr(self, '_handle') and self._handle:
            self._lib.DistConf_Close(self._handle)
            self._handle = None

    def __del__(self):
        self.close()
