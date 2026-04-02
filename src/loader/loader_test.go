package loader

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Bastien-Antigravity/distributed-config/src/core"
)

func TestLoader(t *testing.T) {
	tempDir := t.TempDir()

	t.Run("TestYAML-Parsing", func(t *testing.T) {
		yamlContent := `
common:
  name: "test-app"
capabilities:
  timescale_db:
    db_name: "prod_db"
`
		configPath := filepath.Join(tempDir, "config.yaml")
		if err := os.WriteFile(configPath, []byte(yamlContent), 0644); err != nil {
			t.Fatal(err)
		}

		cfg := &core.Config{}
		if err := LoadConfigFromFile(cfg, configPath); err != nil {
			t.Errorf("Expected success, got error: %v", err)
		}

		if cfg.Common.Name != "test-app" {
			t.Errorf("Expected name 'test-app', got '%s'", cfg.Common.Name)
		}
		if cfg.Capabilities.TimescaleDb.DBName != "prod_db" {
			t.Errorf("Expected db_name 'prod_db', got '%s'", cfg.Capabilities.TimescaleDb.DBName)
		}
	})

	t.Run("TestSecretManagement-EnvExpansion", func(t *testing.T) {
		// Mock environment variables
		t.Setenv("DB_PASSWORD_MOCK", "TopSecret123!")
		t.Setenv("SERVICE_ID", "S123")

		yamlContent := `
common:
  name: "service-${SERVICE_ID}"
capabilities:
  timescale_db:
    password: "${DB_PASSWORD_MOCK}"
`
		configPath := filepath.Join(tempDir, "secrets.yaml")
		if err := os.WriteFile(configPath, []byte(yamlContent), 0644); err != nil {
			t.Fatal(err)
		}

		cfg := &core.Config{}
		if err := LoadConfigFromFile(cfg, configPath); err != nil {
			t.Errorf("Expected success, got error: %v", err)
		}

		if cfg.Common.Name != "service-S123" {
			t.Errorf("Expected name 'service-S123', got '%s'", cfg.Common.Name)
		}
		if cfg.Capabilities.TimescaleDb.Password != "TopSecret123!" {
			t.Errorf("Expected expansion to 'TopSecret123!', got '%s'", cfg.Capabilities.TimescaleDb.Password)
		}
	})

	t.Run("TestDefaultGeneration", func(t *testing.T) {
		configPath := filepath.Join(tempDir, "missing.yaml")
		// File does not exist yet
		
		cfg := &core.Config{}
		if err := LoadConfigFromFile(cfg, configPath); err != nil {
			t.Errorf("LoadConfigFromFile should handle missing files by creating them: %v", err)
		}

		// Check if file was created
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			t.Errorf("Expected file %s to be created using defaults", configPath)
		}
	})
}
