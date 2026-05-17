package cgo_bridge

import (
	"os"
	"testing"
	"time"
)

// -------------------------------------------------------------------------

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

// -------------------------------------------------------------------------

func TestBridge_Security(t *testing.T) {
	t.Run("Decrypt configuration values", func(t *testing.T) {
		// Decrypt is static in this version
		_, _ = Decrypt("test-ciphertext")
	})
}

// -------------------------------------------------------------------------
// HELPERS
// -------------------------------------------------------------------------

func testInit(profile string) uintptr {
	return New(profile)
}
