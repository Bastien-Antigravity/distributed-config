package cgo_bridge

/*
#include <stdlib.h>
#include <stdint.h>
#include "helpers.h"

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
)

// -------------------------------------------------------------------------

//export DistConf_Get
func DistConf_Get(handle uintptr, section, key unsafe.Pointer) unsafe.Pointer {
	FacadeMu.Lock()
	session, ok := FacadeStore[handle]
	FacadeMu.Unlock()

	if !ok || session.Config == nil {
		return nil
	}

	sec := C.GoString((*C.char)(section))
	k := C.GoString((*C.char)(key))

	val := session.Config.Get(sanitizeString(sec), sanitizeString(k))
	if val == "" {
		return nil
	}
	return unsafe.Pointer(C.CString(val))
}

// -------------------------------------------------------------------------

//export DistConf_Set
func DistConf_Set(handle uintptr, section, key, value unsafe.Pointer) {
	FacadeMu.Lock()
	session, ok := FacadeStore[handle]
	FacadeMu.Unlock()

	if !ok || session.Config == nil {
		return
	}

	sec := C.GoString((*C.char)(section))
	k := C.GoString((*C.char)(key))
	v := C.GoString((*C.char)(value))

	updates := map[string]map[string]string{
		sanitizeString(sec): {
			sanitizeString(k): sanitizeString(v),
		},
	}
	
	session.Config.Set(updates)
}

// -------------------------------------------------------------------------

//export DistConf_Sync
func DistConf_Sync(handle uintptr) int {
	FacadeMu.Lock()
	session, ok := FacadeStore[handle]
	FacadeMu.Unlock()

	if !ok || session.Config == nil {
		return 0
	}

	if err := session.Config.Sync(); err != nil {
		C.set_last_error(C.CString(err.Error()))
		return 0
	}
	C.set_last_error(nil)
	return 1
}

// -------------------------------------------------------------------------

//export DistConf_OnLiveConfUpdate
func DistConf_OnLiveConfUpdate(handle uintptr, cb unsafe.Pointer) {
	FacadeMu.Lock()
	session, ok := FacadeStore[handle]
	FacadeMu.Unlock()
	
	if !ok || session.Config == nil {
		return
	}

	actualCb := (C.config_update_cb)(cb)

	session.Config.OnLiveConfUpdate(func(update map[string]map[string]string) {
		jsonData, err := json.Marshal(update)
		if err != nil {
			return
		}

		cStr := C.CString(string(jsonData))
		C.call_config_update_cb(actualCb, C.uintptr_t(handle), cStr)
		C.free(unsafe.Pointer(cStr))
	})
}

// -------------------------------------------------------------------------

//export DistConf_OnRegistryUpdate
func DistConf_OnRegistryUpdate(handle uintptr, cb unsafe.Pointer) {
	FacadeMu.Lock()
	session, ok := FacadeStore[handle]
	FacadeMu.Unlock()
	
	if !ok || session.Config == nil {
		return
	}

	actualCb := (C.config_update_cb)(cb)

	session.Config.OnRegistryUpdate(func(registry map[string][]string) {
		jsonData, err := json.Marshal(registry)
		if err != nil {
			return
		}

		cStr := C.CString(string(jsonData))
		C.call_config_update_cb(actualCb, C.uintptr_t(handle), cStr)
		C.free(unsafe.Pointer(cStr))
	})
}
// -------------------------------------------------------------------------

//export DistConf_ShareConfig
func DistConf_ShareConfig(handle uintptr, json_data *C.char) int {
	FacadeMu.Lock()
	session, ok := FacadeStore[handle]
	FacadeMu.Unlock()

	if !ok || session.Config == nil {
		return 0
	}

	rawJson := C.GoString(json_data)

	// Unmarshal JSON into a map to use with ShareConfig
	var payload interface{}
	if err := json.Unmarshal([]byte(rawJson), &payload); err != nil {
		return 0
	}

	if err := session.Config.ShareConfig(payload); err != nil {
		return 0
	}
	return 1
}
