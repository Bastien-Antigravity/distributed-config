package cgo_bridge

/*
#include <stdlib.h>
*/
import "C"

// -------------------------------------------------------------------------

//export DistConf_IsValid
func DistConf_IsValid(handle uintptr) int {
	FacadeMu.Lock()
	_, ok := FacadeStore[handle]
	FacadeMu.Unlock()

	if !ok {
		return 0
	}
	return 1
}

//export DistConf_ValidateMandatoryServices
func DistConf_ValidateMandatoryServices(handle uintptr) int {
	FacadeMu.Lock()
	session, ok := FacadeStore[handle]
	FacadeMu.Unlock()

	if !ok || session.Config == nil {
		return 0
	}

	if err := session.Config.ValidateMandatoryServices(); err != nil {
		return 0
	}
	return 1
}
