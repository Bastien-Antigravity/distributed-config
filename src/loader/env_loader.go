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

func setCapabilityEnv(config *core.Config, capName, key, envVar string) {
	val, exists := os.LookupEnv(envVar)
	if !exists {
		return
	}
	if config.Capabilities == nil {
		config.Capabilities = make(map[string]interface{})
	}
	if config.Capabilities[capName] == nil {
		config.Capabilities[capName] = make(map[string]interface{})
	}
	if m, ok := config.Capabilities[capName].(map[string]interface{}); ok {
		m[key] = val
	}
}

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

	// Environment-First bootstrapping for mandatory infrastructure services
	setCapabilityEnv(config, "config_server", "ip", "CF_IP")
	setCapabilityEnv(config, "config_server", "port", "CF_PORT")
	setCapabilityEnv(config, "log_server", "ip", "LS_IP")
	setCapabilityEnv(config, "log_server", "port", "LS_PORT")
}
