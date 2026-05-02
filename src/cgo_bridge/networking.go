package main

/*
#include <stdlib.h>
*/
import "C"

import (
	"encoding/json"
)

// -------------------------------------------------------------------------

<<<<<<< HEAD
//export DistConf_GetAddress
func DistConf_GetAddress(handle uintptr, capability *C.char) *C.char {
	facadeMu.Lock()
	session, ok := facadeStore[handle]
	facadeMu.Unlock()
=======
// GetAddress is a Go-native wrapper for DistConf_GetAddress.
func GetAddress(handle uintptr, capability string) (string, error) {
	FacadeMu.Lock()
	session, ok := FacadeStore[handle]
	FacadeMu.Unlock()
>>>>>>> develop

	if !ok || session.Config == nil {
		return "", nil
	}

	return session.Config.GetAddress(sanitizeString(capability))
}

// -------------------------------------------------------------------------

<<<<<<< HEAD
//export DistConf_GetGRPCAddress
func DistConf_GetGRPCAddress(handle uintptr, capability *C.char) *C.char {
	facadeMu.Lock()
	session, ok := facadeStore[handle]
	facadeMu.Unlock()
=======
// GetGRPCAddress is a Go-native wrapper for DistConf_GetGRPCAddress.
func GetGRPCAddress(handle uintptr, capability string) (string, error) {
	FacadeMu.Lock()
	session, ok := FacadeStore[handle]
	FacadeMu.Unlock()
>>>>>>> develop

	if !ok || session.Config == nil {
		return "", nil
	}

	return session.Config.GetGRPCAddress(sanitizeString(capability))
}

// -------------------------------------------------------------------------

<<<<<<< HEAD
//export DistConf_GetCapability
func DistConf_GetCapability(handle uintptr, capability *C.char) *C.char {
	facadeMu.Lock()
	session, ok := facadeStore[handle]
	facadeMu.Unlock()
=======
// GetCapability is a Go-native wrapper for DistConf_GetCapability.
func GetCapability(handle uintptr, capability string) (string, error) {
	FacadeMu.Lock()
	session, ok := FacadeStore[handle]
	FacadeMu.Unlock()
>>>>>>> develop

	if !ok || session.Config == nil {
		return "", nil
	}

	capKey := sanitizeString(capability)
	val, exists := session.Config.Capabilities[capKey]
	if !exists || val == nil {
		return "", nil
	}

	jsonData, err := json.Marshal(val)
	if err != nil {
		return "", err
	}
	return string(jsonData), nil
}

// -------------------------------------------------------------------------

<<<<<<< HEAD
//export DistConf_GetFullConfig
func DistConf_GetFullConfig(handle uintptr) *C.char {
	facadeMu.Lock()
	session, ok := facadeStore[handle]
	facadeMu.Unlock()
=======
// GetFullConfig is a Go-native wrapper for DistConf_GetFullConfig.
func GetFullConfig(handle uintptr) (string, error) {
	FacadeMu.Lock()
	session, ok := FacadeStore[handle]
	FacadeMu.Unlock()
>>>>>>> develop

	if !ok || session.Config == nil {
		return "", nil
	}

	// Create a combined map of all config data
	full := map[string]interface{}{
		"common":       session.Config.Common,
		"capabilities": session.Config.Capabilities,
		"live":         session.Config.LiveConfig.Load(),
	}

	jsonData, err := json.Marshal(full)
	if err != nil {
		return "", err
	}
	return string(jsonData), nil
}
