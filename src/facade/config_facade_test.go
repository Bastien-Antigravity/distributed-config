package facade

import (
	"testing"
)

func TestNewConfig(t *testing.T) {
	profiles := []string{"standalone", "test", "staging", "production"}

	for _, profile := range profiles {
		t.Run("Profile-"+profile, func(t *testing.T) {
			cfg := NewConfig(profile)
			if cfg == nil {
				t.Fatalf("NewConfig(%s) returned nil", profile)
			}
			if cfg.Config == nil {
				t.Fatalf("NewConfig(%s).Config is nil", profile)
			}

			// Test callback registration & local trigger via SetSingle
			triggered := false
			var capturedUpdates map[string]map[string]string
			cfg.OnLiveConfUpdate(func(updates map[string]map[string]string) {
				triggered = true
				capturedUpdates = updates
			})

			cfg.SetSingle("TEST_SECTION", "TEST_KEY", "TEST_VAL")

			if cfg.strategy != nil {
				expectTrigger := profile == "standalone" || profile == "staging"

				if expectTrigger && !triggered {
					t.Errorf("NewConfig(%s): SetSingle should have triggered the local callback", profile)
				} else if !expectTrigger && triggered {
					t.Errorf("NewConfig(%s): SetSingle should NOT have triggered the local callback (Strict Eventing)", profile)
				}

				if expectTrigger {
					if capturedUpdates["TEST_SECTION"]["TEST_KEY"] != "TEST_VAL" {
						t.Errorf("NewConfig(%s): Callback captured wrong value: %v", profile, capturedUpdates["TEST_SECTION"]["TEST_KEY"])
					}
				}
			}

			if cfg.Get("TEST_SECTION", "TEST_KEY") != "TEST_VAL" {
				t.Errorf("NewConfig(%s): Get should return values set via SetSingle", profile)
			}
		})
	}
}
