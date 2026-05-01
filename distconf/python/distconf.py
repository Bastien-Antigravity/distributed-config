#!/usr/bin/env python
# coding:utf-8

from os import getenv as osGetenv
from ctypes import CDLL as ctypesCDLL, string_at as ctypesStringAt, c_char_p, c_void_p, CFUNCTYPE, c_int
from json import dumps as jsonDumps, loads as jsonLoads
from typing import Any, Callable, Dict, Optional

# Callback type: void (*config_update_cb)(uintptr_t handle, const char* json_data)
CALLBACK_TYPE = CFUNCTYPE(None, c_void_p, c_char_p)

class DistConfig:
    """
    ESSENTIAL PROCESS:
    Python wrapper for libdistconf. Provides native access to the distributed configuration ecosystem.
    
    DATA FLOW:
    Loads shared library -> Creates session via CGO bridge -> Reads/writes state directly via ctypes.
    
    KEY PARAMETERS:
    - profile: The environment profile to load (e.g., 'standalone', 'production').
    - lib_path: Optional explicit path to the libdistconf shared library.
    """
    
    def __init__(self, profile: str, lib_path: Optional[str] = None) -> None:
        self._lib = self._load_lib(lib_path)
        if not self._lib:
            raise RuntimeError("Could not load libdistconf shared library")
            
        self._handle = self._lib.DistConf_New(profile.encode('utf-8'))
        if not self._handle:
            raise RuntimeError(f"Failed to initialize DistConf with profile: {profile}")
            
        self._callback_ref = None

    # -----------------------------------------------------------------------------------------------

    def _load_lib(self, lib_path: Optional[str]) -> Optional[Any]:
        if not lib_path:
            lib_path = osGetenv("LIBDISTCONF_PATH", "libdistconf.so")
            
        try:
            lib = ctypesCDLL(lib_path)
            
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

    # -----------------------------------------------------------------------------------------------

    def get(self, section: str, key: str) -> str:
        ptr = self._lib.DistConf_Get(self._handle, section.encode('utf-8'), key.encode('utf-8'))
        if not ptr:
            return ""
        val = ctypesStringAt(ptr).decode('utf-8')
        self._lib.DistConf_FreeString(ptr)
        return val

    # -----------------------------------------------------------------------------------------------

    def set(self, section: str, key: str, value: str) -> None:
        self._lib.DistConf_Set(self._handle, section.encode('utf-8'), key.encode('utf-8'), value.encode('utf-8'))

    # -----------------------------------------------------------------------------------------------

    def sync(self) -> bool:
        return self._lib.DistConf_Sync(self._handle) == 1

    # -----------------------------------------------------------------------------------------------

    def share_config(self, payload: Any) -> bool:
        json_data = jsonDumps(payload)
        return self._lib.DistConf_ShareConfig(self._handle, json_data.encode('utf-8')) == 1

    # -----------------------------------------------------------------------------------------------

    def validate_mandatory_services(self) -> bool:
        return self._lib.DistConf_ValidateMandatoryServices(self._handle) == 1

    # -----------------------------------------------------------------------------------------------

    def get_address(self, capability: str) -> str:
        ptr = self._lib.DistConf_GetAddress(self._handle, capability.encode('utf-8'))
        if not ptr:
            return ""
        val = ctypesStringAt(ptr).decode('utf-8')
        self._lib.DistConf_FreeString(ptr)
        return val

    # -----------------------------------------------------------------------------------------------

    def get_grpc_address(self, capability: str) -> str:
        ptr = self._lib.DistConf_GetGRPCAddress(self._handle, capability.encode('utf-8'))
        if not ptr:
            return ""
        val = ctypesStringAt(ptr).decode('utf-8')
        self._lib.DistConf_FreeString(ptr)
        return val

    # -----------------------------------------------------------------------------------------------

    def get_capability(self, capability: str) -> Dict[str, Any]:
        ptr = self._lib.DistConf_GetCapability(self._handle, capability.encode('utf-8'))
        if not ptr:
            return {}
        val = ctypesStringAt(ptr).decode('utf-8')
        self._lib.DistConf_FreeString(ptr)
        return jsonLoads(val)

    # -----------------------------------------------------------------------------------------------

    def get_full_config(self) -> Dict[str, Any]:
        ptr = self._lib.DistConf_GetFullConfig(self._handle)
        if not ptr:
            return {}
        val = ctypesStringAt(ptr).decode('utf-8')
        self._lib.DistConf_FreeString(ptr)
        return jsonLoads(val)

    # -----------------------------------------------------------------------------------------------

    def decrypt(self, ciphertext: str) -> str:
        ptr = self._lib.DistConf_Decrypt(self._handle, ciphertext.encode('utf-8'))
        if not ptr:
            return ciphertext
        val = ctypesStringAt(ptr).decode('utf-8')
        self._lib.DistConf_FreeString(ptr)
        return val

    # -----------------------------------------------------------------------------------------------

    def on_live_conf_update(self, callback: Callable[[Dict[str, Any]], None]) -> None:
        def _wrapper(handle: int, json_data: bytes):
            data = jsonLoads(json_data.decode('utf-8'))
            callback(data)
            
        self._callback_ref = CALLBACK_TYPE(_wrapper)
        self._lib.DistConf_OnLiveConfUpdate(self._handle, self._callback_ref)

    # -----------------------------------------------------------------------------------------------

    def on_registry_update(self, callback: Callable[[Dict[str, list]], None]) -> None:
        def _wrapper(handle: int, json_data: bytes):
            data = jsonLoads(json_data.decode('utf-8'))
            callback(data)
            
        self._registry_callback_ref = CALLBACK_TYPE(_wrapper)
        self._lib.DistConf_OnRegistryUpdate(self._handle, self._registry_callback_ref)

    # -----------------------------------------------------------------------------------------------

    def close(self) -> None:
        if hasattr(self, '_handle') and self._handle:
            self._lib.DistConf_Close(self._handle)
            self._handle = None

    # -----------------------------------------------------------------------------------------------

    def __del__(self) -> None:
        self.close()
