package cgo_bridge

/*
#include <stdlib.h>
*/
import "C"

import (
	"encoding/json"
)

// -------------------------------------------------------------------------

// IsValid is a Go-native wrapper for checking handle validity.
func IsValid(handle uintptr) bool {
	FacadeMu.RLock()
	defer FacadeMu.RUnlock()
	_, ok := FacadeStore[handle]
	return ok
}

// -------------------------------------------------------------------------

// ValidateMandatoryServices is a Go-native wrapper.
func ValidateMandatoryServices(handle uintptr) error {
	FacadeMu.RLock()
	defer FacadeMu.RUnlock()
	session, ok := FacadeStore[handle]

	if !ok || session.Config == nil {
		return nil
	}

	return session.Config.ValidateMandatoryServices()
}

// -------------------------------------------------------------------------

// ShareConfig is a Go-native wrapper for DistConf_ShareConfig.
func ShareConfig(handle uintptr, jsonData string) error {
	FacadeMu.RLock()
	defer FacadeMu.RUnlock()
	session, ok := FacadeStore[handle]

	if !ok || session.Config == nil {
		return nil
	}

	var payload interface{}
	if err := json.Unmarshal([]byte(jsonData), &payload); err != nil {
		return err
	}
	return session.Config.ShareConfig(payload)
}
