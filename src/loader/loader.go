package loader

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	models "github.com/Bastien-Antigravity/distributed-config/src/core"

	"gopkg.in/yaml.v3"
)

// LoadConfigFromFile loads the YAML configuration from the specified path.
// If the file does not exist, it creates it using the values from NewDefaultConfig.
// -----------------------------------------------------------------------------

func LoadConfigFromFile(config *models.Config, filePath string) error {
	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		fmt.Printf("Config file '%s' does not exist, creating from default models...\n", filePath)

		// Ensure directory exists
		configDir := filepath.Dir(filePath)
		if _, err := os.Stat(configDir); os.IsNotExist(err) {
			if err := os.MkdirAll(configDir, 0755); err != nil {
				return fmt.Errorf("failed to create config directory: %w", err)
			}
		}

		// Generate Default Config
		defaultConfig := models.NewDefaultConfig()

		// Marshal to YAML
		data, err := yaml.Marshal(defaultConfig)
		if err != nil {
			return fmt.Errorf("failed to marshal default config: %w", err)
		}

		// Write File
		if err := os.WriteFile(filePath, data, 0644); err != nil {
			return fmt.Errorf("failed to write config file: %w", err)
		}
	}

	// Read File
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read config file '%s': %w", filePath, err)
	}

	// Parse into a Node tree to allow type control during expansion
	var root yaml.Node
	if err := yaml.Unmarshal(data, &root); err != nil {
		return fmt.Errorf("failed to parse config file '%s' into nodes: %w", filePath, err)
	}

	// Expand Environment Variables and force types
	// This forces type to !!str for all values EXCEPT true/false (!!bool)
	processNode(&root)

	// Decode Node tree into config struct
	if err := root.Decode(config); err != nil {
		return fmt.Errorf("failed to decode config file '%s': %w", filePath, err)
	}

	fmt.Printf("Config Loaded from File: %s\n", filePath)
	return nil
}

// LoadConfigFromFileSafe loads the YAML configuration.
// If the file does not exist, it generates a SKELETON (empty) file and returns an error.
// It does NOT use the Test Defaults.
// -----------------------------------------------------------------------------

func LoadConfigFromFileSafe(config *models.Config, filePath string) error {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		fmt.Printf("Config file '%s' missing (Production/Preprod). Generating Skeleton...\n", filePath)

		skeleton := models.NewSkeletonConfig()
		data, _ := yaml.Marshal(skeleton)
		_ = os.WriteFile(filePath, data, 0644)

		return fmt.Errorf("config file was missing. Generated skeleton at '%s'. Please fill it", filePath)
	}

	return LoadConfigFromFile(config, filePath)
}

// processNode recursively traverses the YAML node tree.
// It expands env variables and forces all scalars to strings, except for booleans.
// -----------------------------------------------------------------------------

func processNode(n *yaml.Node) {
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
		processNode(child)
	}
}
