package loader

// =============================================================================
// ESSENTIAL PROCESS:
// Configuration integrity validator ensuring required service capabilities
// (config_server, log_server, etc.) are present and valid before service boot.
//
// DATA FLOW:
// 1. Input: Populated *core.Config pointer.
// 2. Logic: Enforces test and production IP isolation rules across capability maps.
// 3. Output: Error if forbidden capability IP addresses are detected.
//
// KEY PARAMETERS:
// - DOCKER_ENV: Allows Docker bridge/loopback IPs when set to "true".
// =============================================================================

import (
	"fmt"
	"os"

	"github.com/Bastien-Antigravity/distributed-config/src/core"
)

// -----------------------------------------------------------------------------

func CheckTestIPs(config *core.Config) error {
	for name, capInterface := range config.Capabilities {
		if m, ok := capInterface.(map[string]interface{}); ok {
			if ipRaw, exists := m["ip"]; exists {
				if ip, ok := ipRaw.(string); ok && ip != "127.0.0.2" && ip != "" {
					return fmt.Errorf("test integrity failure: %s IP must be 127.0.0.2 in test mode, got %s", name, ip)
				}
			}
		}
	}
	return nil
}

// CheckProductionIPs ensures NO IP is 127.0.0.2 in bare-metal/VPS production.
// Inside Docker containers (DOCKER_ENV=true), 127.0.0.2 is allowed as the Docker-mode fallback IP.
// -----------------------------------------------------------------------------

func CheckProductionIPs(config *core.Config) error {
	if os.Getenv("DOCKER_ENV") == "true" {
		return nil
	}
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
