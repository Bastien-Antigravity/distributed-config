package strategies

import (
	"fmt"
	"os"

	"github.com/Bastien-Antigravity/distributed-config/src/core"
	"github.com/Bastien-Antigravity/distributed-config/src/loader"
	"github.com/Bastien-Antigravity/distributed-config/src/network"
)

// StagingStrategy: Connects to Config Server (GET only). No Update.
//
// 1. Common Config:
//   - Source: Environment -> Server -> File.
//   - Logic: Local File is the authoritative source.
//
// 2. Live Config:
//   - Behavior: Fetched from Server (GET).
//
// 3. Persistence / Dump:
//   - On Missing File: Optional. Proceeds with Environment/Server data.
//   - Sync: DISABLED (Read-Only). No updates sent to Server.
// -----------------------------------------------------------------------------

type StagingStrategy struct {
	Client *network.Client
}

// -----------------------------------------------------------------------------

func (s *StagingStrategy) Name() string { return "staging" }

// -----------------------------------------------------------------------------

func (s *StagingStrategy) Load(cfg *core.Config) error {
	cfg.Logger.Info("Strategy: Staging")

	// 1. Initial File Load (Gets Capabilities & config_server IP)
	fullPath := loader.ResolveConfigPath("staging")
	_ = loader.LoadConfigFromFileSafe(cfg, fullPath)

	// 2. Env Load (Overrides IP or NAME if provided dynamically)
	loader.LoadCommonFromEnv(cfg)

	// 3. Server Load (GET)
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
			if err == nil {
				cfg.Logger.Info("Staging: Loaded configuration from Server")
				// Deep Merge: Name
				if serverConfig.Common.Name != "" {
					cfg.Common.Name = serverConfig.Common.Name
				}
				// Deep Merge: Capabilities
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
		cfg.Logger.Error("Staging: Required capability 'config_server' is missing!")
	}

	// 4. File Load Override (File Wins)
	// We reload the file to ensure local file edits strictly override whatever the server sent.
	if _, err := os.Stat(fullPath); err == nil {
		if err := loader.LoadConfigFromFile(cfg, fullPath); err != nil {
			return err
		}
	}

	// 5. Mandatory Service Validation
	if err := cfg.ValidateMandatoryServices(); err != nil {
		return fmt.Errorf("staging: validation failed: %w", err)
	}

	return nil
}

// -----------------------------------------------------------------------------

func (s *StagingStrategy) Sync(cfg *core.Config) error {
	cfg.Logger.Info("Staging: Sync disabled (Read-Only Mode)")
	return nil
}

// -----------------------------------------------------------------------------

func (s *StagingStrategy) GetHandler() *network.ConfigProtoHandler {
	if s.Client != nil {
		return s.Client.Handler
	}
	return nil
}
