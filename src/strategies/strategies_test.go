package strategies

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Bastien-Antigravity/distributed-config/src/core"
	"github.com/Bastien-Antigravity/distributed-config/src/utils"
)

func TestStandaloneStrategy(t *testing.T) {
	tempDir := t.TempDir()
	oldCwd, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(oldCwd)

	t.Run("TestStandalone_GeneratesFileAndLoads", func(t *testing.T) {
		cfg := &core.Config{Logger: utils.EnsureSafeLogger(nil)}
		strategy := &StandaloneStrategy{}

		err := strategy.Load(cfg)
		if err != nil {
			t.Errorf("Expected StandaloneStrategy Load to succeed and create default, got: %v", err)
		}

		if cfg.Common.Name != "common" {
			t.Errorf("Expected fallback name 'common' from default generator, got: %v", cfg.Common.Name)
		}

		// Validation should still have passed because NewDefaultConfig() provides log_server and config_server!
		if err := cfg.ValidateMandatoryServices(); err != nil {
			t.Errorf("Expected auto-generated Standalone configuration to pass basic validation, got: %v", err)
		}
	})
}

func TestProductionStrategy(t *testing.T) {
	tempDir := t.TempDir()
	oldCwd, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(oldCwd)

	t.Run("TestProduction_FailsOnMissingEnvData", func(t *testing.T) {
		// In Production, missing files are ignored (returns nil).
		// However! Since no file provides config_server and environment 
		// variables are not set, it will fail the ValidateMandatoryServices check at the end.
		cfg := &core.Config{Logger: utils.EnsureSafeLogger(nil)}
		strategy := &ProductionStrategy{}

		err := strategy.Load(cfg)
		if err == nil {
			t.Errorf("Expected ProductionStrategy to fail-fast due to missing mandatory configuration (Log/Config Server)")
		}

		if cfg.Capabilities != nil && cfg.Capabilities["log_server"] != nil {
			t.Errorf("Should not have generated any capabilities for production missing file.")
		}
	})

	t.Run("TestProduction_IPSanityFailure", func(t *testing.T) {
		// Mock out a bad production file containing a test IP.
		os.MkdirAll("config", 0755)
		badProdYaml := `
capabilities:
  log_server:
    ip: "127.0.0.2"
    port: "5000"
  config_server:
    ip: "192.168.1.100"
    port: "3306"
`
		os.WriteFile("config/production.yaml", []byte(badProdYaml), 0644)

		cfg := &core.Config{Logger: utils.EnsureSafeLogger(nil)}
		strategy := &ProductionStrategy{}

		err := strategy.Load(cfg)
		if err == nil {
			t.Errorf("Expected ProductionStrategy to fail due to CheckProductionIPs catching the 127.0.0.2 leak")
		}
	})
}

func TestStagingStrategy(t *testing.T) {
	tempDir := t.TempDir()
	oldCwd, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(oldCwd)

	t.Run("TestStaging_EnvironmentFirst_MissingDataFails", func(t *testing.T) {
		cfg := &core.Config{Logger: utils.EnsureSafeLogger(nil)}
		strategy := &StagingStrategy{}

		err := strategy.Load(cfg)
		if err == nil {
			t.Errorf("Expected StagingStrategy to fail-fast due to missing mandatory infrastructure")
		}
	})
}
