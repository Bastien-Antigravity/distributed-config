package cgo_bridge

/*
#include <stdlib.h>
*/
import "C"

import (
	"encoding/json"
)

// -------------------------------------------------------------------------

// GetAddress is a Go-native wrapper for DistConf_GetAddress.
func GetAddress(handle uintptr, capability string) (string, error) {
	FacadeMu.RLock()
	defer FacadeMu.RUnlock()
	session, ok := FacadeStore[handle]

	if !ok || session.Config == nil {
		return "", nil
	}

	return session.Config.GetAddress(sanitizeString(capability))
}

// -------------------------------------------------------------------------

// GetGRPCAddress is a Go-native wrapper for DistConf_GetGRPCAddress.
func GetGRPCAddress(handle uintptr, capability string) (string, error) {
	FacadeMu.RLock()
	defer FacadeMu.RUnlock()
	session, ok := FacadeStore[handle]

	if !ok || session.Config == nil {
		return "", nil
	}

	return session.Config.GetGRPCAddress(sanitizeString(capability))
}

// -------------------------------------------------------------------------

// GetGRPCMgmtAddress is a Go-native wrapper for DistConf_GetGRPCMgmtAddress.
func GetGRPCMgmtAddress(handle uintptr, capability string) (string, error) {
	FacadeMu.RLock()
	defer FacadeMu.RUnlock()
	session, ok := FacadeStore[handle]

	if !ok || session.Config == nil {
		return "", nil
	}

	return session.Config.GetGRPCMgmtAddress(sanitizeString(capability))
}

// -------------------------------------------------------------------------

// GetRESTAddress is a Go-native wrapper for DistConf_GetRESTAddress.
func GetRESTAddress(handle uintptr, capability string) (string, error) {
	FacadeMu.RLock()
	defer FacadeMu.RUnlock()
	session, ok := FacadeStore[handle]

	if !ok || session.Config == nil {
		return "", nil
	}

	return session.Config.GetRESTAddress(sanitizeString(capability))
}

// -------------------------------------------------------------------------

// GetCapability is a Go-native wrapper for DistConf_GetCapability.
func GetCapability(handle uintptr, capability string) (string, error) {
	FacadeMu.RLock()
	defer FacadeMu.RUnlock()
	session, ok := FacadeStore[handle]

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

// GetFullConfig is a Go-native wrapper for DistConf_GetFullConfig.
func GetFullConfig(handle uintptr) (string, error) {
	FacadeMu.RLock()
	defer FacadeMu.RUnlock()
	session, ok := FacadeStore[handle]

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
