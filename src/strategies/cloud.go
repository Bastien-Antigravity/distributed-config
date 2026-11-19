package strategies

// =============================================================================
// ESSENTIAL PROCESS:
// Unified cloud configuration strategy implementation for Staging and Production,
// connecting over safe-socket to config-server for dynamic LiveConfig sync.
//
// DATA FLOW:
// 1. Input: Config server network streams, local YAML files, and environment overrides.
// 2. Logic: Merges remote and local configurations; enforces production integrity rules.
// 3. Output: Populated *core.Config with active safe-socket synchronization loops.
//
// KEY PARAMETERS:
// - Profile: "production" (full sync) or "staging" (read-only mode).
// - ReadOnly: Flag disabling writeback/sync mutations in staging environments.
// =============================================================================

import (
	"fmt"
	"os"

	"github.com/Bastien-Antigravity/distributed-config/src/core"
	"github.com/Bastien-Antigravity/distributed-config/src/loader"
	"github.com/Bastien-Antigravity/distributed-config/src/network"
)

// -----------------------------------------------------------------------------

// CloudStrategy: Unified strategy for Staging and Production.
// Connects to Config Server for remote configuration.
//
// 1. Common Config:
//   - Source: Environment -> Server -> File.
//   - Logic: Local File is the authoritative source (overrides Server).
//
// 2. Live Config:
//   - Behavior: Fetched from Server (GET).
//
// 3. Persistence / Sync:
//   - On Missing File: Optional. Proceeds with Environment/Server data.
//   - Production: Sync is ACTIVE (PUT to server). Integrity checks enabled.
//   - Staging: Sync is DISABLED (Read-Only). Integrity checks skipped.
// -----------------------------------------------------------------------------

type CloudStrategy struct {
	Client   *network.Client
	Profile  string // "production" or "staging"
	ReadOnly bool   // Disables remote sync/authoritative updates
}

// -----------------------------------------------------------------------------

func (s *CloudStrategy) Name() string { return s.Profile }

// -----------------------------------------------------------------------------

func (s *CloudStrategy) Load(cfg *core.Config) error {
	cfg.Logger.Info("Strategy: Cloud (%s)", s.Profile)

	// 1. Initial File Load (Gets Capabilities & config_server IP)
	fullPath := loader.ResolveConfigPath(s.Profile)
	_ = loader.LoadConfigFromFileSafe(cfg, fullPath)

	// 2. Env Load (Overrides IP or NAME if provided dynamically)
	loader.LoadCommonFromEnv(cfg)

	// 3. Server Load
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
				cfg.Logger.Info("Cloud (%s): Loaded configuration from Server", s.Profile)
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
			} else {
				cfg.Logger.Warning("Cloud (%s): Failed to load configuration from Server: %v", s.Profile, err)
			}
		} else {
			cfg.Logger.Warning("Cloud (%s): Could not connect to Config Server at %s: %v", s.Profile, addr, err)
		}
	} else {
		cfg.Logger.Error("Cloud (%s): Required capability 'config_server' is missing!", s.Profile)
	}

	// 4. File Load Override (File Wins)
	// We reload the file to ensure local file edits strictly override whatever the server sent.
	if _, err := os.Stat(fullPath); err == nil {
		if err := loader.LoadYAML(fullPath, cfg); err != nil {
			if s.Client != nil {
				_ = s.Client.Close()
				s.Client = nil
			}
			return err
		}
	}
	syncCapabilitiesToLiveConfig(cfg)

	// 5. Integrity Check (Production only)
	if s.Profile == "production" {
		if err := loader.CheckProductionIPs(cfg); err != nil {
			if s.Client != nil {
				_ = s.Client.Close()
				s.Client = nil
			}
			return err
		}
	}

	// 6. Mandatory Service Validation
	if err := cfg.ValidateMandatoryServices(); err != nil {
		if s.Client != nil {
			_ = s.Client.Close()
			s.Client = nil
		}
		return fmt.Errorf("cloud strategy (%s): validation failed: %w", s.Profile, err)
	}

	return nil
}

// -----------------------------------------------------------------------------

func (s *CloudStrategy) Sync(cfg *core.Config) error {
	if s.ReadOnly {
		cfg.Logger.Info("Cloud (%s): Sync disabled (Read-Only Mode)", s.Profile)
		return nil
	}

	if s.Client != nil {
		cfg.Logger.Info("Cloud (%s): Syncing updates to Server...", s.Profile)
		return s.Client.UpdateConfig(cfg)
	}
	return nil
}

// -----------------------------------------------------------------------------

func (s *CloudStrategy) Set(cfg *core.Config, updates map[string]map[string]string) error {
	if s.ReadOnly {
		// Staging/Read-only behavior: Update locally only.
		cfg.Set(updates)
		return nil
	}

	if s.Client == nil {
		// Offline fallback when config-server was not reachable at startup
		cfg.Set(updates)
		return nil
	}

	// 1. Prepare the full state we WANT to reach (Preview)
	nextState := cfg.PreviewSet(updates)
	if nextState == nil {
		return fmt.Errorf("cloud strategy (%s): failed to calculate next configuration state", s.Profile)
	}

	// 2. Push to server (Authoritative Check)
	cfg.Logger.Info("Cloud (%s): Pushing authoritative update request to Server...", s.Profile)
	if err := s.Client.UpdateConfigMap(nextState); err != nil {
		return fmt.Errorf("cloud strategy (%s): server rejected update: %w", s.Profile, err)
	}

	// 3. Success! Now apply locally
	cfg.Apply(nextState)
	return nil
}

// -----------------------------------------------------------------------------

// SetWithOCC pushes updates with an expected version to validate OCC.
func (s *CloudStrategy) SetWithOCC(cfg *core.Config, updates map[string]map[string]string, expectedVersion uint64) error {
	if s.ReadOnly {
		return fmt.Errorf("cloud strategy (%s): cannot update in read-only mode", s.Profile)
	}

	if s.Client == nil {
		cfg.Set(updates)
		return nil
	}

	cfg.Logger.Info("Cloud (%s): Pushing OCC update request (expected version: %d)...", s.Profile, expectedVersion)
	if err := s.Client.UpdateConfigMapWithOCC(updates, expectedVersion); err != nil {
		return err
	}

	cfg.Set(updates)
	return nil
}

// -----------------------------------------------------------------------------

// DeleteKeys sends key deletion request to the server.
func (s *CloudStrategy) DeleteKeys(cfg *core.Config, deletions map[string][]string) error {
	if s.ReadOnly {
		return fmt.Errorf("cloud strategy (%s): cannot delete in read-only mode", s.Profile)
	}

	if s.Client == nil {
		cfg.Delete(deletions)
		return nil
	}

	cfg.Logger.Info("Cloud (%s): Pushing key deletions to Server...", s.Profile)
	if err := s.Client.DeleteKeys(deletions); err != nil {
		return err
	}

	cfg.Delete(deletions)
	return nil
}

// -----------------------------------------------------------------------------

// DeleteKeysWithOCC sends key deletion request to the server with OCC validation.
func (s *CloudStrategy) DeleteKeysWithOCC(cfg *core.Config, deletions map[string][]string, expectedVersion uint64) error {
	if s.ReadOnly {
		return fmt.Errorf("cloud strategy (%s): cannot delete in read-only mode", s.Profile)
	}

	if s.Client == nil {
		cfg.Delete(deletions)
		return nil
	}

	cfg.Logger.Info("Cloud (%s): Pushing OCC key deletions to Server (expected version: %d)...", s.Profile, expectedVersion)
	if err := s.Client.DeleteKeysWithOCC(deletions, expectedVersion); err != nil {
		return err
	}

	cfg.Delete(deletions)
	return nil
}

// -----------------------------------------------------------------------------

func (s *CloudStrategy) GetHandler() *network.ConfigProtoHandler {
	if s.Client != nil {
		return s.Client.Handler
	}
	return nil
}

// -----------------------------------------------------------------------------

func (s *CloudStrategy) Close() error {
	if s.Client != nil {
		return s.Client.Close()
	}
	return nil
}

// -----------------------------------------------------------------------------

// syncCapabilitiesToLiveConfig extracts scalar string/number/bool fields from cfg.Capabilities
// into a map[string]map[string]string delta and merges them into cfg.LiveConfig via cfg.Set.
func syncCapabilitiesToLiveConfig(cfg *core.Config) {
	if len(cfg.Capabilities) == 0 {
		return
	}
	capUpdates := make(map[string]map[string]string, len(cfg.Capabilities))
	for capName, capVal := range cfg.Capabilities {
		sec := make(map[string]string)
		switch capMap := capVal.(type) {
		case map[string]interface{}:
			for k, v := range capMap {
				switch typed := v.(type) {
				case string:
					if typed != "" {
						sec[k] = typed
					}
				case bool, int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
					sec[k] = fmt.Sprintf("%v", typed)
				}
			}
		case map[string]string:
			for k, v := range capMap {
				if v != "" {
					sec[k] = v
				}
			}
		}
		if len(sec) > 0 {
			capUpdates[capName] = sec
		}
	}
	if len(capUpdates) > 0 {
		cfg.Set(capUpdates)
	}
}
