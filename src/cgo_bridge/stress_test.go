package cgo_bridge

// =============================================================================
// ESSENTIAL PROCESS:
// High-concurrency stress test suite verifying thread-safety and race-free
// operation of concurrent session creation, reads, writes, and deletions.
//
// DATA FLOW:
// 1. Input: 50 concurrent goroutines performing 100 iterations of session lifecycle actions.
// 2. Logic: Exercises FacadeMu under high contention with concurrent Get/Set/Close.
// 3. Output: Validates zero data races and absence of use-after-free panic conditions.
//
// KEY PARAMETERS:
// - numGoroutines: Concurrency level (50).
// - iterations: Repetitions per goroutine (100).
// =============================================================================

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

// -----------------------------------------------------------------------------

func TestCGOBridge_StressRace(t *testing.T) {
	// Stress test for concurrent New, Get, Set, and Close.
	// This should be run with -race flag.
	tempDir := t.TempDir()
	fixturePath := filepath.Join(tempDir, "stress_standalone.yaml")
	fixture := `
common:
  name: "stress-test"
capabilities:
  log_server:
    ip: "127.0.0.1"
    port: "9020"
  config_server:
    ip: "127.0.0.1"
    port: "3306"
`
	if err := os.WriteFile(fixturePath, []byte(fixture), 0644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("CONFIG_PATH", fixturePath)

	const numGoroutines = 50
	const iterations = 100
	var wg sync.WaitGroup

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				// 1. Create a new handle using standalone to test handle lifecycle without network noise
				handle := New("standalone")
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
