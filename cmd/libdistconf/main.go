package main

/*
#include <stdlib.h>
#include <stdint.h>
#include "../../src/cgo_bridge/helpers.h"

// Define the callback type for C
typedef void (*config_update_cb)(uintptr_t handle, const char* json_data);

// Helper to safely execute a C callback from Go
static void call_config_update_cb(config_update_cb cb, uintptr_t handle, const char* json_data) {
    if (cb != NULL) {
        cb(handle, json_data);
    }
}
*/
import "C"

import (
	"encoding/json"
	"unsafe"
	"github.com/Bastien-Antigravity/distributed-config/src/cgo_bridge"
)

func main() {}

// -------------------------------------------------------------------------

//export DistConf_FreeString
func DistConf_FreeString(ptr *C.char) {
	if ptr != nil {
		C.free(unsafe.Pointer(ptr))
	}
}

//export DistConf_New
func DistConf_New(profile *C.char) uintptr {
	return cgo_bridge.New(C.GoString(profile))
}

//export DistConf_Close
func DistConf_Close(handle uintptr) {
	cgo_bridge.Close(handle)
}

//export DistConf_Get
func DistConf_Get(handle uintptr, section, key *C.char) *C.char {
	val := cgo_bridge.Get(handle, C.GoString(section), C.GoString(key))
	if val == "" {
		C.set_last_error(C.CString("key not found"))
		return nil
	}
	C.set_last_error(nil)
	return C.CString(val)
}

//export DistConf_Set
func DistConf_Set(handle uintptr, section, key, value *C.char) int {
	if err := cgo_bridge.Set(handle, C.GoString(section), C.GoString(key), C.GoString(value)); err != nil {
		C.set_last_error(C.CString(err.Error()))
		return 0
	}
	C.set_last_error(nil)
	return 1
}

//export DistConf_Sync
func DistConf_Sync(handle uintptr) int {
	if err := cgo_bridge.Sync(handle); err != nil {
		C.set_last_error(C.CString(err.Error()))
		return 0
	}
	C.set_last_error(nil)
	return 1
}

//export DistConf_GetAddress
func DistConf_GetAddress(handle uintptr, capability *C.char) *C.char {
	addr, err := cgo_bridge.GetAddress(handle, C.GoString(capability))
	if err != nil {
		return nil
	}
	return C.CString(addr)
}

//export DistConf_GetGRPCAddress
func DistConf_GetGRPCAddress(handle uintptr, capability *C.char) *C.char {
	addr, err := cgo_bridge.GetGRPCAddress(handle, C.GoString(capability))
	if err != nil {
		return nil
	}
	return C.CString(addr)
}

//export DistConf_GetCapability
func DistConf_GetCapability(handle uintptr, capability *C.char) *C.char {
	val, err := cgo_bridge.GetCapability(handle, C.GoString(capability))
	if err != nil {
		return nil
	}
	return C.CString(val)
}

//export DistConf_GetFullConfig
func DistConf_GetFullConfig(handle uintptr) *C.char {
	val, err := cgo_bridge.GetFullConfig(handle)
	if err != nil {
		return nil
	}
	return C.CString(val)
}

//export DistConf_GetLastError
func DistConf_GetLastError() *C.char {
	return C.last_error
}

//export DistConf_Decrypt
func DistConf_Decrypt(handle uintptr, ciphertext *C.char) *C.char {
	decrypted, err := cgo_bridge.Decrypt(C.GoString(ciphertext))
	if err != nil {
		C.set_last_error(C.CString(err.Error()))
		return nil
	}
	C.set_last_error(nil)
	return C.CString(decrypted)
}

//export DistConf_IsValid
func DistConf_IsValid(handle uintptr) int {
	if ok := cgo_bridge.IsValid(handle); ok {
		return 1
	}
	return 0
}

//export DistConf_ValidateMandatoryServices
func DistConf_ValidateMandatoryServices(handle uintptr) int {
	if err := cgo_bridge.ValidateMandatoryServices(handle); err != nil {
		return 0
	}
	return 1
}

//export DistConf_ShareConfig
func DistConf_ShareConfig(handle uintptr, jsonData *C.char) int {
	if err := cgo_bridge.ShareConfig(handle, C.GoString(jsonData)); err != nil {
		return 0
	}
	return 1
}

// -------------------------------------------------------------------------
// CALLBACKS
// -------------------------------------------------------------------------

//export DistConf_OnLiveConfUpdate
func DistConf_OnLiveConfUpdate(handle uintptr, cb C.config_update_cb) {
	cgo_bridge.FacadeMu.Lock()
	session, ok := cgo_bridge.FacadeStore[handle]
	cgo_bridge.FacadeMu.Unlock()

	if !ok || session.Config == nil {
		return
	}

	session.Config.OnLiveConfUpdate(func(update map[string]map[string]string) {
		jsonData, err := json.Marshal(update)
		if err != nil {
			return
		}

		cStr := C.CString(string(jsonData))
		C.call_config_update_cb(cb, C.uintptr_t(handle), cStr)
		C.free(unsafe.Pointer(cStr))
	})
}

//export DistConf_OnRegistryUpdate
func DistConf_OnRegistryUpdate(handle uintptr, cb C.config_update_cb) {
	cgo_bridge.FacadeMu.Lock()
	session, ok := cgo_bridge.FacadeStore[handle]
	cgo_bridge.FacadeMu.Unlock()

	if !ok || session.Config == nil {
		return
	}

	session.Config.OnRegistryUpdate(func(registry map[string][]string) {
		jsonData, err := json.Marshal(registry)
		if err != nil {
			return
		}

		cStr := C.CString(string(jsonData))
		C.call_config_update_cb(cb, C.uintptr_t(handle), cStr)
		C.free(unsafe.Pointer(cStr))
	})
}
