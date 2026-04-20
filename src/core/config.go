package core

import (
	"encoding/json"
	"fmt"
	"github.com/Bastien-Antigravity/distributed-config/src/utils"
)

// Common Config
// -----------------------------------------------------------------------------

type CommonConfig struct {
	Name           string `yaml:"name" json:"name"`
	CommonFilePath string `yaml:"common_file_path" json:"common_file_path"`
	Reset          bool   `yaml:"reset" json:"reset"`
}

// Private Config
// -----------------------------------------------------------------------------

type PrivateConfig struct {
	Name            string                 `yaml:"name" json:"name"`
	PrivateFilePath string                 `yaml:"private_file_path" json:"private_file_path"`
	Private         map[string]interface{} `yaml:"private" json:"private"`
}

// Config Data Struct (Pure Data)
// -----------------------------------------------------------------------------



type Config struct {
	// Distributed system name
	Common       CommonConfig           `yaml:"common" json:"common"`

	// Data storage for Config params 
	// main config 
	Capabilities map[string]interface{} `yaml:"capabilities" json:"capabilities"`
	LiveConfig map[string]map[string]string `yaml:"-"`
	PrivateConfig	      map[string]interface{} `yaml:"private" json:"private"`

	// Internal state
	ConfigPath string       `yaml:"-"`
	Logger         utils.Logger `yaml:"-"`
}	

// -----------------------------------------------------------------------------

// Get returns a value from a specified section and key.
// Returns an empty string if not found.
func (c *Config) Get(section, key string) string {
	if s, ok := c.LiveConfig[section]; ok {
		if val, ok := s[key]; ok {
			return val
		}
	}
	return ""
}

// -----------------------------------------------------------------------------

// Set sets a value for a specified section and key.
// Initializes LiveConfig and section maps if they are nil.
func (c *Config) Set(section, key, value string) {
	if c.LiveConfig == nil {
		c.LiveConfig = make(map[string]map[string]string)
	}
	if _, ok := c.LiveConfig[section]; !ok {
		c.LiveConfig[section] = make(map[string]string)
	}
	c.LiveConfig[section][key] = value
}

// -----------------------------------------------------------------------------

// GetCapability extracts a specific capability dictionary and unmarshals it into the target struct.
// It uses JSON round-tripping for easy conversion from nested map[string]interface{} to strongly typed structs.
func (c *Config) GetCapability(key string, target interface{}) error {
	val, ok := c.Capabilities[key]
	if !ok || val == nil {
		return fmt.Errorf("capability '%s' is strictly required but missing", key)
	}
	data, err := json.Marshal(val)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, target)
}

// ValidateMandatoryServices checks the current configuration against the defined strict service schemas.
// It fails if any ecosystem-standard service (Log, Config, etc.) is missing mandatory parameters.
// -----------------------------------------------------------------------------

func (c *Config) ValidateMandatoryServices() error {
	// 1. Log Server
	var ls LogServerCap
	if err := c.GetCapability("log_server", &ls); err != nil {
		return err
	}
	if err := ls.Validate(); err != nil {
		return err
	}

	// 2. Config Server
	var cs ConfigServerCap
	if err := c.GetCapability("config_server", &cs); err != nil {
		return err
	}
	if err := cs.Validate(); err != nil {
		return err
	}

	// Optional check for notification or other systems can be added here
	// consistently with the strictness level required.

	return nil
}

// -----------------------------------------------------------------------------

// ShareObject serializes any arbitrary struct/payload into the target section
// of LiveConfig so it can be broadcasted globally.
func (c *Config) ShareObject(sectionKey string, payload interface{}) error {
	if payload == nil {
		return nil
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	c.Set(sectionKey, "shared_data", string(data))
	return nil
}

// -----------------------------------------------------------------------------

// GetAddress returns the address (host:port) for a given capability.
// It looks for "ip" and "port" keys in the capability configuration.
func (c *Config) GetAddress(capability string) (string, error) {
	return c.getAddr(capability, "ip", "port")
}

// -----------------------------------------------------------------------------

// GetGRPCAddress returns the gRPC address for a given capability.
// Requires strict declaration of 'grpc_ip' and 'grpc_port'.
func (c *Config) GetGRPCAddress(capability string) (string, error) {
	return c.getAddr(capability, "grpc_ip", "grpc_port")
}

func (c *Config) getAddr(capability, hostKey, portKey string) (string, error) {
	if c.Capabilities == nil {
		return "", fmt.Errorf("no capabilities found")
	}
	capRaw, ok := c.Capabilities[capability]
	if !ok {
		return "", fmt.Errorf("capability %s not found", capability)
	}

	cap, ok := capRaw.(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("invalid capability format for %s", capability)
	}

	host, ok := cap[hostKey].(string)
	if !ok || host == "" {
		return "", fmt.Errorf("host key %s missing or empty in capability %s", hostKey, capability)
	}

	p, ok := cap[portKey].(string)
	if !ok || p == "" {
		return "", fmt.Errorf("port key %s missing or empty in capability %s", portKey, capability)
	}

	return fmt.Sprintf("%s:%s", host, p), nil
}
