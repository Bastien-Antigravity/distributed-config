package cgo_bridge

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Bastien-Antigravity/distributed-config"
)

// This test verifies the bridge logic against the "Standard ecosystem scenarios"
// described in TESTING.md.

func TestBridgeStandardScenarios(t *testing.T) {
	// 1. Environment Expansion
	t.Run("EnvironmentExpansion", func(t *testing.T) {
		tempDir := t.TempDir()
		oldCwd, _ := os.Getwd()
		_ = os.Chdir(tempDir)
		defer func() { _ = os.Chdir(oldCwd) }()

		t.Setenv("TEST_APP_NAME", "dynamic-bridge-app")
		yamlContent := `
common:
  name: "${TEST_APP_NAME:fallback}"
`
		_ = os.MkdirAll("config", 0755)
		_ = os.WriteFile("config/standalone.yaml", []byte(yamlContent), 0644)

		handle := testInit("standalone")
		defer DistConf_Close(handle)

		facadeMu.Lock()
		session := facadeStore[handle]
		facadeMu.Unlock()

		if session.Config.Common.Name != "dynamic-bridge-app" {
			t.Errorf("Expected expanded name 'dynamic-bridge-app', got '%s'", session.Config.Common.Name)
		}
	})

	// 2. Auto-Generation
	t.Run("AutoGeneration", func(t *testing.T) {
		tempDir := t.TempDir()
		oldCwd, _ := os.Getwd()
		_ = os.Chdir(tempDir)
		defer func() { _ = os.Chdir(oldCwd) }()

		handle := testInit("standalone")
		defer DistConf_Close(handle)

		// Check for generated file
		found := false
		_ = filepath.Walk(tempDir, func(path string, info os.FileInfo, err error) error {
			if !info.IsDir() && (filepath.Base(path) == "standalone.yaml" || filepath.Base(path) == "cgo_bridge.yaml") {
				found = true
			}
			return nil
		})

		if !found {
			t.Error("Standalone profile should auto-generate missing config file")
		}
	})

	// 3. Callback Integrity
	t.Run("CallbackIntegrity", func(t *testing.T) {
		tempDir := t.TempDir()
		oldCwd, _ := os.Getwd()
		_ = os.Chdir(tempDir)
		defer func() { _ = os.Chdir(oldCwd) }()

		handle := testInit("standalone")
		defer DistConf_Close(handle)

		facadeMu.Lock()
		session := facadeStore[handle]
		facadeMu.Unlock()

		updated := make(chan bool, 1)
		session.Config.OnLiveConfUpdate(func(updates map[string]map[string]string) {
			if updates["common"]["name"] == "updated-via-bridge" {
				updated <- true
			}
		})

		_ = session.Config.SetSingle("common", "name", "updated-via-bridge")

		select {
		case <-updated:
			// Success
		case <-time.After(200 * time.Millisecond):
			t.Error("Callback not triggered")
		}
	})

	// 4. Sync and Share
	t.Run("SyncAndShare", func(t *testing.T) {
		tempDir := t.TempDir()
		oldCwd, _ := os.Getwd()
		_ = os.Chdir(tempDir)
		defer func() { _ = os.Chdir(oldCwd) }()

		handle := testInit("standalone")
		defer DistConf_Close(handle)

		facadeMu.Lock()
		session := facadeStore[handle]
		facadeMu.Unlock()

		// Test Sync
		if res := DistConf_Sync_Internal(handle); res != 1 {
			t.Error("Sync failed")
		}

		// Test ShareConfig
		payload := `{"status": "online"}`
		if res := DistConf_ShareConfig_Internal(handle, payload); res != 1 {
			t.Error("ShareConfig failed")
		}

		// Verify (ShareConfig with flat map uses "shared" section by default)
		val := session.Config.Get("shared", "status")
		if val != "online" {
			t.Fatalf("Shared data not found or incorrect: %s", val)
		}
	})

	// 5. Validation
	t.Run("Validation", func(t *testing.T) {
		tempDir := t.TempDir()
		oldCwd, _ := os.Getwd()
		_ = os.Chdir(tempDir)
		defer func() { _ = os.Chdir(oldCwd) }()

		handle := testInit("standalone")
		defer DistConf_Close(handle)

		if res := DistConf_ValidateMandatoryServices_Internal(handle); res != 1 {
			t.Error("Validation failed")
		}
	})
}

// Helpers
func testInit(profile string) uintptr {
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
}
