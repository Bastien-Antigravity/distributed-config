#!/usr/bin/env python
# coding:utf-8

from ctypes import CDLL as ctypesCDLL, CFUNCTYPE as ctypesCFUNCTYPE, c_char_p as ctypesCCharP, c_int as ctypesCInt, c_void_p as ctypesCVoidP, string_at as ctypesStringAt
from json import dumps as jsonDumps, loads as jsonLoads
from os import getenv as osGetenv
from os.path import dirname as osPathDirname, exists as osPathExists, join as osPathJoin
from platform import system as platformSystem
from typing import Any as typeAny, Callable as typeCallable, Dict as typeDict, Optional as typeOptional

# Callback type: void (*config_update_cb)(uintptr_t handle, const char* json_data)
CALLBACK_TYPE = ctypesCFUNCTYPE(None, ctypesCVoidP, ctypesCCharP)

# Standardized Error Codes (must match helpers.h)
DISTCONF_SUCCESS                = 0
DISTCONF_ERR_GENERIC            = 1
DISTCONF_ERR_INVALID_HANDLE      = 2
DISTCONF_ERR_KEY_NOT_FOUND       = 3
DISTCONF_ERR_VALIDATION_FAILED   = 4
DISTCONF_ERR_NETWORK_FAILURE     = 5
DISTCONF_ERR_DECRYPTION_FAILED   = 6
DISTCONF_ERR_INVALID_INPUT       = 7

class DistConfError(Exception):
    """Base exception for all DistConf errors."""
    def __init__(self, message: str, code: int) -> None:
        super().__init__(message)
        self.code = code

class DistConfig:
    """
    ESSENTIAL PROCESS:
    Python wrapper for libdistconf. Provides native access to the distributed configuration ecosystem.
    
    DATA FLOW:
    Loads shared library -> Creates session via CGO bridge -> Reads/writes state directly via ctypes.
    """
    
    def __init__(self, profile: str, lib_path: typeOptional[str] = None) -> None:
        self._lib = self._load_lib(lib_path)
        if not self._lib:
            raise RuntimeError("Could not load libdistconf shared library")
            
        self._handle = self._lib.DistConf_New(profile.encode('utf-8'))
        if not self._handle:
            self._raise_last_error()
            
        self._callback_ref = None

    # -----------------------------------------------------------------------------------------------

    def _load_lib(self, lib_path: typeOptional[str]) -> typeOptional[typeAny]:
        if not lib_path:
            system = platformSystem()
            if system == "Darwin":
                ext = ".dylib"
            elif system == "Windows":
                ext = ".dll"
            else:
                ext = ".so"
            
            # 1. Check environment variable
            lib_path = osGetenv("LIBDISTCONF_PATH")
            if not lib_path:
                # 2. Check current package directory (for bundled wheels)
                pkg_dir = osPathDirname(__file__)
                local_lib = osPathJoin(pkg_dir, f"libdistconf{ext}")
                if osPathExists(local_lib):
                    lib_path = local_lib
                else:
                    # 3. Fallback to system search
                    lib_path = f"libdistconf{ext}"
            
        try:
            lib = ctypesCDLL(lib_path)
            
            # Signatures
            lib.DistConf_New.argtypes = [ctypesCCharP]
            lib.DistConf_New.restype = ctypesCVoidP
            
            lib.DistConf_Close.argtypes = [ctypesCVoidP]
            lib.DistConf_Close.restype = None
            
            lib.DistConf_Get.argtypes = [ctypesCVoidP, ctypesCCharP, ctypesCCharP]
            lib.DistConf_Get.restype = ctypesCVoidP
            
            lib.DistConf_Set.argtypes = [ctypesCVoidP, ctypesCCharP, ctypesCCharP, ctypesCCharP]
            lib.DistConf_Set.restype = ctypesCInt
            
            lib.DistConf_Sync.argtypes = [ctypesCVoidP]
            lib.DistConf_Sync.restype = ctypesCInt
            
            lib.DistConf_OnLiveConfUpdate.argtypes = [ctypesCVoidP, CALLBACK_TYPE]
            lib.DistConf_OnLiveConfUpdate.restype = None

            lib.DistConf_OnRegistryUpdate.argtypes = [ctypesCVoidP, CALLBACK_TYPE]
            lib.DistConf_OnRegistryUpdate.restype = None
            
            lib.DistConf_ShareConfig.argtypes = [ctypesCVoidP, ctypesCCharP]
            lib.DistConf_ShareConfig.restype = ctypesCInt
            
            lib.DistConf_ValidateMandatoryServices.argtypes = [ctypesCVoidP]
            lib.DistConf_ValidateMandatoryServices.restype = ctypesCInt
            
            lib.DistConf_GetAddress.argtypes = [ctypesCVoidP, ctypesCCharP]
            lib.DistConf_GetAddress.restype = ctypesCVoidP
            
            lib.DistConf_GetGRPCAddress.argtypes = [ctypesCVoidP, ctypesCCharP]
            lib.DistConf_GetGRPCAddress.restype = ctypesCVoidP

            lib.DistConf_GetRESTAddress.argtypes = [ctypesCVoidP, ctypesCCharP]
            lib.DistConf_GetRESTAddress.restype = ctypesCVoidP
            
            lib.DistConf_GetCapability.argtypes = [ctypesCVoidP, ctypesCCharP]
            lib.DistConf_GetCapability.restype = ctypesCVoidP
            
            lib.DistConf_GetFullConfig.argtypes = [ctypesCVoidP]
            lib.DistConf_GetFullConfig.restype = ctypesCVoidP
            
            lib.DistConf_Decrypt.argtypes = [ctypesCVoidP, ctypesCCharP]
            lib.DistConf_Decrypt.restype = ctypesCVoidP
            
            lib.DistConf_FreeString.argtypes = [ctypesCVoidP]
            lib.DistConf_FreeString.restype = None

            lib.DistConf_GetLastError.argtypes = []
            lib.DistConf_GetLastError.restype = ctypesCCharP

            lib.DistConf_GetLastErrorCode.argtypes = []
            lib.DistConf_GetLastErrorCode.restype = ctypesCInt
            
            return lib
        except Exception as e:
            raise RuntimeError(f"Could not load libdistconf from '{lib_path}': {e}") from e

    def _raise_last_error(self):
        code = self._lib.DistConf_GetLastErrorCode()
        msg_ptr = self._lib.DistConf_GetLastError()
        msg = msg_ptr.decode('utf-8') if msg_ptr else "Unknown error"
        raise DistConfError(msg, code)

    # -----------------------------------------------------------------------------------------------

    def get(self, section: str, key: str) -> str:
        ptr = self._lib.DistConf_Get(self._handle, section.encode('utf-8'), key.encode('utf-8'))
        if not ptr:
            # We don't necessarily want to raise on 'key not found' if it's expected
            # but for standardization, let's see if we should.
            # return "" for now to maintain behavior, but we COULD raise.
            return ""
        val = ctypesStringAt(ptr).decode('utf-8')
        self._lib.DistConf_FreeString(ptr)
        return val

    # -----------------------------------------------------------------------------------------------

    def set(self, section: str, key: str, value: str) -> bool:
        if self._lib.DistConf_Set(self._handle, section.encode('utf-8'), key.encode('utf-8'), value.encode('utf-8')) == 0:
            self._raise_last_error()
        return True

    # -----------------------------------------------------------------------------------------------

    def sync(self) -> bool:
        if self._lib.DistConf_Sync(self._handle) == 0:
            self._raise_last_error()
        return True

    # -----------------------------------------------------------------------------------------------

    def share_config(self, payload: typeAny) -> bool:
        json_data = jsonDumps(payload)
        if self._lib.DistConf_ShareConfig(self._handle, json_data.encode('utf-8')) == 0:
            self._raise_last_error()
        return True

    # -----------------------------------------------------------------------------------------------

    def validate_mandatory_services(self) -> bool:
        if self._lib.DistConf_ValidateMandatoryServices(self._handle) == 0:
            self._raise_last_error()
        return True

    # -----------------------------------------------------------------------------------------------

    def get_address(self, capability: str) -> str:
        ptr = self._lib.DistConf_GetAddress(self._handle, capability.encode('utf-8'))
        if not ptr:
            self._raise_last_error()
        val = ctypesStringAt(ptr).decode('utf-8')
        self._lib.DistConf_FreeString(ptr)
        return val

    # -----------------------------------------------------------------------------------------------

    def get_grpc_address(self, capability: str) -> str:
        ptr = self._lib.DistConf_GetGRPCAddress(self._handle, capability.encode('utf-8'))
        if not ptr:
            self._raise_last_error()
        val = ctypesStringAt(ptr).decode('utf-8')
        self._lib.DistConf_FreeString(ptr)
        return val

    # -----------------------------------------------------------------------------------------------

    def get_rest_address(self, capability: str) -> str:
        ptr = self._lib.DistConf_GetRESTAddress(self._handle, capability.encode('utf-8'))
        if not ptr:
            self._raise_last_error()
        val = ctypesStringAt(ptr).decode('utf-8')
        self._lib.DistConf_FreeString(ptr)
        return val

    # -----------------------------------------------------------------------------------------------

    def get_capability(self, capability: str) -> typeDict[str, typeAny]:
        ptr = self._lib.DistConf_GetCapability(self._handle, capability.encode('utf-8'))
        if not ptr:
            self._raise_last_error()
        val = ctypesStringAt(ptr).decode('utf-8')
        self._lib.DistConf_FreeString(ptr)
        return jsonLoads(val)

    # -----------------------------------------------------------------------------------------------

    def get_full_config(self) -> typeDict[str, typeAny]:
        ptr = self._lib.DistConf_GetFullConfig(self._handle)
        if not ptr:
            self._raise_last_error()
        val = ctypesStringAt(ptr).decode('utf-8')
        self._lib.DistConf_FreeString(ptr)
        return jsonLoads(val)

    # -----------------------------------------------------------------------------------------------

    def decrypt(self, ciphertext: str) -> str:
        ptr = self._lib.DistConf_Decrypt(self._handle, ciphertext.encode('utf-8'))
        if not ptr:
            # Decryption failure is critical
            self._raise_last_error()
        val = ctypesStringAt(ptr).decode('utf-8')
        self._lib.DistConf_FreeString(ptr)
        return val

    # -----------------------------------------------------------------------------------------------

    def on_live_conf_update(self, callback: typeCallable[[typeDict[str, typeAny]], None]) -> None:
        def _wrapper(handle: int, json_data: bytes):
            data = jsonLoads(json_data.decode('utf-8'))
            callback(data)
            
        self._callback_ref = CALLBACK_TYPE(_wrapper)
        self._lib.DistConf_OnLiveConfUpdate(self._handle, self._callback_ref)

    # -----------------------------------------------------------------------------------------------

    def on_registry_update(self, callback: typeCallable[[typeDict[str, list]], None]) -> None:
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
