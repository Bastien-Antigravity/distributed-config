import sys
import os
import time
import ctypes
from typing import Optional

# Path to the shared library
LIB_PATH = os.path.join(os.path.dirname(__file__), "..", "libdistconf", "libdistconf.dylib") # Adjust for OS
if sys.platform == "win32":
    LIB_PATH = os.path.join(os.path.dirname(__file__), "..", "libdistconf", "libdistconf.dll")
elif sys.platform == "linux":
    LIB_PATH = os.path.join(os.path.dirname(__file__), "..", "libdistconf", "libdistconf.so")

def validate_ffi():
    print(f"--- FFI Validation: Python ---")
    
    if not os.path.exists(LIB_PATH):
        print(f"Skipping: Library not found at {LIB_PATH}")
        return

    lib = ctypes.CDLL(LIB_PATH)

    # Function signatures
    lib.DistConf_New.restype = ctypes.c_void_p
    lib.DistConf_New.argtypes = [ctypes.c_char_p]
    
    lib.DistConf_Get.restype = ctypes.c_char_p
    lib.DistConf_Get.argtypes = [ctypes.c_void_p, ctypes.c_char_p, ctypes.c_char_p]
    
    lib.DistConf_Set.restype = ctypes.c_int
    lib.DistConf_Set.argtypes = [ctypes.c_void_p, ctypes.c_char_p, ctypes.c_char_p, ctypes.c_char_p]
    
    lib.DistConf_Close.argtypes = [ctypes.c_void_p]

    # 1. Initialize
    handle = lib.DistConf_New(b"standalone")
    if not handle:
        print("FAIL: Failed to create handle")
        sys.exit(1)
    print(f"OK: Created handle {handle}")

    # 2. Set/Get
    lib.DistConf_Set(handle, b"section1", b"key1", b"value1")
    val = lib.DistConf_Get(handle, b"section1", b"key1")
    if val == b"value1":
        print(f"OK: Get returned expected value: {val.decode()}")
    else:
        print(f"FAIL: Get returned {val}")
        sys.exit(1)

    # 3. Close
    lib.DistConf_Close(handle)
    print("OK: Closed handle")
    
    print("--- FFI Validation Success ---")

if __name__ == "__main__":
    validate_ffi()
