package core

// =============================================================================
// ESSENTIAL PROCESS:
// Core configuration model defining the unified Config struct, atomic LiveConfig
// snapshot pointer swaps (RCU), and typed capability extraction helpers.
//
// DATA FLOW:
// 1. Input: Static YAML map structures and dynamic LiveConfig update maps.
// 2. Logic: Enforces atomic pointer swaps for lock-free reads; unmarshals capabilities.
// 3. Output: Strongly typed Capability structs, raw string values, and service addresses.
//
// KEY PARAMETERS:
// - LiveConfig: atomic.Pointer to map[string]map[string]string for lock-free dynamic reads.
// - Capabilities: Map of registered ecosystem capabilities (ports, hosts, credentials).
// =============================================================================

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/Bastien-Antigravity/distributed-config/src/utils"
)

// -----------------------------------------------------------------------------

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

	// Writer lock serializing Set, Delete, Apply, and ReplaceLive mutations
	// while keeping Get, GetCapability, and GetAddress 100% lock-free.
	writeMu sync.Mutex `yaml:"-"`

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
// Performs a thread-safe atomic swap (Read-Copy-Update) serialized by writeMu.
func (c *Config) Set(updates map[string]map[string]string) {
	if updates == nil {
		return
	}

	c.writeMu.Lock()
	defer c.writeMu.Unlock()

	// 1. Calculate new state under writeMu
	newMap := c.PreviewSet(updates)

	// 2. Atomically swap the pointer
	c.LiveConfig.Store(newMap)
}

// Apply replaces the current live configuration with a pre-calculated map.
// Serialized by writeMu so concurrent updates do not race.
func (c *Config) Apply(newMap *map[string]map[string]string) {
	if newMap == nil {
		return
	}
	c.writeMu.Lock()
	defer c.writeMu.Unlock()
	c.LiveConfig.Store(newMap)
}

// ReplaceLive atomically replaces the entire live configuration snapshot under writeMu.
func (c *Config) ReplaceLive(sections map[string]map[string]string) {
	cp := make(map[string]map[string]string, len(sections))
	for s, kv := range sections {
		secCopy := make(map[string]string, len(kv))
		for k, v := range kv {
			secCopy[k] = v
		}
		cp[s] = secCopy
	}

	c.writeMu.Lock()
	defer c.writeMu.Unlock()
	c.LiveConfig.Store(&cp)
}

// PreviewSet calculates the resulting configuration map after applying updates
// but DOES NOT store it. Useful for strategies that need to push to a server
// before committing locally. Supports "__DELETED__" sentinel values to remove keys.
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

	// 2. Apply the changes (Merge / Sentinel Delete)
	for section, kv := range updates {
		if _, ok := newMap[section]; !ok {
			newMap[section] = make(map[string]string)
		}
		for k, v := range kv {
			if v == "__DELETED__" {
				delete(newMap[section], k)
			} else {
				newMap[section][k] = v
			}
		}
		if len(newMap[section]) == 0 {
			delete(newMap, section)
		}
	}

	return &newMap
}

// -----------------------------------------------------------------------------

// Delete removes specified section keys from the live configuration.
// If keys slice is empty, removes the entire section.
func (c *Config) Delete(deletions map[string][]string) {
	if len(deletions) == 0 {
		return
	}

	c.writeMu.Lock()
	defer c.writeMu.Unlock()

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

	for section, keys := range deletions {
		if secMap, ok := newMap[section]; ok {
			if len(keys) == 0 {
				delete(newMap, section)
			} else {
				for _, k := range keys {
					delete(secMap, k)
				}
				if len(secMap) == 0 {
					delete(newMap, section)
				}
			}
		}
	}

	c.LiveConfig.Store(&newMap)
}

// -----------------------------------------------------------------------------

// GetCapability extracts a specific capability dictionary and unmarshals it into the target struct.
func (c *Config) GetCapability(key string, target interface{}) error {
	val, ok := c.Capabilities[key]
	if !ok || val == nil {
		val = make(map[string]interface{})
	}

	capMap, ok := val.(map[string]interface{})
	if !ok {
		data, err := json.Marshal(val)
		if err != nil {
			return err
		}
		return json.Unmarshal(data, target)
	}

	// Deep copy to avoid mutating the original
	mergedMap := make(map[string]interface{})
	for k, v := range capMap {
		mergedMap[k] = v
	}

	ptr := c.LiveConfig.Load()
	if ptr != nil {
		live := *ptr
		if overrides, ok := live[key]; ok {
			for k, v := range overrides {
				mergedMap[k] = v
			}
		}
		if capsSection, ok := live["capabilities"]; ok {
			prefix := key + "."
			for k, v := range capsSection {
				if strings.HasPrefix(k, prefix) {
					subKey := strings.TrimPrefix(k, prefix)
					mergedMap[subKey] = v
				}
			}
		}
	}

	if len(mergedMap) == 0 {
		return fmt.Errorf("capability '%s' is strictly required but missing from all sources", key)
	}

	// Ensure scalar identifiers and ports are represented as strings
	for _, portKey := range []string{"port", "grpc_port", "chat_id", "telegram_id"} {
		if v, exists := mergedMap[portKey]; exists {
			if _, isStr := v.(string); !isStr {
				mergedMap[portKey] = fmt.Sprintf("%v", v)
			}
		}
	}

	data, err := json.Marshal(mergedMap)
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
					if s, ok := nv.(string); ok {
						section[nk] = s
					} else {
						section[nk] = fmt.Sprintf("%v", nv)
					}
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
				if s, ok := v.(string); ok {
					updates["shared"][k] = s
				} else {
					updates["shared"][k] = fmt.Sprintf("%v", v)
				}
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
// Requires explicit 'grpc_port' (and 'grpc_ip' falling back to 'ip').
func (c *Config) GetGRPCAddress(capability string) (string, error) {
	return c.getAddr(capability, "grpc_ip", "grpc_port")
}

// -----------------------------------------------------------------------------

// GetRESTAddress returns the REST management address for a given capability.
// Requires explicit 'rest_port' (and 'rest_ip' falling back to 'ip').
func (c *Config) GetRESTAddress(capability string) (string, error) {
	return c.getAddr(capability, "rest_ip", "rest_port")
}

// -----------------------------------------------------------------------------

func isWildcardBindIP(ip string) bool {
	trimmed := strings.TrimSpace(ip)
	return trimmed == "0.0.0.0" || trimmed == "::"
}

// -----------------------------------------------------------------------------

func (c *Config) getStaticCapabilityField(capability, key string) string {
	if c.Capabilities == nil {
		return ""
	}
	capRaw, ok := c.Capabilities[capability]
	if !ok {
		return ""
	}
	switch capMap := capRaw.(type) {
	case map[string]interface{}:
		if capMap[key] != nil {
			if s, ok := capMap[key].(string); ok {
				return s
			}
			return fmt.Sprintf("%v", capMap[key])
		}
	case map[string]string:
		if s, exists := capMap[key]; exists {
			return s
		}
	}
	return ""
}

// -----------------------------------------------------------------------------

func (c *Config) getCapabilityField(capability, key string) string {
	// 1. Check LiveConfig (Overrides from CLI or Server)
	if v := c.Get(capability, key); v != "" {
		return v
	}
	if v := c.Get("capabilities", capability+"."+key); v != "" {
		return v
	}

	// 2. Check static Capabilities map
	return c.getStaticCapabilityField(capability, key)
}

// -----------------------------------------------------------------------------

func (c *Config) getListenHostField(capability, key string) string {
	if staticVal := c.getStaticCapabilityField(capability, key); isWildcardBindIP(staticVal) {
		return strings.TrimSpace(staticVal)
	}
	return c.getCapabilityField(capability, key)
}

// -----------------------------------------------------------------------------

func (c *Config) getAddr(capability, hostKey, portKey string) (string, error) {
	_, hasStatic := c.Capabilities[capability]
	hasLivePort := c.getCapabilityField(capability, portKey) != ""
	hasLiveHost := c.getCapabilityField(capability, hostKey) != "" || c.getCapabilityField(capability, "ip") != ""

	if !hasStatic && !hasLivePort && !hasLiveHost {
		if c.Capabilities == nil {
			return "", fmt.Errorf("no capabilities found and no live override for %s", capability)
		}
		return "", fmt.Errorf("capability %s not found", capability)
	}

	host := c.getListenHostField(capability, hostKey)
	if hostKey != "ip" && c.getStaticCapabilityField(capability, hostKey) == "" {
		if staticIP := c.getStaticCapabilityField(capability, "ip"); isWildcardBindIP(staticIP) {
			host = strings.TrimSpace(staticIP)
		}
	}
	if host == "" && hostKey != "ip" {
		host = c.getListenHostField(capability, "ip")
	}
	if host == "" {
		return "", fmt.Errorf("host key %s (or 'ip') missing or empty in capability %s", hostKey, capability)
	}

	port := c.getCapabilityField(capability, portKey)
	if port == "" {
		return "", fmt.Errorf("port key %s missing or empty in capability %s", portKey, capability)
	}

	return fmt.Sprintf("%s:%s", host, port), nil
}
