package strategies

// =============================================================================
// ESSENTIAL PROCESS:
// Standalone configuration strategy operating completely offline without
// network dependencies, loading strictly from local YAML files and env vars.
//
// DATA FLOW:
// 1. Input: Local YAML files and system environment variables.
// 2. Logic: Merges local file contents into default config; sync is disabled.
// 3. Output: Populated *core.Config instance without active network clients.
//
// KEY PARAMETERS:
// - CONFIG_PATH: Optional environment override to local YAML path.
// =============================================================================

import (
	"github.com/Bastien-Antigravity/distributed-config/src/core"
	"github.com/Bastien-Antigravity/distributed-config/src/loader"
	"github.com/Bastien-Antigravity/distributed-config/src/network"
)

// -----------------------------------------------------------------------------

// StandaloneStrategy: Local YAML only. No Server.
//
// 1. Common Config:
//   - Source: Generated Defaults -> File -> Environment Overrides.
//   - Logic: Local file authoritative.
//
// 2. Live Config:
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
	cfg.Logger.Info("Strategy: Standalone")

	// 1. Resolve Path (Exe name or config/ fallback)
	fullPath := loader.ResolveConfigPath("standalone")

	// 2. Load File (Generates default if missing - standard loader behavior)
	if err := loader.LoadConfigFromFile(cfg, fullPath); err != nil {
		return err
	}

	// 3. Env Load (Overrides NAME/RESET if provided dynamically)
	loader.LoadCommonFromEnv(cfg)
	return nil
}

// -----------------------------------------------------------------------------

func (s *StandaloneStrategy) Sync(cfg *core.Config) error {
	// No sync in standalone
	return nil
}

// -----------------------------------------------------------------------------

func (s *StandaloneStrategy) Set(cfg *core.Config, updates map[string]map[string]string) error {
	cfg.Set(updates)
	return nil
}

// -----------------------------------------------------------------------------

func (s *StandaloneStrategy) GetHandler() *network.ConfigProtoHandler {
	return nil
}

// -----------------------------------------------------------------------------

func (s *StandaloneStrategy) Close() error {
	return nil
}
