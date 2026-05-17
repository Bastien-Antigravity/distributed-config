package cgo_bridge

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestCGOBridge_StressRace(t *testing.T) {
	// Stress test for concurrent New, Get, Set, and Close.
	// This should be run with -race flag.

	const numGoroutines = 50
	const iterations = 100
	var wg sync.WaitGroup

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				// 1. Create a new handle
				profile := "standalone"
				if j%2 == 0 {
					profile = "test"
				}
				handle := New(profile)
				if handle == 0 {
					continue
				}

				// 2. Perform concurrent operations on this handle
				var innerWg sync.WaitGroup
				innerWg.Add(3)

				// Getter
				go func() {
					defer innerWg.Done()
					for k := 0; k < 10; k++ {
						_ = Get(handle, "section", "key")
						time.Sleep(1 * time.Microsecond)
					}
				}()

				// Setter
				go func() {
					defer innerWg.Done()
					for k := 0; k < 10; k++ {
						_ = Set(handle, "section", "key", fmt.Sprintf("val-%d", k))
						time.Sleep(1 * time.Microsecond)
					}
				}()

				// Closer (triggered after a short delay)
				go func() {
					defer innerWg.Done()
					time.Sleep(5 * time.Microsecond)
					Close(handle)
				}()

				innerWg.Wait()
			}
		}(i)
	}

	wg.Wait()
}
