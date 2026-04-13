package core

import (
	"encoding/json"
	"fmt"
	"github.com/Bastien-Antigravity/distributed-config/src/utils"
)

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
	COMMON_FILE_PATH string       `yaml:"-"`
	MemConfigPath    string       `yaml:"-"`
	Logger           utils.Logger `yaml:"-"`

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

// -----------------------------------------------------------------------------

// GetAddress returns the address (host:port) for a given capability.
// It looks for "ip" and "port" keys in the capability configuration.
func (c *Config) GetAddress(capability string) (string, error) {
	return c.getAddr(capability, "ip", "port")
}

// GetGRPCAddress returns the gRPC address for a given capability.
// It first looks for "grpc_ip" and "grpc_port". If not found, it falls back
// to the "ip" and "port" (with port incremented by 1).
func (c *Config) GetGRPCAddress(capability string) (string, error) {
	addr, err := c.getAddr(capability, "grpc_ip", "grpc_port")
	if err == nil {
		return addr, nil
	}

	// Fallback to convention: ip:port+1
	capRaw, ok := c.Capabilities[capability]
	if !ok {
		return "", fmt.Errorf("capability %s not found for gRPC fallback", capability)
	}
	cap, ok := capRaw.(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("invalid format for capability %s", capability)
	}

	host := "0.0.0.0"
	if h, ok := cap["ip"].(string); ok && h != "" {
		host = h
	}

	port := 8080
	if p, ok := cap["port"].(string); ok && p != "" {
		fmt.Sscanf(p, "%d", &port)
	}

	return fmt.Sprintf("%s:%d", host, port+1), nil
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

	host := "0.0.0.0"
	if h, ok := cap[hostKey].(string); ok && h != "" {
		host = h
	}

	p, ok := cap[portKey].(string)
	if !ok || p == "" {
		return "", fmt.Errorf("port key %s missing or empty in capability %s", portKey, capability)
	}

	return fmt.Sprintf("%s:%s", host, p), nil
}
