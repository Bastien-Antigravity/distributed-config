package loader

import (
	"testing"

	"github.com/Bastien-Antigravity/distributed-config/src/core"
	"github.com/Bastien-Antigravity/distributed-config/src/models"
)

func TestValidator(t *testing.T) {
	t.Run("TestValidateCommonConfig", func(t *testing.T) {
		cfg := &core.Config{}
		
		// 1. Missing Name
		if err := ValidateCommonConfig(cfg); err == nil {
			t.Error("Expected error for missing name, got nil")
		}

		cfg.Common.Name = "test-app"
		// 2. Missing Server Details
		if err := ValidateCommonConfig(cfg); err == nil {
			t.Error("Expected error for missing server details, got nil")
		}

		cfg.Capabilities.ConfigServer = &models.ConfigServerCapability{
			IP:   "127.0.0.1",
			Port: "5000",
		}
		// 3. Valid
		if err := ValidateCommonConfig(cfg); err != nil {
			t.Errorf("Expected success, got error: %v", err)
		}
	})

	t.Run("TestCheckTestIPs", func(t *testing.T) {
		cfg := &core.Config{}
		cfg.Capabilities.ConfigServer = &models.ConfigServerCapability{IP: "127.0.0.2"}
		cfg.Capabilities.TimescaleDb = &models.TimescaleDbCapability{IP: "127.0.0.2"}

		// 1. Valid (All are 127.0.0.2 or empty)
		if err := CheckTestIPs(cfg); err != nil {
			t.Errorf("Expected success for test IPs, got error: %v", err)
		}

		// 2. Invalid (Real IP in test mode)
		cfg.Capabilities.TimescaleDb.IP = "192.168.1.50"
		if err := CheckTestIPs(cfg); err == nil {
			t.Error("Expected error when non-test IP is used in test mode, got nil")
		}
	})

	t.Run("TestCheckProductionIPs", func(t *testing.T) {
		cfg := &core.Config{}
		cfg.Capabilities.ConfigServer = &models.ConfigServerCapability{IP: "54.12.34.56"}
		cfg.Capabilities.TimescaleDb = &models.TimescaleDbCapability{IP: "10.0.0.5"}

		// 1. Valid (No 127.0.0.2)
		if err := CheckProductionIPs(cfg); err != nil {
			t.Errorf("Expected success for production IPs, got error: %v", err)
		}

		// 2. Invalid (Test IP leaked to production)
		cfg.Capabilities.TimescaleDb.IP = "127.0.0.2"
		if err := CheckProductionIPs(cfg); err == nil {
			t.Error("Expected error when test IP (127.0.0.2) is detected in production, got nil")
		}
	})
}
