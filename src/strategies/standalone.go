package strategies

import (
	"fmt"
	"github.com/Bastien-Antigravity/distributed-config/src/core"
	"github.com/Bastien-Antigravity/distributed-config/src/loader"
	"github.com/Bastien-Antigravity/distributed-config/src/network"
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

	// 1. Resolve Path (Exe name or config/ fallback)
	fullPath := loader.ResolveConfigPath("standalone")

	// 2. Load File (Generates default if missing - standard loader behavior)
	return loader.LoadConfigFromFile(cfg, fullPath)
}

// -----------------------------------------------------------------------------

func (s *StandaloneStrategy) Sync(cfg *core.Config) error {
	// No sync in standalone
	return nil
}

// -----------------------------------------------------------------------------

func (s *StandaloneStrategy) GetHandler() *network.ConfigProtoHandler {
	return nil
}
