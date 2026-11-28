package facade

// =============================================================================
// ESSENTIAL PROCESS:
// Unit test suite verifying facade initialization, multi-profile loading, and
// local callback execution upon SetSingle mutations.
//
// DATA FLOW:
// 1. Input: Profile names ("standalone", "test", "staging", "production").
// 2. Logic: Initializes facade for each profile and validates callback triggers.
// 3. Output: Test assertion results verifying strict eventing and value retention.
//
// KEY PARAMETERS:
// - t: Testing context handle.
// =============================================================================

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
)

// -----------------------------------------------------------------------------

func writeTempConfigFixture(t *testing.T) {
	t.Helper()
	closedPort := 59999
	if ln, err := net.Listen("tcp", "127.0.0.1:0"); err == nil {
		closedPort = ln.Addr().(*net.TCPAddr).Port
		_ = ln.Close()
	}
	tempDir := t.TempDir()
	fixturePath := filepath.Join(tempDir, "facade_test.yaml")
	fixture := fmt.Sprintf(`
common:
  name: "facade-test"
capabilities:
  log_server:
    ip: "127.0.0.1"
    port: "9020"
  config_server:
    ip: "127.0.0.1"
    port: "%d"
`, closedPort)
	if err := os.WriteFile(fixturePath, []byte(fixture), 0644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("CONFIG_PATH", fixturePath)
	t.Setenv("DOCKER_ENV", "true")
}

// -----------------------------------------------------------------------------

func TestNewConfig(t *testing.T) {
	writeTempConfigFixture(t)
	profiles := []string{"standalone", "test", "staging", "production"}

	for _, profile := range profiles {
		t.Run("Profile-"+profile, func(t *testing.T) {
			cfg := NewConfig(profile)
			if cfg == nil {
				t.Fatalf("NewConfig(%s) returned nil", profile)
			}
			defer func() { _ = cfg.Close() }()
			if cfg.Config == nil {
				t.Fatalf("NewConfig(%s).Config is nil", profile)
			}

			// Test callback registration & local trigger via SetSingle
			var triggered atomic.Bool
			var mu sync.Mutex
			var capturedUpdates map[string]map[string]string
			cfg.OnLiveConfUpdate(func(updates map[string]map[string]string) {
				mu.Lock()
				capturedUpdates = updates
				mu.Unlock()
				triggered.Store(true)
			})

			_ = cfg.SetSingle("TEST_SECTION", "TEST_KEY", "TEST_VAL")

			if cfg.strategy != nil {
				expectTrigger := profile == "standalone" || profile == "staging"

				if expectTrigger && !triggered.Load() {
					t.Errorf("NewConfig(%s): SetSingle should have triggered the local callback", profile)
				} else if !expectTrigger && triggered.Load() {
					t.Errorf("NewConfig(%s): SetSingle should NOT have triggered the local callback (Strict Eventing)", profile)
				}

				if expectTrigger {
					mu.Lock()
					gotVal := capturedUpdates["TEST_SECTION"]["TEST_KEY"]
					mu.Unlock()
					if gotVal != "TEST_VAL" {
						t.Errorf("NewConfig(%s): Callback captured wrong value: %v", profile, gotVal)
					}
				}
			}

			if cfg.Get("TEST_SECTION", "TEST_KEY") != "TEST_VAL" {
				t.Errorf("NewConfig(%s): Get should return values set via SetSingle", profile)
			}
		})
	}
}

// -----------------------------------------------------------------------------

func TestFacade_OCCAndDeletions(t *testing.T) {
	writeTempConfigFixture(t)
	cfg := NewConfig("standalone")
	if cfg == nil {
		t.Fatal("Failed to create standalone config")
	}

	// 1. Initial GetVersion
	if cfg.GetVersion() != 0 {
		t.Errorf("Expected initial version 0, got %d", cfg.GetVersion())
	}

	// 2. SetWithOCC
	err := cfg.SetWithOCC(map[string]map[string]string{
		"DATABASE": {
			"HOST": "127.0.0.1",
			"PORT": "5432",
		},
		"APP": {
			"ENV": "test",
		},
	}, 0)
	if err != nil {
		t.Fatalf("SetWithOCC failed: %v", err)
	}

	if cfg.Get("DATABASE", "HOST") != "127.0.0.1" {
		t.Errorf("Expected DATABASE.HOST=127.0.0.1, got %v", cfg.Get("DATABASE", "HOST"))
	}

	// 3. DeleteKey (single key)
	if err := cfg.DeleteKey("DATABASE", "PORT"); err != nil {
		t.Fatalf("DeleteKey failed: %v", err)
	}
	if cfg.Get("DATABASE", "PORT") != "" {
		t.Errorf("Expected DATABASE.PORT to be empty after deletion, got %v", cfg.Get("DATABASE", "PORT"))
	}
	if cfg.Get("DATABASE", "HOST") != "127.0.0.1" {
		t.Errorf("Expected DATABASE.HOST to be retained, got %v", cfg.Get("DATABASE", "HOST"))
	}

	// 4. DeleteKeys (multiple keys / entire section)
	if err := cfg.DeleteKeys(map[string][]string{
		"DATABASE": {"HOST"},
		"APP":      {}, // empty slice deletes whole APP section
	}); err != nil {
		t.Fatalf("DeleteKeys failed: %v", err)
	}

	if cfg.Get("DATABASE", "HOST") != "" {
		t.Errorf("Expected DATABASE.HOST to be deleted")
	}
	if cfg.Get("APP", "ENV") != "" {
		t.Errorf("Expected APP section to be deleted")
	}

	// 5. DeleteKeysWithOCC
	_ = cfg.SetSingle("SEC", "K", "V")
	if err := cfg.DeleteKeysWithOCC(map[string][]string{"SEC": {"K"}}, 0); err != nil {
		t.Fatalf("DeleteKeysWithOCC failed: %v", err)
	}
	if cfg.Get("SEC", "K") != "" {
		t.Errorf("Expected SEC.K to be deleted via OCC delete")
	}

	// 6. ShareConfigWithOCC
	if err := cfg.ShareConfigWithOCC(map[string]interface{}{
		"SHARED_SEC": map[string]interface{}{
			"PARAM": "VALUE",
		},
	}, 0); err != nil {
		t.Fatalf("ShareConfigWithOCC failed: %v", err)
	}
	if cfg.Get("SHARED_SEC", "PARAM") != "VALUE" {
		t.Errorf("Expected SHARED_SEC.PARAM=VALUE, got %v", cfg.Get("SHARED_SEC", "PARAM"))
	}
}
