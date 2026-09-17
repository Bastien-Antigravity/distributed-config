package cgo_bridge

// =============================================================================
// ESSENTIAL PROCESS:
// CGO networking functions providing capability lookups, endpoint address
// resolution, and manual server synchronization for polyglot microservices.
//
// DATA FLOW:
// 1. Input: Session handle (uintptr), capability keys, or JSON payload strings.
// 2. Logic: Resolves session and calls underlying facade methods (GetAddress, GetCapability, Sync).
// 3. Output: Resolved address strings, serialized JSON capability dictionaries, or error codes.
//
// KEY PARAMETERS:
// - handle: Safe handle identifier for the active configuration session.
// =============================================================================

/*
#include <stdlib.h>
*/
import "C"

// -----------------------------------------------------------------------------

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
	var mergedMap map[string]interface{}
	if err := session.Config.GetCapability(capKey, &mergedMap); err != nil {
		return "", nil
	}

	jsonData, err := json.Marshal(mergedMap)
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
