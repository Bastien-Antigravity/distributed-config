package strategies

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Bastien-Antigravity/distributed-config/src/core"
	"github.com/Bastien-Antigravity/distributed-config/src/loader"
)

// StandaloneStrategy: Local YAML only. No Server.
//
// 1. Common Config:
//   - Source: Generated Defaults -> File -> Environment Overrides.
//   - Logic: Local file authoritative.
//
// 2. Mem Config:
//   - Behavior: Remains empty (No Server connection).
//
// 3. Persistence / Dump:
//   - On Missing File: Generates a FULL functional config file with defaults.
//   - Sync: Disabled.
// -----------------------------------------------------------------------------

type StandaloneStrategy struct{}

// -----------------------------------------------------------------------------

func (s *StandaloneStrategy) Name() string { return "standalone" }

// -----------------------------------------------------------------------------

func (s *StandaloneStrategy) Load(cfg *core.Config) error {
	fmt.Println("Strategy: Standalone")

	// 1. Determine File Path (Exe Name)
	exePath, err := os.Executable()
	if err != nil {
		return err
	}
	exeName := filepath.Base(exePath)
	baseName := strings.TrimSuffix(exeName, filepath.Ext(exeName))
	yamlName := baseName + ".yaml"
	fullPath := filepath.Join(filepath.Dir(exePath), yamlName)

	// 2. Load File (Generates default if missing - standard loader behavior)
	// For standalone, standard defaults (which might be the Test defaults currently) are used.
	// If we want specific "Standalone" defaults different from "Test", we'd need a specific generator.
	// For now, using standard loader.
	return loader.LoadConfigFromFile(cfg, fullPath)
}

// -----------------------------------------------------------------------------

func (s *StandaloneStrategy) Sync(cfg *core.Config) error {
	// No sync in standalone
	return nil
}
