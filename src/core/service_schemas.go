package core

// =============================================================================
// ESSENTIAL PROCESS:
// Structural schema definitions and field validators for mandatory ecosystem
// infrastructure capabilities (log_server, config_server).
//
// DATA FLOW:
// 1. Input: Unmarshaled capability JSON/map payloads.
// 2. Logic: Enforces non-empty IP and Port validations for required daemon connectivity.
// 3. Output: Validation error if mandatory capability attributes are missing.
//
// KEY PARAMETERS:
// - LogServerCap: Validated logging daemon endpoint descriptor.
// - ConfigServerCap: Validated configuration registry endpoint descriptor.
// =============================================================================

import "fmt"

// -----------------------------------------------------------------------------

// LogServerCap defines the mandatory capabilities for the centralized logging server.
type LogServerCap struct {
	IP   string      `json:"ip"`
	Port interface{} `json:"port"`
}

// Validate ensures all mandatory fields are present.
func (l *LogServerCap) Validate() error {
	if l.IP == "" || l.Port == nil || fmt.Sprintf("%v", l.Port) == "" {
		return fmt.Errorf("log_server: ip and port are mandatory")
	}
	return nil
}

// -----------------------------------------------------------------------------

// ConfigServerCap defines the mandatory capabilities for the centralized configuration registry.
type ConfigServerCap struct {
	IP   string      `json:"ip"`
	Port interface{} `json:"port"`
}

// Validate ensures all mandatory fields are present.
func (c *ConfigServerCap) Validate() error {
	if c.IP == "" || c.Port == nil || fmt.Sprintf("%v", c.Port) == "" {
		return fmt.Errorf("config_server: ip and port are mandatory")
	}
	return nil
}
