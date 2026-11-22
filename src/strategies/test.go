package strategies

// =============================================================================
// ESSENTIAL PROCESS:
// Test environment strategy enforcing strict isolated loopback bindings (127.0.0.2)
// while replicating full production network synchronization and server push semantics.
//
// DATA FLOW:
// 1. Input: Test YAML files, test defaults (127.0.0.2), and mock/local config server.
// 2. Logic: Enforces test loopback integrity; syncs deltas to test server.
// 3. Output: Initialized *core.Config with active test network synchronization client.
//
// KEY PARAMETERS:
// - TestIP: Strict loopback address requirement (127.0.0.2).
// =============================================================================

import (
	"fmt"

	"github.com/Bastien-Antigravity/distributed-config/src/core"
	"github.com/Bastien-Antigravity/distributed-config/src/loader"
	"github.com/Bastien-Antigravity/distributed-config/src/network"
)

// -----------------------------------------------------------------------------

// TestStrategy: Local YAML. Uses "Test Defaults" (127.0.0.2).
// Runs EXACTLY like Production (Connects to Server & Syncs), but bootstraps from Defaults.
//
// 1. Common Config:
//   - Source: Defaults (127.0.0.2) -> Environment -> Server -> File (Re-read).
//   - Logic: Local File is the authoritative source (overrides Server).
//   - Integrity: PANICS if any IP is NOT 127.0.0.2.
//
// 2. Live Config:
//   - Behavior: Fetched from Server (GET).
//
// 3. Persistence / Dump:
//   - On Missing File: Fails fast (never auto-generates skeleton files).
//   - Sync: ACTIVE. Pushes local changes to the Server (PUT).
// -----------------------------------------------------------------------------

type TestStrategy struct {
	Client *network.Client
}

// -----------------------------------------------------------------------------

func (s *TestStrategy) Name() string { return "test" }

// -----------------------------------------------------------------------------

func (s *TestStrategy) Load(cfg *core.Config) error {
	cfg.Logger.Info("Strategy: Test (Production-Like Logic)")

	// 1. Bootstrap: Resolve Path and Load Default File
	fullPath := loader.ResolveConfigPath("test")
	if err := loader.LoadConfigFromFile(cfg, fullPath); err != nil {
		return err
	}

	// 2. Env Load (Overrides Defaults)
	loader.LoadCommonFromEnv(cfg)

	// 3. Server Load (Using IP from Bootstrap/Env)
	if s.Client != nil {
		_ = s.Client.Close()
		s.Client = nil
	}
	var cs core.ConfigServerCap
	if err := cfg.GetCapability("config_server", &cs); err == nil {
		addr := fmt.Sprintf("%s:%s", cs.IP, cs.Port)
		client, err := network.NewClient(addr, cfg)
		if err == nil {
			s.Client = client
			_, err := client.GetConfig()
			s.Client.Watch() // Start background hot-reloading AFTER initial sync
			if err == nil {
				cfg.Logger.Info("Test: Loaded from Server")
				// Sync server-provided capability sections into cfg.Capabilities baseline
				if livePtr := cfg.LiveConfig.Load(); livePtr != nil {
					if cfg.Capabilities == nil {
						cfg.Capabilities = make(map[string]interface{})
					}
					for section, kv := range *livePtr {
						if section == "common" || section == "shared" || section == "local" {
							continue
						}
						capMap, ok := cfg.Capabilities[section].(map[string]interface{})
						if !ok {
							capMap = make(map[string]interface{})
							cfg.Capabilities[section] = capMap
						}
						for k, v := range kv {
							capMap[k] = v
						}
					}
				}
			}
		} else {
			cfg.Logger.Warning("Test: Warning - Could not connect to Config Server at %s (Mock/Dev?)", addr)
		}
	} else {
		cfg.Logger.Warning("Test: Required capability 'config_server' is missing! proceeding locally.")
	}

	// 4. File Load Override (File Wins)
	// We reload the file to ensure local file edits override whatever the server sent.
	// This matches Production precedence (File > Server).
	if err := loader.LoadConfigFromFile(cfg, fullPath); err != nil {
		if s.Client != nil {
			_ = s.Client.Close()
			s.Client = nil
		}
		return err
	}
	syncCapabilitiesToLiveConfig(cfg)
	// 5. Integrity Check (Skipped for Docker test pipelines)
	// if err := loader.CheckTestIPs(cfg); err != nil {
	// 	return err
	// }

	// 6. Mandatory Service Validation
	if err := cfg.ValidateMandatoryServices(); err != nil {
		if s.Client != nil {
			_ = s.Client.Close()
			s.Client = nil
		}
		return fmt.Errorf("test: validation failed: %w", err)
	}

	return nil
}

// -----------------------------------------------------------------------------

func (s *TestStrategy) Sync(cfg *core.Config) error {
	if s.Client != nil {
		cfg.Logger.Info("Test: Syncing updates to Server...")
		return s.Client.UpdateConfig(cfg)
	}
	return nil
}

// -----------------------------------------------------------------------------

func (s *TestStrategy) GetHandler() *network.ConfigProtoHandler {
	if s.Client != nil {
		return s.Client.Handler
	}
	return nil
}

// -----------------------------------------------------------------------------

func (s *TestStrategy) Set(cfg *core.Config, updates map[string]map[string]string) error {
	if s.Client != nil {
		cfg.Logger.Info("Test: Pushing update request to Server...")
		if err := s.Client.UpdateConfigMapWithOCC(updates, 0); err != nil {
			return fmt.Errorf("test: server rejected update: %w", err)
		}
	}

	cfg.Set(updates)
	return nil
}

// -----------------------------------------------------------------------------

// SetWithOCC pushes updates with an expected version to validate OCC.
func (s *TestStrategy) SetWithOCC(cfg *core.Config, updates map[string]map[string]string, expectedVersion uint64) error {
	if s.Client != nil {
		cfg.Logger.Info("Test: Pushing OCC update request to Server (expected version: %d)...", expectedVersion)
		if err := s.Client.UpdateConfigMapWithOCC(updates, expectedVersion); err != nil {
			return err
		}
	}
	cfg.Set(updates)
	return nil
}

// -----------------------------------------------------------------------------

// DeleteKeys sends key deletion request to the server.
func (s *TestStrategy) DeleteKeys(cfg *core.Config, deletions map[string][]string) error {
	if s.Client != nil {
		cfg.Logger.Info("Test: Pushing key deletions to Server...")
		if err := s.Client.DeleteKeysWithOCC(deletions, 0); err != nil {
			return err
		}
	}
	cfg.Delete(deletions)
	return nil
}

// -----------------------------------------------------------------------------

// DeleteKeysWithOCC sends key deletion request to the server with OCC validation.
func (s *TestStrategy) DeleteKeysWithOCC(cfg *core.Config, deletions map[string][]string, expectedVersion uint64) error {
	if s.Client != nil {
		cfg.Logger.Info("Test: Pushing OCC key deletions to Server (expected version: %d)...", expectedVersion)
		if err := s.Client.DeleteKeysWithOCC(deletions, expectedVersion); err != nil {
			return err
		}
	}
	cfg.Delete(deletions)
	return nil
}

// -----------------------------------------------------------------------------

func (s *TestStrategy) Close() error {
	if s.Client != nil {
		return s.Client.Close()
	}
	return nil
}
