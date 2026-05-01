package main

/*
#include <stdlib.h>
*/
import "C"

import (
	"sync"

	"github.com/Bastien-Antigravity/distributed-config"
)

// ConfigSession holds the state for a single library instantiation.
type ConfigSession struct {
	Config *distributed_config.Config
}

var (
	facadeMu    sync.Mutex
	facadeStore         = make(map[uintptr]*ConfigSession)
	facadeId    uintptr = 1
)

func main() {}

// -------------------------------------------------------------------------

//export DistConf_New
func DistConf_New(profile *C.char) uintptr {
	prof := sanitizeString(C.GoString(profile))
	
	cfg := distributed_config.New(prof)
	if cfg == nil {
		return 0
	}

	facadeMu.Lock()
	defer facadeMu.Unlock()

	id := facadeId
	facadeStore[id] = &ConfigSession{
		Config: cfg,
	}
	facadeId++
	return id
}

// -------------------------------------------------------------------------

//export DistConf_Close
func DistConf_Close(handle uintptr) {
	facadeMu.Lock()
	defer facadeMu.Unlock()
	if _, ok := facadeStore[handle]; ok {
		delete(facadeStore, handle)
	}
}
