package loader

import (
	"fmt"
	"os"
	"path/filepath"

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

	// Expand Environment Variables
	expandedData := os.ExpandEnv(string(data))

	// Unmarshal YAML
	if err := yaml.Unmarshal([]byte(expandedData), config); err != nil {
		return fmt.Errorf("failed to parse config file '%s': %w", filePath, err)
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
