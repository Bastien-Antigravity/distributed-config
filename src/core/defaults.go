package core

// =============================================================================
// ESSENTIAL PROCESS:
// Instantiates default in-memory configuration structures pre-populated with
// ecosystem-wide fallback ports, hosts, and capability definitions.
//
// DATA FLOW:
// 1. Input: None (hardcoded fleet defaults).
// 2. Logic: Constructs a *Config populated with standard capability maps and common settings.
// 3. Output: Initialized baseline *Config struct for zero-config fallbacks and skeleton generation.
//
// KEY PARAMETERS:
// - CF_PORT: Default config_server port fallback (3306).
// - LS_PORT: Default log_server port fallback (9020).
// =============================================================================

// -----------------------------------------------------------------------------

func NewDefaultConfig() *Config {
	initialLive := make(map[string]map[string]string)
	cfg := &Config{
		Common: CommonConfig{
			Name:        "common",
			Reset:       false,
			PublicIP:    "127.0.0.1",
			RetryBaseMS: "100",
			RetryMaxSec: "5",
		},
		Capabilities: make(map[string]interface{}),
	}
	cfg.LiveConfig.Store(&initialLive)
	return cfg
}
