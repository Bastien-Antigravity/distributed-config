package loader

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	models "github.com/Bastien-Antigravity/distributed-config/src/core"

	"gopkg.in/yaml.v3"
)


// -----------------------------------------------------------------------------

// EnsureFileExists checks if a file exists. If it doesn't, it creates it using the provided payload.
func EnsureFileExists(filePath string, payload interface{}) error {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		configDir := filepath.Dir(filePath)
		if _, err := os.Stat(configDir); os.IsNotExist(err) {
			if err := os.MkdirAll(configDir, 0755); err != nil {
				return fmt.Errorf("failed to create config directory: %w", err)
			}
		}

		data, err := yaml.Marshal(payload)
		if err != nil {
			return fmt.Errorf("failed to marshal default config: %w", err)
		}

		if err := os.WriteFile(filePath, data, 0644); err != nil {
			return fmt.Errorf("failed to write config file: %w", err)
		}
	}
	return nil
}

// -----------------------------------------------------------------------------

// LoadConfigFromFile loads the core configuration.
// Behavior: If the file is missing, it automatically creates it using standard
// ecosystem Defaults (from NewDefaultConfig). This is intended for Standalone
// and Test profiles to allow instant "Zero-Config" local bootstrap.
func LoadConfigFromFile(config *models.Config, filePath string) error {
	if err := EnsureFileExists(filePath, models.NewDefaultConfig()); err != nil {
		return err
	}
	err := LoadYAML(filePath, config)
	if err == nil {
		config.Logger.Info("Config Loaded from File: %s", filePath)
	}
	return err
}

// -----------------------------------------------------------------------------

// LoadConfigFromFileSafe loads the YAML configuration with safety checks.
// Behavior: If the file is missing, it returns nil (no error). It DOES NOT
// generate skeletons. This allows for Environment-First bootstrapping where
// config tags like CF_IP/CF_PORT are provided via ENV variables.
func LoadConfigFromFileSafe(config *models.Config, filePath string) error {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		config.Logger.Info("Config file missing at '%s'. Proceeding with Environment/Server discovery...", filePath)
		return nil
	}

	return LoadYAML(filePath, config)
}


// -----------------------------------------------------------------------------

// ProcessNode recursively traverses the YAML node tree.
// It expands env variables and forces all scalars to strings, except for booleans.
func ProcessNode(n *yaml.Node) {
	if n.Kind == yaml.ScalarNode {
		// Expand Environment Variables
		if strings.Contains(n.Value, "${") {
			n.Value = os.Expand(n.Value, func(s string) string {
				parts := strings.SplitN(s, ":", 2)
				val := os.Getenv(parts[0])
				if val == "" && len(parts) > 1 {
					val = parts[1]
				}
				return strings.Trim(val, "\"")
			})
		}

		// Force types: Booleans remain bool, everything else becomes string
		lowerVal := strings.ToLower(n.Value)
		if lowerVal == "true" || lowerVal == "false" {
			n.Tag = "!!bool"
			n.Style = 0 // Plain style for booleans
		} else {
			n.Tag = "!!str"
			n.Style = yaml.DoubleQuotedStyle
		}
	}
	for _, child := range n.Content {
		ProcessNode(child)
	}
}

// -----------------------------------------------------------------------------

// LoadYAML is the generic low-level utility to load and parse any YAML file.
// It applies Environment Variable Expansion (${VAR:default}) and enforces
// consistent string-typing (except for booleans) natively on the node tree.
func LoadYAML(filePath string, target interface{}) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file '%s': %w", filePath, err)
	}

	var root yaml.Node
	if err := yaml.Unmarshal(data, &root); err != nil {
		return fmt.Errorf("failed to parse yaml file '%s' into nodes: %w", filePath, err)
	}

	ProcessNode(&root)

	if err := root.Decode(target); err != nil {
		return fmt.Errorf("failed to decode yaml file '%s': %w", filePath, err)
	}

	return nil
}
