package main

/*
#include <stdlib.h>
*/
import "C"

// -------------------------------------------------------------------------

<<<<<<< HEAD
//export DistConf_Get
func DistConf_Get(handle uintptr, section, key *C.char) *C.char {
	facadeMu.Lock()
	session, ok := facadeStore[handle]
	facadeMu.Unlock()
=======
// Get is a Go-native wrapper for DistConf_Get, using string types for cross-package compatibility.
func Get(handle uintptr, section, key string) string {
	FacadeMu.Lock()
	session, ok := FacadeStore[handle]
	FacadeMu.Unlock()
>>>>>>> develop

	if !ok || session.Config == nil {
		return ""
	}

	return session.Config.Get(sanitizeString(section), sanitizeString(key))
}

// -------------------------------------------------------------------------

<<<<<<< HEAD
//export DistConf_Set
func DistConf_Set(handle uintptr, section, key, value *C.char) int {
	facadeMu.Lock()
	session, ok := facadeStore[handle]
	facadeMu.Unlock()
=======
// Set is a Go-native wrapper for DistConf_Set, using string types.
func Set(handle uintptr, section, key, value string) error {
	FacadeMu.Lock()
	session, ok := FacadeStore[handle]
	FacadeMu.Unlock()
>>>>>>> develop

	if !ok || session.Config == nil {
		return nil // Or handle error
	}

	updates := map[string]map[string]string{
		sanitizeString(section): {
			sanitizeString(key): sanitizeString(value),
		},
	}
	
	return session.Config.Set(updates)
}

// -------------------------------------------------------------------------

<<<<<<< HEAD
//export DistConf_Sync
func DistConf_Sync(handle uintptr) int {
	facadeMu.Lock()
	session, ok := facadeStore[handle]
	facadeMu.Unlock()
=======
// Sync is a Go-native wrapper for DistConf_Sync.
func Sync(handle uintptr) error {
	FacadeMu.Lock()
	session, ok := FacadeStore[handle]
	FacadeMu.Unlock()
>>>>>>> develop

	if !ok || session.Config == nil {
		return nil
	}

<<<<<<< HEAD
	if err := session.Config.Sync(); err != nil {
		C.set_last_error(C.CString(err.Error()))
		return 0
	}
	C.set_last_error(nil)
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

		cStr := C.CString(string(jsonData))
		C.call_config_update_cb(cb, C.uintptr_t(handle), cStr)
		C.free(unsafe.Pointer(cStr))
	})
}

// -------------------------------------------------------------------------

//export DistConf_OnRegistryUpdate
func DistConf_OnRegistryUpdate(handle uintptr, cb C.config_update_cb) {
	facadeMu.Lock()
	session, ok := facadeStore[handle]
	facadeMu.Unlock()
	
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
// -------------------------------------------------------------------------

//export DistConf_ShareConfig
func DistConf_ShareConfig(handle uintptr, json_data *C.char) int {
	facadeMu.Lock()
	session, ok := facadeStore[handle]
	facadeMu.Unlock()

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
=======
	return session.Config.Sync()
>>>>>>> develop
}
