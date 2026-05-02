package main

/*
#include <stdlib.h>
#include "helpers.h"
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

// New is a Go-native wrapper for DistConf_New.
func New(profile string) uintptr {
	prof := sanitizeString(profile)
	
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

<<<<<<< HEAD
//export DistConf_Close
func DistConf_Close(handle uintptr) {
	facadeMu.Lock()
	defer facadeMu.Unlock()
	delete(facadeStore, handle)
=======
// Close is a Go-native wrapper for DistConf_Close.
func Close(handle uintptr) {
	FacadeMu.Lock()
	defer FacadeMu.Unlock()
	delete(FacadeStore, handle)
>>>>>>> develop
}
