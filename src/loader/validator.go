package loader

import (
	"fmt"

	"github.com/Bastien-Antigravity/distributed-config/src/core"
)

// ValidateCommonConfig checks if the minimal required configuration is present.
// -----------------------------------------------------------------------------

func ValidateCommonConfig(config *core.Config) error {
	if config.Common.Name == "" {
		return fmt.Errorf("missing required common config: NAME")
	}

	// Assuming Config Server details are required for Distributed Mode
	val, ok := config.Capabilities["config_server"]
	if !ok {
		return fmt.Errorf("missing required config server details (CF_IP, CF_PORT) for distributed mode")
	}
	if m, ok := val.(map[string]interface{}); ok {
		ip, _ := m["ip"].(string)
		port, _ := m["port"].(string)
		if ip == "" || port == "" {
			return fmt.Errorf("missing required config server details")
		}
	} else {
		return fmt.Errorf("invalid config_server capability map")
	}

	return nil
}

// CheckTestIPs ensures all defined IPs are 127.0.0.2
// -----------------------------------------------------------------------------

func CheckTestIPs(config *core.Config) error {
	for name, capInterface := range config.Capabilities {
		if m, ok := capInterface.(map[string]interface{}); ok {
			if ipRaw, exists := m["ip"]; exists {
				if ip, ok := ipRaw.(string); ok && ip != "" && ip != "127.0.0.2" {
					return fmt.Errorf("test integrity failure: %s IP must be 127.0.0.2, got %s", name, ip)
				}
			}
		}
	}
	return nil
}

// CheckProductionIPs ensures NO IP is 127.0.0.2
// -----------------------------------------------------------------------------

func CheckProductionIPs(config *core.Config) error {
	for name, capInterface := range config.Capabilities {
		if m, ok := capInterface.(map[string]interface{}); ok {
			if ipRaw, exists := m["ip"]; exists {
				if ip, ok := ipRaw.(string); ok && ip == "127.0.0.2" {
					return fmt.Errorf("production integrity failure: %s IP cannot be 127.0.0.2 (Test IP detected)", name)
				}
			}
		}
	}
	return nil
}
