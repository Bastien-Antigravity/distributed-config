package main

/*
#include <stdlib.h>
*/
import "C"

// -------------------------------------------------------------------------

//export DistConf_ValidateMandatoryServices
func DistConf_ValidateMandatoryServices(handle uintptr) int {
	facadeMu.Lock()
	session, ok := facadeStore[handle]
	facadeMu.Unlock()

	if !ok || session.Config == nil {
		return 0
	}

	if err := session.Config.ValidateMandatoryServices(); err != nil {
		return 0
	}
	return 1
}
