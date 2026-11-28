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

		// 3. Invalid grpc_ip / rest_ip in test mode
		cfg.Capabilities["timescale_db"] = map[string]interface{}{"ip": "127.0.0.2", "grpc_ip": "10.0.0.8"}
		if err := CheckTestIPs(cfg); err == nil {
			t.Error("Expected error when non-test grpc_ip is used in test mode, got nil")
		}
		cfg.Capabilities["timescale_db"] = map[string]interface{}{"ip": "127.0.0.2", "rest_ip": "10.0.0.9"}
		if err := CheckTestIPs(cfg); err == nil {
			t.Error("Expected error when non-test rest_ip is used in test mode, got nil")
		}

		// 4. Wildcard bind IPs allowed in test mode
		cfg.Capabilities["timescale_db"] = map[string]interface{}{"ip": "0.0.0.0", "grpc_ip": "::"}
		if err := CheckTestIPs(cfg); err != nil {
			t.Errorf("Expected wildcard bind IPs to be allowed in test mode, got: %v", err)
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

		// 2b. Invalid grpc_ip / rest_ip (Test IP leaked to production)
		cfg.Capabilities["timescale_db"] = map[string]interface{}{"ip": "10.0.0.5", "grpc_ip": "127.0.0.2"}
		if err := CheckProductionIPs(cfg); err == nil {
			t.Error("Expected error when test grpc_ip (127.0.0.2) is detected in production, got nil")
		}
		cfg.Capabilities["timescale_db"] = map[string]interface{}{"ip": "10.0.0.5", "rest_ip": "127.0.0.2"}
		if err := CheckProductionIPs(cfg); err == nil {
			t.Error("Expected error when test rest_ip (127.0.0.2) is detected in production, got nil")
		}

		// 3. Allowed inside Docker container (DOCKER_ENV == "true")
		t.Setenv("DOCKER_ENV", "true")
		if err := CheckProductionIPs(cfg); err != nil {
			t.Errorf("Expected 127.0.0.2 to be allowed when DOCKER_ENV=true, got error: %v", err)
		}
	})
}
