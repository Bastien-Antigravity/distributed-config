package cgo_bridge

// =============================================================================
// ESSENTIAL PROCESS:
// CGO configuration accessors (Get, Set, Close) providing thread-safe session
// retrieval and modification across language boundaries.
//
// DATA FLOW:
// 1. Input: Session handle (uintptr), section names, keys, and values.
// 2. Logic: Acquires FacadeMu read/write lock and delegates to underlying facade.Config.
// 3. Output: Retrieved configuration values or updated in-memory state.
//
// KEY PARAMETERS:
// - handle: Unique session identifier referencing an active configuration instance.
// =============================================================================

/*
#include <stdlib.h>
*/
import "C"

// -----------------------------------------------------------------------------

// Get is a Go-native wrapper for DistConf_Get, using string types for cross-package compatibility.
func Get(handle uintptr, section, key string) string {
	FacadeMu.RLock()
	defer FacadeMu.RUnlock()
	session, ok := FacadeStore[handle]

	if !ok || session.Config == nil {
		return ""
	}

	return session.Config.Get(sanitizeString(section), sanitizeString(key))
}

// -------------------------------------------------------------------------

// Set is a Go-native wrapper for DistConf_Set, using string types.
func Set(handle uintptr, section, key, value string) error {
	FacadeMu.RLock()
	defer FacadeMu.RUnlock()
	session, ok := FacadeStore[handle]

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
	FacadeMu.RLock()
	defer FacadeMu.RUnlock()
	session, ok := FacadeStore[handle]

	if !ok || session.Config == nil {
		return nil
	}

	return session.Config.Sync()
}

// ApplyFileOverride is a Go-native wrapper for DistConf_ApplyFileOverride.
func ApplyFileOverride(handle uintptr, filePath string) (string, error) {
	FacadeMu.RLock()
	defer FacadeMu.RUnlock()
	session, ok := FacadeStore[handle]

	if !ok || session.Config == nil {
		return "{}", nil
	}

	return session.Config.ApplyFileOverride(sanitizeString(filePath))
}
