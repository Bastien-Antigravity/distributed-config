package main

/*
#include <stdlib.h>
#include <stdint.h>

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
func DistConf_Get(handle uintptr, section, key *C.char) *C.char {
	facadeMu.Lock()
	session, ok := facadeStore[handle]
	facadeMu.Unlock()

	if !ok || session.Config == nil {
		return nil
	}

	val := session.Config.Get(sanitizeString(C.GoString(section)), sanitizeString(C.GoString(key)))
	if val == "" {
		return nil
	}
	return C.CString(val)
}

// -------------------------------------------------------------------------

//export DistConf_Set
func DistConf_Set(handle uintptr, section, key, value *C.char) {
	facadeMu.Lock()
	session, ok := facadeStore[handle]
	facadeMu.Unlock()

	if ok && session.Config != nil {
		session.Config.Set(sanitizeString(C.GoString(section)), sanitizeString(C.GoString(key)), sanitizeString(C.GoString(value)))
	}
}

// -------------------------------------------------------------------------

//export DistConf_Sync
func DistConf_Sync(handle uintptr) int {
	facadeMu.Lock()
	session, ok := facadeStore[handle]
	facadeMu.Unlock()

	if !ok || session.Config == nil {
		return 0
	}

	if err := session.Config.Sync(); err != nil {
		return 0
	}
	return 1
}

// -------------------------------------------------------------------------

//export DistConf_OnLiveConfUpdate
func DistConf_OnLiveConfUpdate(handle uintptr, cb C.config_update_cb) {
	facadeMu.Lock()
	session, ok := facadeStore[handle]
	facadeMu.Unlock()
	
	if !ok || session.Config == nil {
		return
	}

	session.Config.OnLiveConfUpdate(func(update map[string]map[string]string) {
		jsonData, err := json.Marshal(update)
		if err != nil {
			return
		}

		// Run callback in a goroutine to avoid deadlocks with language runtimes (e.g. Python GIL)
		go func() {
			cStr := C.CString(string(jsonData))
			C.call_config_update_cb(cb, C.uintptr_t(handle), cStr)
			C.free(unsafe.Pointer(cStr))
		}()
	})
}
// -------------------------------------------------------------------------

//export DistConf_ShareObject
func DistConf_ShareObject(handle uintptr, section, json_data *C.char) int {
	facadeMu.Lock()
	session, ok := facadeStore[handle]
	facadeMu.Unlock()

	if !ok || session.Config == nil {
		return 0
	}

	sec := sanitizeString(C.GoString(section))
	rawJson := C.GoString(json_data)

	// Unmarshal JSON into a map to use with ShareObject
	var payload interface{}
	if err := json.Unmarshal([]byte(rawJson), &payload); err != nil {
		return 0
	}

	if err := session.Config.ShareObject(sec, payload); err != nil {
		return 0
	}
	return 1
}
