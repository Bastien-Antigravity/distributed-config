package loader

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Bastien-Antigravity/distributed-config/src/core"
	"github.com/Bastien-Antigravity/distributed-config/src/utils"
)

type MockTS struct {
	DBName   string `json:"db_name"`
	Password string `json:"password"`
}

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

		cfg := &core.Config{Logger: utils.EnsureSafeLogger(nil)}
		if err := LoadConfigFromFile(cfg, configPath); err != nil {
			t.Errorf("Expected success, got error: %v", err)
		}

		if cfg.Common.Name != "test-app" {
			t.Errorf("Expected name 'test-app', got '%s'", cfg.Common.Name)
		}
		
		var ts MockTS
		if err := cfg.GetCapability("timescale_db", &ts); err != nil {
			t.Fatalf("Failed to get capability: %v", err)
		}
		if ts.DBName != "prod_db" {
			t.Errorf("Expected db_name 'prod_db', got '%s'", ts.DBName)
		}
	})

	t.Run("TestSecretManagement-EnvExpansion", func(t *testing.T) {
		// Mock environment variables
		t.Setenv("TS_PASSWORD_MOCK", "TopSecret123!")
		t.Setenv("TS_SERVICE_ID", "S123")

		yamlContent := `
common:
  name: "service-${TS_SERVICE_ID}"
capabilities:
  timescale_db:
    password: "${TS_PASSWORD_MOCK}"
`
		configPath := filepath.Join(tempDir, "secrets.yaml")
		if err := os.WriteFile(configPath, []byte(yamlContent), 0644); err != nil {
			t.Fatal(err)
		}

		cfg := &core.Config{Logger: utils.EnsureSafeLogger(nil)}
		if err := LoadConfigFromFile(cfg, configPath); err != nil {
			t.Errorf("Expected success, got error: %v", err)
		}

		if cfg.Common.Name != "service-S123" {
			t.Errorf("Expected name 'service-S123', got '%s'", cfg.Common.Name)
		}
		
		var ts MockTS
		if err := cfg.GetCapability("timescale_db", &ts); err != nil {
			t.Fatalf("Failed to get capability: %v", err)
		}
		if ts.Password != "TopSecret123!" {
			t.Errorf("Expected expansion to 'TopSecret123!', got '%s'", ts.Password)
		}
	})

	t.Run("TestEnvExpansion-WithDefaults", func(t *testing.T) {
		t.Setenv("VAR_EXISTS", "RealValue")
		t.Setenv("VAR_EMPTY", "")
		// VAR_UNSET is not set

		yamlContent := `
common:
  name: "${VAR_EXISTS:DefaultValue}"
capabilities:
  timescale_db:
    db_name: "${VAR_UNSET:DefaultDB}"
    password: "${VAR_EMPTY:DefaultPass}"
`
		configPath := filepath.Join(tempDir, "defaults.yaml")
		if err := os.WriteFile(configPath, []byte(yamlContent), 0644); err != nil {
			t.Fatal(err)
		}

		cfg := &core.Config{Logger: utils.EnsureSafeLogger(nil)}
		if err := LoadConfigFromFile(cfg, configPath); err != nil {
			t.Errorf("Expected success, got error: %v", err)
		}

		if cfg.Common.Name != "RealValue" {
			t.Errorf("Expected 'RealValue', got '%s'", cfg.Common.Name)
		}
		
		var ts MockTS
		if err := cfg.GetCapability("timescale_db", &ts); err != nil {
			t.Fatalf("Failed to get capability: %v", err)
		}
		if ts.DBName != "DefaultDB" {
			t.Errorf("Expected 'DefaultDB', got '%s'", ts.DBName)
		}
		if ts.Password != "DefaultPass" {
			t.Errorf("Expected 'DefaultPass', got '%s'", ts.Password)
		}
	})

	t.Run("TestEnvExpansion-ForcesString", func(t *testing.T) {
		t.Setenv("PORT_VAR", "8080")
		yamlContent := `
capabilities:
  server:
    port: ${PORT_VAR:9090}
`
		configPath := filepath.Join(tempDir, "force_str.yaml")
		if err := os.WriteFile(configPath, []byte(yamlContent), 0644); err != nil {
			t.Fatal(err)
		}

		cfg := &core.Config{Logger: utils.EnsureSafeLogger(nil)}
		if err := LoadConfigFromFile(cfg, configPath); err != nil {
			t.Fatal(err)
		}

		server, ok := cfg.Capabilities["server"].(map[string]interface{})
		if !ok {
			t.Fatal("expected server map")
		}
		port, ok := server["port"].(string)
		if !ok {
			t.Errorf("expected port to be string, got %T (%v)", server["port"], server["port"])
		}
		if port != "8080" {
			t.Errorf("expected 8080, got %v", port)
		}
	})

	t.Run("TestTypes-ConditionalForcing", func(t *testing.T) {
		t.Setenv("ENABLE_VAR", "true")
		yamlContent := `
capabilities:
  server:
    port: 8080
    enabled: ${ENABLE_VAR:false}
    debug: true
    label: true-app
`
		configPath := filepath.Join(tempDir, "conditional.yaml")
		if err := os.WriteFile(configPath, []byte(yamlContent), 0644); err != nil {
			t.Fatal(err)
		}

		cfg := &core.Config{Logger: utils.EnsureSafeLogger(nil)}
		if err := LoadConfigFromFile(cfg, configPath); err != nil {
			t.Fatal(err)
		}

		server := cfg.Capabilities["server"].(map[string]interface{})
		
		// 8080 should be string
		if _, ok := server["port"].(string); !ok {
			t.Errorf("expected port to be string, got %T", server["port"])
		}

		// enabled (from env var true) should be bool
		if _, ok := server["enabled"].(bool); !ok {
			t.Errorf("expected enabled to be bool, got %T", server["enabled"])
		}

		// debug (true) should be bool
		if _, ok := server["debug"].(bool); !ok {
			t.Errorf("expected debug to be bool, got %T", server["debug"])
		}

		// label (true-app) should be string
		if _, ok := server["label"].(string); !ok {
			t.Errorf("expected label to be string, got %T", server["label"])
		}
	})

	t.Run("TestDefaultGeneration", func(t *testing.T) {
		configPath := filepath.Join(tempDir, "missing.yaml")
		cfg := &core.Config{Logger: utils.EnsureSafeLogger(nil)}
		if err := LoadConfigFromFile(cfg, configPath); err != nil {
			t.Errorf("LoadConfigFromFile should handle missing files: %v", err)
		}
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			t.Errorf("Expected file %s to be created", configPath)
		}
	})

	t.Run("TestEnvironmentFirstBoot-SafeLoader", func(t *testing.T) {
		configPath := filepath.Join(tempDir, "missing_safe.yaml")
		cfg := &core.Config{Logger: utils.EnsureSafeLogger(nil)}
		
		// LoadConfigFromFileSafe should NOT create a file and should NOT return an error
		err := LoadConfigFromFileSafe(cfg, configPath)
		if err != nil {
			t.Errorf("Expected nil error for missing file in safe mode, got: %v", err)
		}

		if _, err := os.Stat(configPath); !os.IsNotExist(err) {
			t.Errorf("Expected file %s to NOT be created in safe mode", configPath)
		}
	})
}
