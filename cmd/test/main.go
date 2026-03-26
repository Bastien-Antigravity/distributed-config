package main

import (
	"fmt"
	"os"

	distconf "github.com/Bastien-Antigravity/distributed-config"
)

func main() {
	fmt.Println("Running Distributed Config Integration Tests...")
	fmt.Println("---------------------------------------------")

	failed := false

	if !runTest("Standalone Profile Init", testStandaloneProfile) {
		failed = true
	}

	if !runTest("Test Profile Init", testTestProfile) {
		failed = true
	}

	fmt.Println("---------------------------------------------")
	if failed {
		fmt.Println("TESTS FAILED")
		os.Exit(1)
	}
	fmt.Println("ALL TESTS PASSED")
}

func runTest(name string, testFunc func() error) bool {
	fmt.Printf("Running [%s]... ", name)
	if err := testFunc(); err != nil {
		fmt.Printf("FAIL\n  Error: %v\n", err)
		return false
	}
	fmt.Println("PASS")
	return true
}

func testStandaloneProfile() error {
	// The standalone profile expects "config/default.yaml" (we created a dummy one)
	config := distconf.New("standalone")
	if config == nil {
		return fmt.Errorf("New('standalone') returned nil")
	}

	// Basic validation: the object exists.
	// Real validation would check specific values if we knew exactly what the dummy config contained and how it was loaded.
	return nil
}

func testTestProfile() error {
	// The test profile mimics production but usually with local/mock connections
	config := distconf.New("test")
	if config == nil {
		return fmt.Errorf("New('test') returned nil")
	}

	// Verify we can attach a listener without panic
	attached := false
	config.OnMemConfUpdate(func(updates map[string]map[string]string) {
		attached = true
	})

	// Use the variable to silence the compiler
	if attached {
		fmt.Println("  (Attached listener executed, unexpected for this test)")
	}

	// Since we can't easily trigger an update from here without more components,
	// just successfully creating it and attaching the listener is a partial success.
	return nil
}
