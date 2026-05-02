package cgo_bridge

/*
#include <stdlib.h>
*/
import "C"

import (
	"sync"
	"unsafe"

	"github.com/Bastien-Antigravity/distributed-config"
)

// ConfigSession holds the state for a single library instantiation.
type ConfigSession struct {
	Config *distributed_config.Config
}

var (
	FacadeMu    sync.Mutex
	FacadeStore         = make(map[uintptr]*ConfigSession)
	FacadeId    uintptr = 1
)

// -------------------------------------------------------------------------

//export DistConf_New
func DistConf_New(profile unsafe.Pointer) uintptr {
	prof := sanitizeString(C.GoString((*C.char)(profile)))
	
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

//export DistConf_Close
func DistConf_Close(handle uintptr) {
	FacadeMu.Lock()
	defer FacadeMu.Unlock()
	delete(FacadeStore, handle)
}
