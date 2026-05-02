package main

/*
#include <stdlib.h>
*/
import "C"

import (
	"encoding/json"
)

// -------------------------------------------------------------------------

//export DistConf_GetAddress
func DistConf_GetAddress(handle uintptr, capability *C.char) *C.char {
	facadeMu.Lock()
	session, ok := facadeStore[handle]
	facadeMu.Unlock()

	if !ok || session.Config == nil {
		return nil
	}

	addr, err := session.Config.GetAddress(sanitizeString(C.GoString(capability)))
	if err != nil {
		return nil
	}
	return C.CString(addr)
}

// -------------------------------------------------------------------------

//export DistConf_GetGRPCAddress
func DistConf_GetGRPCAddress(handle uintptr, capability *C.char) *C.char {
	facadeMu.Lock()
	session, ok := facadeStore[handle]
	facadeMu.Unlock()

	if !ok || session.Config == nil {
		return nil
	}

	addr, err := session.Config.GetGRPCAddress(sanitizeString(C.GoString(capability)))
	if err != nil {
		return nil
	}
	return C.CString(addr)
}

// -------------------------------------------------------------------------

//export DistConf_GetCapability
func DistConf_GetCapability(handle uintptr, capability *C.char) *C.char {
	facadeMu.Lock()
	session, ok := facadeStore[handle]
	facadeMu.Unlock()

	if !ok || session.Config == nil {
		return nil
	}

	capKey := sanitizeString(C.GoString(capability))
	val, exists := session.Config.Capabilities[capKey]
	if !exists || val == nil {
		return nil
	}

	jsonData, err := json.Marshal(val)
	if err != nil {
		return nil
	}
	return C.CString(string(jsonData))
}

// -------------------------------------------------------------------------

//export DistConf_GetFullConfig
func DistConf_GetFullConfig(handle uintptr) *C.char {
	facadeMu.Lock()
	session, ok := facadeStore[handle]
	facadeMu.Unlock()

	if !ok || session.Config == nil {
		return nil
	}

	// Create a combined map of all config data
	full := map[string]interface{}{
		"common":       session.Config.Common,
		"capabilities": session.Config.Capabilities,
		"live":         session.Config.LiveConfig.Load(),
	}

	jsonData, err := json.Marshal(full)
	if err != nil {
		return nil
	}
	return C.CString(string(jsonData))
}
