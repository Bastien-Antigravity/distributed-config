package loader

// =============================================================================
// ESSENTIAL PROCESS:
// Environment variable extractor populating CommonConfig parameters from system
// environment variables (e.g., NAME, RESET, PUBLIC_KEY).
//
// DATA FLOW:
// 1. Input: OS environment variables via os.LookupEnv.
// 2. Logic: Checks for variable presence and parses typed values (strings, booleans).
// 3. Output: Updates CommonConfig fields on the target *core.Config pointer.
//
// KEY PARAMETERS:
// - NAME: Service name identifier.
// - RESET: Database/state reset flag.
// =============================================================================

import (
	"os"
	"strings"

	"github.com/Bastien-Antigravity/distributed-config/src/core"
)

// -----------------------------------------------------------------------------

func LoadCommonFromEnv(config *core.Config) {
	if val, exists := os.LookupEnv("NAME"); exists {
		config.Common.Name = val
	}

	if val, exists := os.LookupEnv("RESET"); exists {
		if strings.ToLower(val) == "true" {
			config.Common.Reset = true
		} else {
			config.Common.Reset = false
		}
	}

	// Wait: generic approach for ConfigServer bootstrapping if missing YAML
	if val, exists := os.LookupEnv("CF_IP"); exists {
		if config.Capabilities == nil {
			config.Capabilities = make(map[string]interface{})
		}
		if config.Capabilities["config_server"] == nil {
			config.Capabilities["config_server"] = make(map[string]interface{})
		}
		// Go type assertion trick for nested maps
		if m, ok := config.Capabilities["config_server"].(map[string]interface{}); ok {
			m["ip"] = val
		}
	}
	if val, exists := os.LookupEnv("CF_PORT"); exists {
		if config.Capabilities == nil {
			config.Capabilities = make(map[string]interface{})
		}
		if config.Capabilities["config_server"] == nil {
			config.Capabilities["config_server"] = make(map[string]interface{})
		}
		if m, ok := config.Capabilities["config_server"].(map[string]interface{}); ok {
			m["port"] = val
		}
	}
}
