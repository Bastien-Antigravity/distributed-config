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
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// -----------------------------------------------------------------------------

func setupBridgeTempConfig(t *testing.T) {
	t.Helper()
	tempDir := t.TempDir()
	fixturePath := filepath.Join(tempDir, "bridge_standalone.yaml")
	fixture := `
common:
  name: "bridge-test"
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
}

// -----------------------------------------------------------------------------

func TestBridge_ExpandedName(t *testing.T) {
	setupBridgeTempConfig(t)
	// Set the environment variable for testing expansion
	t.Setenv("NAME", "dynamic-bridge-app")

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
	setupBridgeTempConfig(t)
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
	setupBridgeTempConfig(t)
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
	setupBridgeTempConfig(t)
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
	setupBridgeTempConfig(t)
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
		if err := Set(handle, "config_server", "port", "9999"); err != nil {
			t.Fatalf("Set failed: %v", err)
		}

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

func TestBridge_ProfileAliasesAndApplyFileOverride(t *testing.T) {
	setupBridgeTempConfig(t)

	for _, alias := range []string{"dev", "stage"} {
		h := New(alias)
		if h == 0 {
			t.Fatalf("Expected non-zero handle for profile alias %q", alias)
		}
		Close(h)
	}

	if bad := New("invalid-profile"); bad != 0 {
		Close(bad)
		t.Fatalf("Expected handle 0 for invalid profile, got %d", bad)
	}

	handle := New("standalone")
	if handle == 0 {
		t.Fatal("Expected non-zero handle for standalone")
	}
	defer Close(handle)

	overridePath := filepath.Join(t.TempDir(), "override.yaml")
	overrideYAML := `
common:
  name: "overridden-app"
capabilities:
  custom_svc:
    ip: "10.20.30.40"
    port: 7777
local:
  feature_flag: true
  token: "false"
`
	if err := os.WriteFile(overridePath, []byte(overrideYAML), 0644); err != nil {
		t.Fatal(err)
	}

	localJSON, err := ApplyFileOverride(handle, overridePath)
	if err != nil {
		t.Fatalf("ApplyFileOverride failed: %v", err)
	}
	if !strings.Contains(localJSON, `"feature_flag":true`) || !strings.Contains(localJSON, `"token":"false"`) {
		t.Fatalf("Unexpected local JSON from ApplyFileOverride: %s", localJSON)
	}

	addr, err := GetAddress(handle, "custom_svc")
	if err != nil || addr != "10.20.30.40:7777" {
		t.Fatalf("Expected custom_svc 10.20.30.40:7777, got %q (err=%v)", addr, err)
	}
}

// -----------------------------------------------------------------------------
// HELPERS
// -----------------------------------------------------------------------------

func testInit(profile string) uintptr {
	return New(profile)
}
