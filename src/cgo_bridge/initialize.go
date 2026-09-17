package cgo_bridge

// =============================================================================
// ESSENTIAL PROCESS:
// CGO session manager tracking active library instances in thread-safe memory,
// assigning integer uintptr handles to callers across FFI boundaries.
//
// DATA FLOW:
// 1. Input: Profile string from C caller.
// 2. Logic: Creates distributed_config.New(profile), allocates handle ID, stores session.
// 3. Output: Opaque uintptr handle identifying the session for future C calls.
//
// KEY PARAMETERS:
// - FacadeStore: In-memory map from uintptr handles to *ConfigSession structs.
// =============================================================================

/*
#include <stdlib.h>
#include "helpers.h"
*/
import "C"

// -----------------------------------------------------------------------------

import (
	"sync"

	"github.com/Bastien-Antigravity/distributed-config"
)

// ConfigSession holds the state for a single library instantiation.
type ConfigSession struct {
	Config *distributed_config.Config
}

var (
	FacadeMu    sync.RWMutex
	FacadeStore         = make(map[uintptr]*ConfigSession)
	FacadeId    uintptr = 1
)

// -------------------------------------------------------------------------

// New is a Go-native wrapper for DistConf_New.
func New(profile string) uintptr {
	prof := sanitizeString(profile)

	switch prof {
	case "standalone", "test", "staging", "production":
		// Allowed profiles
	default:
		return 0
	}

	cfg := distributed_config.New(prof)
	if cfg == nil {
		return 0
	}

	FacadeMu.Lock()
	defer FacadeMu.Unlock()

	id := FacadeId
	FacadeStore[id] = &ConfigSession{
		Config: cfg,
	}
	FacadeId++
	return id
}

// -------------------------------------------------------------------------

// Close is a Go-native wrapper for DistConf_Close.
func Close(handle uintptr) {
	FacadeMu.Lock()
	session, ok := FacadeStore[handle]
	if ok {
		delete(FacadeStore, handle)
	}
	FacadeMu.Unlock()

	if ok && session.Config != nil {
		_ = session.Config.Close()
	}
}
