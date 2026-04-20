package core

import "fmt"

// LogServerCap defines the mandatory capabilities for the centralized logging server.
type LogServerCap struct {
	IP   string `json:"ip"`
	Port string `json:"port"`
}

// Validate ensures all mandatory fields are present.
func (l *LogServerCap) Validate() error {
	if l.IP == "" || l.Port == "" {
		return fmt.Errorf("log_server: ip and port are mandatory")
	}
	return nil
}

// ConfigServerCap defines the mandatory capabilities for the centralized configuration registry.
type ConfigServerCap struct {
	IP   string `json:"ip"`
	Port string `json:"port"`
}

// Validate ensures all mandatory fields are present.
func (c *ConfigServerCap) Validate() error {
	if c.IP == "" || c.Port == "" {
		return fmt.Errorf("config_server: ip and port are mandatory")
	}
	return nil
}
