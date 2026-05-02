package cgo_bridge

/*
#include "helpers.h"
*/
import "C"

import (
	"unsafe"
	"github.com/Bastien-Antigravity/distributed-config"
)

//export DistConf_GetLastError
func DistConf_GetLastError() unsafe.Pointer {
	return unsafe.Pointer(C.last_error)
}

// -------------------------------------------------------------------------

//export DistConf_Decrypt
func DistConf_Decrypt(handle uintptr, ciphertext unsafe.Pointer) unsafe.Pointer {
	// Note: Handle is not strictly needed here as Decrypt is a static package method,
	// but we keep it for consistency and potential future handle-based decryption (e.g. session-specific keys).
	
	decrypted, err := distributed_config.Decrypt(sanitizeString(C.GoString((*C.char)(ciphertext))))
	if err != nil {
		C.set_last_error(C.CString(err.Error()))
		return nil
	}
	C.set_last_error(nil)
	return unsafe.Pointer(C.CString(decrypted))
}
