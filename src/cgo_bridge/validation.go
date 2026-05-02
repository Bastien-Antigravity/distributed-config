package cgo_bridge

/*
#include <stdlib.h>
*/
import "C"

// -------------------------------------------------------------------------

// IsValid is a Go-native wrapper for checking handle validity.
func IsValid(handle uintptr) bool {
	FacadeMu.Lock()
	_, ok := FacadeStore[handle]
	FacadeMu.Unlock()
	return ok
}

// -------------------------------------------------------------------------

// ValidateMandatoryServices is a Go-native wrapper.
func ValidateMandatoryServices(handle uintptr) error {
	FacadeMu.Lock()
	session, ok := FacadeStore[handle]
	FacadeMu.Unlock()

	if !ok || session.Config == nil {
		return nil
	}

	return session.Config.ValidateMandatoryServices()
}
