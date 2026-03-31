import ctypes
import os
import sys
import json
from typing import Optional, Dict, Any

# Locate the shared library
base_path = os.path.dirname(os.path.abspath(__file__))
lib_path = None

if sys.platform == "darwin":
    lib_path = os.path.join(base_path, "libdistributed_config.dylib")
elif sys.platform == "win32":
    lib_path = os.path.join(base_path, "libdistributed_config.dll")
else:
    lib_path = os.path.join(base_path, "libdistributed_config.so")

# Pre-checks
if not os.path.exists(lib_path):
    # Try parent directory for dev environment
    lib_path = os.path.join(os.path.dirname(base_path), "distributed_config", os.path.basename(lib_path))

class DistributedConfigError(Exception):
    """Exception raised for errors in the DistributedConfig library."""
    pass

class DistributedConfig:
    def __init__(self, profile: str):
        self._lib = self._load_library()
        self.profile = profile
        self.handle = self._lib.CreateConfig(profile.encode('utf-8'))
        if self.handle == -1:
            raise DistributedConfigError(self._get_last_error())
        self._closed = False

    def _load_library(self):
        # Allow overriding lib_path via environment variable
        env_path = os.environ.get("DISTRIBUTED_CONFIG_LIB_PATH")
        local_lib_path = env_path if env_path else lib_path

        if not os.path.exists(local_lib_path):
            # Fallback for development if not installed
            alt_path = os.path.join(os.getcwd(), os.path.basename(lib_path))
            if os.path.exists(alt_path):
                local_lib_path = alt_path
            else:
                raise FileNotFoundError(f"Shared library not found at: {local_lib_path}. Please run 'make build' first.")
        
        lib = ctypes.CDLL(local_lib_path)
        
        # Define argument and return types
        lib.CreateConfig.argtypes = [ctypes.c_char_p]
        lib.CreateConfig.restype = ctypes.c_int32
        
        # Use c_void_p to get the raw pointer so we can free it ourselves
        lib.GetConfigJSON.argtypes = [ctypes.c_int32]
        lib.GetConfigJSON.restype = ctypes.c_void_p
        
        lib.FreeString.argtypes = [ctypes.c_void_p]
        lib.FreeString.restype = None
        
        lib.CloseConfig.argtypes = [ctypes.c_int32]
        lib.CloseConfig.restype = None
        
        lib.GetLastError.argtypes = []
        lib.GetLastError.restype = ctypes.c_char_p
        
        return lib

    def _get_last_error(self) -> str:
        err_ptr = self._lib.GetLastError()
        if err_ptr:
            return err_ptr.decode('utf-8')
        return "Unknown error"

    def get_config(self) -> Dict[str, Any]:
        """Returns the configuration as a Python dictionary."""
        json_ptr = self._lib.GetConfigJSON(self.handle)
        if not json_ptr:
            raise DistributedConfigError(self._get_last_error())
        
        try:
            # Convert pointer to string (ctypes.string_at returns bytes)
            json_str = ctypes.string_at(json_ptr).decode('utf-8')
            config_dict = json.loads(json_str)
            return config_dict
        finally:
            # Free the C string memory on the Go side
            self._lib.FreeString(json_ptr)

    def close(self):
        """Closes the configuration instance and releases resources."""
        if not self._closed:
            self._lib.CloseConfig(self.handle)
            self._closed = True

    def __enter__(self):
        return self

    def __exit__(self, exc_type, exc_val, exc_tb):
        self.close()

    def __del__(self):
        self.close()
