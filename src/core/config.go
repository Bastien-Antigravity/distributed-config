package core

import (
	"encoding/json"
	"fmt"
	"sync/atomic"

	"github.com/Bastien-Antigravity/distributed-config/src/utils"
)

// Common Config
// -----------------------------------------------------------------------------

type CommonConfig struct {
	Name           string `yaml:"name" json:"name"`
	CommonFilePath string `yaml:"common_file_path" json:"common_file_path"`
	PublicKey      string `yaml:"public_key" json:"public_key"`
	Reset          bool   `yaml:"reset" json:"reset"`
}

// Config Data Struct (Pure Data)
// -----------------------------------------------------------------------------

type Config struct {
	// Distributed system name
	Common       CommonConfig           `yaml:"common" json:"common"`

	// Data storage for Config params 
	// main config 
	Capabilities map[string]interface{} `yaml:"capabilities" json:"capabilities"`
	LiveConfig   atomic.Pointer[map[string]map[string]string] `yaml:"-"`

	// Internal state
	ConfigPath string       `yaml:"-"`
	Logger     utils.Logger `yaml:"-"`
}	

// -----------------------------------------------------------------------------

// Get returns a value from a specified section and key.
// Returns an empty string if not found.
func (c *Config) Get(section, key string) string {
	ptr := c.LiveConfig.Load()
	if ptr == nil {
		return ""
	}
	live := *ptr
	if s, ok := live[section]; ok {
		if val, ok := s[key]; ok {
			return val
		}
	}
	return ""
}

// -----------------------------------------------------------------------------

// Set merges multiple configuration updates into the live configuration.
// Performs a thread-safe atomic swap (Read-Copy-Update).
func (c *Config) Set(updates map[string]map[string]string) {
	if updates == nil {
		return
	}

	// 1. Calculate new state
	newMap := c.PreviewSet(updates)

	// 2. Atomically swap the pointer
	c.LiveConfig.Store(newMap)
}

// Apply replaces the current live configuration with a pre-calculated map.
// Use this in conjunction with PreviewSet to avoid redundant calculations.
func (c *Config) Apply(newMap *map[string]map[string]string) {
	if newMap != nil {
		c.LiveConfig.Store(newMap)
	}
}

// PreviewSet calculates the resulting configuration map after applying updates
// but DOES NOT store it. Useful for strategies that need to push to a server
// before committing locally.
func (c *Config) PreviewSet(updates map[string]map[string]string) *map[string]map[string]string {
	if updates == nil {
		return c.LiveConfig.Load()
	}

	// 1. Create a deep copy of the current state
	newMap := make(map[string]map[string]string)
	oldPtr := c.LiveConfig.Load()
	if oldPtr != nil {
		for s, kv := range *oldPtr {
			newMap[s] = make(map[string]string)
			for k, v := range kv {
				newMap[s][k] = v
			}
		}
	}

	// 2. Apply the changes (Merge)
	for section, kv := range updates {
		if _, ok := newMap[section]; !ok {
			newMap[section] = make(map[string]string)
		}
		for k, v := range kv {
			newMap[section][k] = v
		}
	}

	return &newMap
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

// ShareConfig merges the provided configuration updates into the LiveConfig.
// It accepts either map[string]map[string]string (multi-section) 
// or map[string]string (single section, using "shared" as default).
// -----------------------------------------------------------------------------

func (c *Config) ShareConfig(payload interface{}) error {
	if payload == nil {
		return nil
	}

	var updates map[string]map[string]string

	switch p := payload.(type) {
	case map[string]map[string]string:
		updates = p
	case map[string]string:
		updates = map[string]map[string]string{
			"shared": p,
		}
	case map[string]interface{}:
		updates = make(map[string]map[string]string)
		for k, v := range p {
			if nestedMap, ok := v.(map[string]interface{}); ok {
				// Nested map becomes its own section
				section := make(map[string]string)
				for nk, nv := range nestedMap {
					section[nk] = fmt.Sprintf("%v", nv)
				}
				updates[k] = section
			} else if nestedMap, ok := v.(map[string]string); ok {
				// Pre-cast nested map
				updates[k] = nestedMap
			} else {
				// Flat key, goes to 'shared' section
				if updates["shared"] == nil {
					updates["shared"] = make(map[string]string)
				}
				updates["shared"][k] = fmt.Sprintf("%v", v)
			}
		}
	default:
		return fmt.Errorf("unsupported payload type for ShareConfig: %T", payload)
	}

	c.Set(updates)
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

// -----------------------------------------------------------------------------

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
