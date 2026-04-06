package loader

import (
	"testing"

	"github.com/Bastien-Antigravity/distributed-config/src/core"
)

func TestValidator(t *testing.T) {
	t.Run("TestValidateCommonConfig", func(t *testing.T) {
		cfg := &core.Config{}
		cfg.Capabilities = make(map[string]interface{})

		// 1. Missing Name
		if err := ValidateCommonConfig(cfg); err == nil {
			t.Error("Expected error for missing name, got nil")
		}

		cfg.Common.Name = "test-app"
		// 2. Missing Server Details
		if err := ValidateCommonConfig(cfg); err == nil {
			t.Error("Expected error for missing server details, got nil")
		}

		cfg.Capabilities["config_server"] = map[string]interface{}{
			"ip":   "127.0.0.1",
			"port": "5000",
		}
		// 3. Valid
		if err := ValidateCommonConfig(cfg); err != nil {
			t.Errorf("Expected success, got error: %v", err)
		}
	})

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

		// 2. Invalid (Test IP leaked to production)
		cfg.Capabilities["timescale_db"].(map[string]interface{})["ip"] = "127.0.0.2"
		if err := CheckProductionIPs(cfg); err == nil {
			t.Error("Expected error when test IP (127.0.0.2) is detected in production, got nil")
		}
	})
}
