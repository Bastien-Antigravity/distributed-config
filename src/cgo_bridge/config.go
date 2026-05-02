package cgo_bridge

/*
#include <stdlib.h>
*/
import "C"

// -------------------------------------------------------------------------

// Get is a Go-native wrapper for DistConf_Get, using string types for cross-package compatibility.
func Get(handle uintptr, section, key string) string {
	FacadeMu.Lock()
	session, ok := FacadeStore[handle]
	FacadeMu.Unlock()

	if !ok || session.Config == nil {
		return ""
	}

	return session.Config.Get(sanitizeString(section), sanitizeString(key))
}

// -------------------------------------------------------------------------

// Set is a Go-native wrapper for DistConf_Set, using string types.
func Set(handle uintptr, section, key, value string) error {
	FacadeMu.Lock()
	session, ok := FacadeStore[handle]
	FacadeMu.Unlock()

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

// Sync is a Go-native wrapper for DistConf_Sync.
func Sync(handle uintptr) error {
	FacadeMu.Lock()
	session, ok := FacadeStore[handle]
	FacadeMu.Unlock()

	if !ok || session.Config == nil {
		return nil
	}

	return session.Config.Sync()
}
