package loader

// =============================================================================
// ESSENTIAL PROCESS:
// Unit test suite verifying common configuration validation rules and
// mandatory service schema validation logic.
//
// DATA FLOW:
// 1. Input: Incomplete and complete *core.Config test instances.
// 2. Logic: Invokes CheckTestIPs and CheckProductionIPs and checks expected error responses.
// 3. Output: Test assertion results verifying validation boundaries.
//
// KEY PARAMETERS:
// - t: Testing harness handle.
// =============================================================================

import (
	"testing"

	"github.com/Bastien-Antigravity/distributed-config/src/core"
)

// -----------------------------------------------------------------------------

func TestValidator(t *testing.T) {
	t.Run("TestCheckTestIPs", func(t *testing.T) {
		cfg := &core.Config{}
		cfg.Capabilities = make(map[string]interface{})
		cfg.Capabilities["config_server"] = map[string]interface{}{"ip": "127.0.0.2"}
		cfg.Capabilities["timescale_db"] = map[string]interface{}{"ip": "127.0.0.2"}

		// 1. Valid (All are 127.0.0.2 or empty)
		if err := CheckTestIPs(cfg); err != nil {
			t.Errorf("Expected success for test IPs, got error: %v", err)
		}

		// 2. Invalid (Real IP in test mode)
		cfg.Capabilities["timescale_db"].(map[string]interface{})["ip"] = "192.168.1.50"
		if err := CheckTestIPs(cfg); err == nil {
			t.Error("Expected error when non-test IP is used in test mode, got nil")
		}
	})

	t.Run("TestCheckProductionIPs", func(t *testing.T) {
		cfg := &core.Config{}
		cfg.Capabilities = make(map[string]interface{})
		cfg.Capabilities["config_server"] = map[string]interface{}{"ip": "54.12.34.56"}
		cfg.Capabilities["timescale_db"] = map[string]interface{}{"ip": "10.0.0.5"}

		// 1. Valid (No 127.0.0.2)
		if err := CheckProductionIPs(cfg); err != nil {
			t.Errorf("Expected success for production IPs, got error: %v", err)
		}

		// 2. Invalid (Test IP leaked to production when DOCKER_ENV != "true")
		t.Setenv("DOCKER_ENV", "")
		cfg.Capabilities["timescale_db"].(map[string]interface{})["ip"] = "127.0.0.2"
		if err := CheckProductionIPs(cfg); err == nil {
			t.Error("Expected error when test IP (127.0.0.2) is detected in production, got nil")
		}

		// 3. Allowed inside Docker container (DOCKER_ENV == "true")
		t.Setenv("DOCKER_ENV", "true")
		if err := CheckProductionIPs(cfg); err != nil {
			t.Errorf("Expected 127.0.0.2 to be allowed when DOCKER_ENV=true, got error: %v", err)
		}
	})
}
