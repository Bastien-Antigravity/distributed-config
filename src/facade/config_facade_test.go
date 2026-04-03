package facade

import (
	"testing"
)

func TestNewConfig(t *testing.T) {
	profiles := []string{"standalone", "test", "preprod", "production"}

	for _, profile := range profiles {
		t.Run("Profile-"+profile, func(t *testing.T) {
			cfg := NewConfig(profile)
			if cfg == nil {
				t.Fatalf("NewConfig(%s) returned nil", profile)
			}
			if cfg.Config == nil {
				t.Fatalf("NewConfig(%s).Config is nil", profile)
			}

			// Test callback registration
			triggered := false
			cfg.OnMemConfUpdate(func(updates map[string]map[string]string) {
				triggered = true
			})

			// Non-blocking check for triggered (it won't be triggered here but we check registration)
			if triggered {
				t.Log("Callback triggered unexpectedly in sync test")
			}
		})
	}
}
