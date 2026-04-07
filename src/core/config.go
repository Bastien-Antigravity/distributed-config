package core

import "encoding/json"

// Common Config
// -----------------------------------------------------------------------------

type CommonConfig struct {
	Name  string `yaml:"name" json:"name"`
	Reset bool   `yaml:"reset" json:"reset"`
}

// Config Data Struct (Pure Data)
// -----------------------------------------------------------------------------

type Config struct {
	Common       CommonConfig           `yaml:"common" json:"common"`
	Capabilities map[string]interface{} `yaml:"capabilities" json:"capabilities"`

	// Internal state
	COMMON_FILE_PATH string
	MemConfigPath    string

	// Data storage for MemConfig
	MemConfig map[string]map[string]string `yaml:"-"`
}

// -----------------------------------------------------------------------------

// Get returns a value from a specified section and key.
// Returns an empty string if not found.
func (c *Config) Get(section, key string) string {
	if s, ok := c.MemConfig[section]; ok {
		if val, ok := s[key]; ok {
			return val
		}
	}
	return ""
}

// -----------------------------------------------------------------------------

// Set sets a value for a specified section and key.
// Initializes MemConfig and section maps if they are nil.
func (c *Config) Set(section, key, value string) {
	if c.MemConfig == nil {
		c.MemConfig = make(map[string]map[string]string)
	}
	if _, ok := c.MemConfig[section]; !ok {
		c.MemConfig[section] = make(map[string]string)
	}
	c.MemConfig[section][key] = value
}

// -----------------------------------------------------------------------------

// GetCapability extracts a specific capability dictionary and unmarshals it into the target struct.
// It uses JSON round-tripping for easy conversion from nested map[string]interface{} to strongly typed structs.
func (c *Config) GetCapability(key string, target interface{}) error {
	val, ok := c.Capabilities[key]
	if !ok || val == nil {
		return nil // Not found, but not strictly an error (just empty target)
	}
	data, err := json.Marshal(val)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, target)
}
