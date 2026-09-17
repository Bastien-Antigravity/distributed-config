package cgo_bridge

// =============================================================================
// ESSENTIAL PROCESS:
// Integration test suite verifying CGO bridge session lifecycle, handle
// allocation, value get/set, and callback triggers from Go tests.
//
// DATA FLOW:
// 1. Input: Simulated C-like calls to initialize, get, set, and close sessions.
// 2. Logic: Exercises bridge function wrappers across sessions.
// 3. Output: Test assertion results verifying handle safety and value retention.
//
// KEY PARAMETERS:
// - t: Standard testing harness handle.
// =============================================================================

import (
	"encoding/json"
	"os"
	"testing"
	"time"
)

// -----------------------------------------------------------------------------

func TestBridge_ExpandedName(t *testing.T) {
	// Set the environment variable for testing expansion
	_ = os.Setenv("NAME", "dynamic-bridge-app")
	defer func() {
		_ = os.Unsetenv("NAME")
	}()

	t.Run("Initialize and check expanded name", func(t *testing.T) {
		handle := testInit("standalone")
		defer Close(handle)

		FacadeMu.Lock()
		session := FacadeStore[handle]
		FacadeMu.Unlock()

		if session.Config.Common.Name != "dynamic-bridge-app" {
			t.Errorf("Expected expanded name 'dynamic-bridge-app', got '%s'", session.Config.Common.Name)
		}
	})
}

// -------------------------------------------------------------------------

func TestBridge_GetSet(t *testing.T) {
	t.Run("Set and Get configuration", func(t *testing.T) {
		handle := testInit("standalone")
		defer Close(handle)

		section := "test-section"
		key := "test-key"
		val := "test-value"

		// Set value
		if err := Set(handle, section, key, val); err != nil {
			t.Errorf("Set failed: %v", err)
		}

		// Get value
		res := Get(handle, section, key)
		if res != val {
			t.Errorf("Expected '%s', got '%s'", val, res)
		}
	})
}

// -------------------------------------------------------------------------

func TestBridge_LiveUpdate(t *testing.T) {
	t.Run("Receive live update callback", func(t *testing.T) {
		handle := testInit("standalone")
		defer Close(handle)

		FacadeMu.Lock()
		session := FacadeStore[handle]
		FacadeMu.Unlock()

		updated := make(chan bool, 1)
		session.Config.OnLiveConfUpdate(func(updates map[string]map[string]string) {
			if updates["live"]["key"] == "new-value" {
				updated <- true
			}
		})

		// Simulate live update by calling Set
		updates := map[string]map[string]string{
			"live": {"key": "new-value"},
		}
		if err := session.Config.Set(updates); err != nil {
			t.Fatalf("Failed to set configuration: %v", err)
		}

		select {
		case <-updated:
			// Success
		case <-time.After(1 * time.Second):
			t.Error("Timed out waiting for live update callback")
		}
	})
}

// -------------------------------------------------------------------------

func TestBridge_Sync(t *testing.T) {
	t.Run("Perform configuration sync", func(t *testing.T) {
		handle := testInit("standalone")
		defer Close(handle)

		// Test Sync
		if err := Sync(handle); err != nil {
			t.Errorf("Sync failed: %v", err)
		}
	})
}

// -----------------------------------------------------------------------------

func TestBridge_LiveCapabilityUpdate(t *testing.T) {
	t.Run("GetCapability reflects live config overrides", func(t *testing.T) {
		handle := testInit("standalone")
		defer Close(handle)

		// Check initial capability
		initialCap, err := GetCapability(handle, "config_server")
		if err != nil {
			t.Fatalf("GetCapability failed: %v", err)
		}
		if initialCap == "" {
			t.Fatal("Expected non-empty capability for config_server")
		}

		// Inject a LiveConfig update modifying port to 9999
		Set(handle, "config_server", "port", "9999")

		// Query capability again
		updatedCap, err := GetCapability(handle, "config_server")
		if err != nil {
			t.Fatalf("GetCapability failed after update: %v", err)
		}

		var parsed map[string]interface{}
		if err := json.Unmarshal([]byte(updatedCap), &parsed); err != nil {
			t.Fatalf("Failed to parse capability JSON: %v", err)
		}

		if parsed["port"] != "9999" {
			t.Errorf("Expected LiveConfig overridden port 9999, got %v", parsed["port"])
		}
	})
}

// -----------------------------------------------------------------------------
// HELPERS
// -----------------------------------------------------------------------------

func testInit(profile string) uintptr {
	return New(profile)
}
