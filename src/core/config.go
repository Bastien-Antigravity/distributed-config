package core

import (
	"encoding/json"
	"fmt"
	"strings"
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

	// Network Settings
	PublicIP    string `yaml:"public_ip" json:"public_ip"`         // Binding interface, defaults to 127.0.0.1
	RetryBaseMS string `yaml:"retry_base_ms" json:"retry_base_ms"` // Base delay for backoff, defaults to "100"
	RetryMaxSec string `yaml:"retry_max_sec" json:"retry_max_sec"` // Max delay for backoff, defaults to "5"
}

// Config Data Struct (Pure Data)
// -----------------------------------------------------------------------------

type Config struct {
	// Distributed system name
	Common CommonConfig `yaml:"common" json:"common"`

	// Data storage for Config params
	// main config
	Capabilities map[string]interface{}                       `yaml:"capabilities" json:"capabilities"`
	LiveConfig   atomic.Pointer[map[string]map[string]string] `yaml:"-"`

	// Internal state
	Logger utils.Logger `yaml:"-"`
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
// It also merges overrides from LiveConfig if they exist.
func (c *Config) GetCapability(key string, target interface{}) error {
	val, ok := c.Capabilities[key]
	if !ok || val == nil {
		// Even if not in static Capabilities, it might be in LiveConfig
		val = make(map[string]interface{})
	}

	// 1. Convert static/base capability to map for merging
	capMap, ok := val.(map[string]interface{})
	if !ok {
		// If it's not a map, we can't easily merge, but we still try to marshal it
		data, err := json.Marshal(val)
		if err != nil {
			return err
		}
		return json.Unmarshal(data, target)
	}

	// 2. Check for LiveConfig overrides
	// We check both the direct section (e.g. "log_server") AND the "capabilities" section (legacy/alternative)
	ptr := c.LiveConfig.Load()
	if ptr != nil {
		live := *ptr
		// Direct section override (Priority 1)
		if overrides, ok := live[key]; ok {
			for k, v := range overrides {
				capMap[k] = v
			}
		}
		// "capabilities" section override (e.g. key "log_server.ip") (Priority 2 - Legacy)
		if capsSection, ok := live["capabilities"]; ok {
			prefix := key + "."
			for k, v := range capsSection {
				if strings.HasPrefix(k, prefix) {
					subKey := strings.TrimPrefix(k, prefix)
					capMap[subKey] = v
				}
			}
		}
	}

	if len(capMap) == 0 {
		return fmt.Errorf("capability '%s' is strictly required but missing from all sources", key)
	}

	data, err := json.Marshal(capMap)
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
	updates, err := ParseSharePayload(payload)
	if err != nil {
		return err
	}
	c.Set(updates)
	return nil
}

// ParseSharePayload converts a generic payload into a structured configuration update map.
// Supports map[string]map[string]string, map[string]string, and map[string]interface{}.
func ParseSharePayload(payload interface{}) (map[string]map[string]string, error) {
	if payload == nil {
		return nil, nil
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
		return nil, fmt.Errorf("unsupported payload type for ShareConfig: %T", payload)
	}

	return updates, nil
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
	// 1. Check LiveConfig (Overrides from CLI or Server)
	// Try direct section first (e.g. section "log_server" key "ip")
	host := c.Get(capability, hostKey)
	port := c.Get(capability, portKey)

	// Fallback to "capabilities" section (legacy/alternative)
	if host == "" {
		host = c.Get("capabilities", capability+"."+hostKey)
	}
	if port == "" {
		port = c.Get("capabilities", capability+"."+portKey)
	}

	// 2. Fallback to static Capabilities map if missing from LiveConfig
	if host == "" || port == "" {
		if c.Capabilities == nil {
			return "", fmt.Errorf("no capabilities found and no live override for %s", capability)
		}
		capRaw, ok := c.Capabilities[capability]
		if !ok {
			if host == "" {
				return "", fmt.Errorf("capability %s not found", capability)
			}
		} else {
			cap, ok := capRaw.(map[string]interface{})
			if ok {
				if host == "" {
					h, _ := cap[hostKey].(string)
					host = h
				}
				if port == "" {
					p, _ := cap[portKey].(string)
					port = p
				}
			}
		}
	}

	if host == "" {
		return "", fmt.Errorf("host key %s missing or empty in capability %s", hostKey, capability)
	}
	if port == "" {
		return "", fmt.Errorf("port key %s missing or empty in capability %s", portKey, capability)
	}

	return fmt.Sprintf("%s:%s", host, port), nil
}
