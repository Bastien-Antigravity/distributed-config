package main

/*
#include <stdlib.h>
#include <string.h>
*/
import "C"
import (
	"encoding/json"
	"fmt"
	"sync"
	"unsafe"

	"github.com/Bastien-Antigravity/distributed-config"
)

var (
	registry    sync.Map
	nextHandle  int32
	lastError   [1024]byte
	registryMut sync.Mutex
)

// setError sets the global last error string for C access.
func setError(err error) {
	if err != nil {
		s := err.Error()
		if len(s) > 1023 {
			s = s[:1023]
		}
		copy(lastError[:], s)
		lastError[len(s)] = 0
	} else {
		lastError[0] = 0
	}
}

//export GetLastError
func GetLastError() *C.char {
	return (*C.char)(unsafe.Pointer(&lastError[0]))
}

//export CreateConfig
func CreateConfig(profile *C.char) int32 {
	p := C.GoString(profile)
	
	cfg := distributed_config.New(p)
	if cfg == nil {
		setError(fmt.Errorf("failed to create config for profile: %s", p))
		return -1
	}

	registryMut.Lock()
	defer registryMut.Unlock()
	nextHandle++
	registry.Store(nextHandle, cfg)
	return nextHandle
}

//export GetConfigJSON
func GetConfigJSON(handle int32) *C.char {
	val, ok := registry.Load(handle)
	if !ok {
		setError(fmt.Errorf("invalid handle: %d", handle))
		return nil
	}

	cfg, ok := val.(*distributed_config.Config)
	if !ok {
		setError(fmt.Errorf("handle %d is not a valid Config instance", handle))
		return nil
	}

	// Serialize the core config data (which has yaml tags, but json tags work too if not specified)
	data, err := json.Marshal(cfg.Config)
	if err != nil {
		setError(err)
		return nil
	}

	return C.CString(string(data))
}

//export FreeString
func FreeString(s *C.char) {
	C.free(unsafe.Pointer(s))
}

//export CloseConfig
func CloseConfig(handle int32) {
	registry.Delete(handle)
}

func main() {}
