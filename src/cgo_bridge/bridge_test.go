package main

import (
	"os"
	"testing"
	"time"
)

// -------------------------------------------------------------------------

func TestBridge_ExpandedName(t *testing.T) {
	// Set the environment variable for testing expansion
	os.Setenv("APP_NAME", "dynamic-bridge-app")
	defer os.Unsetenv("APP_NAME")

	t.Run("Initialize and check expanded name", func(t *testing.T) {
		handle := testInit("standalone")
		defer Close(handle)

		facadeMu.Lock()
		session := facadeStore[handle]
		facadeMu.Unlock()

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

		facadeMu.Lock()
		session := facadeStore[handle]
		facadeMu.Unlock()

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
<<<<<<< HEAD
		defer DistConf_Close(handle)

		facadeMu.Lock()
		session := facadeStore[handle]
		facadeMu.Unlock()
=======
		defer Close(handle)
>>>>>>> develop

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
<<<<<<< HEAD
	cfg := distributed_config.New(profile)
	if cfg == nil {
		return 0
	}
	facadeMu.Lock()
	defer facadeMu.Unlock()
	id := facadeId
	facadeStore[id] = &ConfigSession{Config: cfg}
	facadeId++
	return id
}

func DistConf_Sync_Internal(handle uintptr) int {
	facadeMu.Lock()
	session, ok := facadeStore[handle]
	facadeMu.Unlock()
	if !ok {
		return 0
	}
	if err := session.Config.Sync(); err != nil {
		return 0
	}
	return 1
}

func DistConf_ShareConfig_Internal(handle uintptr, jsonData string) int {
	facadeMu.Lock()
	session, ok := facadeStore[handle]
	facadeMu.Unlock()
	if !ok {
		return 0
	}
	var payload interface{}
	_ = json.Unmarshal([]byte(jsonData), &payload)
	if err := session.Config.ShareConfig(payload); err != nil {
		return 0
	}
	return 1
}

func DistConf_ValidateMandatoryServices_Internal(handle uintptr) int {
	facadeMu.Lock()
	session, ok := facadeStore[handle]
	facadeMu.Unlock()
	if !ok {
		return 0
	}
	if err := session.Config.ValidateMandatoryServices(); err != nil {
		return 0
	}
	return 1
=======
	return New(profile)
>>>>>>> develop
}
