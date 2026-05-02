package cgo_bridge

/*
#include <stdlib.h>
*/
import "C"

import (
	"encoding/json"
)

// -------------------------------------------------------------------------

//export DistConf_GetAddress
func DistConf_GetAddress(handle uintptr, capability unsafe.Pointer) unsafe.Pointer {
	FacadeMu.Lock()
	session, ok := FacadeStore[handle]
	FacadeMu.Unlock()

	if !ok || session.Config == nil {
		return nil
	}

	addr, err := session.Config.GetAddress(sanitizeString(C.GoString((*C.char)(capability))))
	if err != nil {
		return nil
	}
	return unsafe.Pointer(C.CString(addr))
}

// -------------------------------------------------------------------------

//export DistConf_GetGRPCAddress
func DistConf_GetGRPCAddress(handle uintptr, capability unsafe.Pointer) unsafe.Pointer {
	FacadeMu.Lock()
	session, ok := FacadeStore[handle]
	FacadeMu.Unlock()

	if !ok || session.Config == nil {
		return nil
	}

	addr, err := session.Config.GetGRPCAddress(sanitizeString(C.GoString((*C.char)(capability))))
	if err != nil {
		return nil
	}
	return unsafe.Pointer(C.CString(addr))
}

// -------------------------------------------------------------------------

//export DistConf_GetCapability
func DistConf_GetCapability(handle uintptr, capability unsafe.Pointer) unsafe.Pointer {
	FacadeMu.Lock()
	session, ok := FacadeStore[handle]
	FacadeMu.Unlock()

	if !ok || session.Config == nil {
		return nil
	}

	capKey := sanitizeString(C.GoString((*C.char)(capability)))
	val, exists := session.Config.Capabilities[capKey]
	if !exists || val == nil {
		return nil
	}

	jsonData, err := json.Marshal(val)
	if err != nil {
		return nil
	}
	return unsafe.Pointer(C.CString(string(jsonData)))
}

// -------------------------------------------------------------------------

//export DistConf_GetFullConfig
func DistConf_GetFullConfig(handle uintptr) unsafe.Pointer {
	FacadeMu.Lock()
	session, ok := FacadeStore[handle]
	FacadeMu.Unlock()

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
	return unsafe.Pointer(C.CString(string(jsonData)))
}
