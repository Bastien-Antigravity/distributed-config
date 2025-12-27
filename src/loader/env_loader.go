package loader

import (
	"os"
	"strings"

	"github.com/Bastien-Antigravity/distributed-config/src/core"
	"github.com/Bastien-Antigravity/distributed-config/src/models"
)

// LoadCommonFromEnv populates the Common configuration from Environment Variables.
// It maps specific ENV keys (e.g., "NAME", "RESET") to the struct.
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

	// Example: If Common requires Config Server IP to be bootstrapped from Env
	if val, exists := os.LookupEnv("CF_IP"); exists {
		if config.Capabilities.ConfigServer == nil {
			config.Capabilities.ConfigServer = &models.ConfigServerCapability{}
		}
		config.Capabilities.ConfigServer.IP = val
	}
	if val, exists := os.LookupEnv("CF_PORT"); exists {
		if config.Capabilities.ConfigServer == nil {
			config.Capabilities.ConfigServer = &models.ConfigServerCapability{}
		}
		config.Capabilities.ConfigServer.Port = val
	}
}
