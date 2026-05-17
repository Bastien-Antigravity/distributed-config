package strategies

import (
	"fmt"
	"os"

	"github.com/Bastien-Antigravity/distributed-config/src/core"
	"github.com/Bastien-Antigravity/distributed-config/src/loader"
	"github.com/Bastien-Antigravity/distributed-config/src/network"
)

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
	type ConfigServerCap struct {
		IP   string `json:"ip"`
		Port string `json:"port"`
	}
	var cs ConfigServerCap
	if err := cfg.GetCapability("config_server", &cs); err == nil {
		addr := fmt.Sprintf("%s:%s", cs.IP, cs.Port)
		client, err := network.NewClient(addr, cfg)
		if err == nil {
			s.Client = client
			serverConfig, err := client.GetConfig()
			s.Client.Watch() // Start background hot-reloading AFTER initial sync
			if err == nil {
				cfg.Logger.Info("Cloud (%s): Loaded configuration from Server", s.Profile)
				// Deep Merge: Name
				if serverConfig.Common.Name != "" {
					cfg.Common.Name = serverConfig.Common.Name
				}
				// Deep Merge: Capabilities (Server provides the baseline)
				if cfg.Capabilities == nil {
					cfg.Capabilities = make(map[string]interface{})
				}
				for k, v := range serverConfig.Capabilities {
					if _, exists := cfg.Capabilities[k]; !exists {
						cfg.Capabilities[k] = v
					}
				}
			}
		}
	} else {
		cfg.Logger.Error("Cloud (%s): Required capability 'config_server' is missing!", s.Profile)
	}

	// 4. File Load Override (File Wins)
	// We reload the file to ensure local file edits strictly override whatever the server sent.
	if _, err := os.Stat(fullPath); err == nil {
		if err := loader.LoadYAML(fullPath, cfg); err != nil {
			return err
		}
	}

	// 5. Integrity Check (Production only)
	if s.Profile == "production" {
		if err := loader.CheckProductionIPs(cfg); err != nil {
			return err
		}
	}

	// 6. Mandatory Service Validation
	if err := cfg.ValidateMandatoryServices(); err != nil {
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
		return fmt.Errorf("cloud strategy (%s): config server client not initialized", s.Profile)
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
