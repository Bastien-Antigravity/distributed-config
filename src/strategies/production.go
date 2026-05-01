package strategies

import (
	"fmt"
	"os"

	"github.com/Bastien-Antigravity/distributed-config/src/core"
	"github.com/Bastien-Antigravity/distributed-config/src/loader"
	"github.com/Bastien-Antigravity/distributed-config/src/network"
)

// ProductionStrategy: Connects to Config Server (GET & PUT). Full Sync.
//
// 1. Common Config:
//   - Source: Environment -> Server -> File.
//   - Logic: Local File is the authoritative source (overrides Server).
//   - Integrity: PANICS if any IP IS 127.0.0.2 (Safety Check).
//
// 2. Live Config:
//   - Behavior: Fetched from Server (GET).
//
// 3. Persistence / Dump:
//   - On Missing File: Optional. proceeds with Environment/Server data.
//   - Sync: ACTIVE. Pushes local changes to the Server (PUT).
// -----------------------------------------------------------------------------

type ProductionStrategy struct {
	Client *network.Client
}

// -----------------------------------------------------------------------------

func (s *ProductionStrategy) Name() string { return "production" }

// -----------------------------------------------------------------------------

func (s *ProductionStrategy) Load(cfg *core.Config) error {
	cfg.Logger.Info("Strategy: Production")

	// 1. Initial File Load (Gets Capabilities & config_server IP)
	fullPath := loader.ResolveConfigPath("production")
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
				cfg.Logger.Info("Production: Loaded configuration from Server")
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
		cfg.Logger.Error("Production: Required capability 'config_server' is missing!")
	}

	// 4. File Load Override (File Wins)
	// We reload the file to ensure local file edits strictly override whatever the server sent.
	if _, err := os.Stat(fullPath); err == nil {
		if err := loader.LoadYAML(fullPath, cfg); err != nil {
			return err
		}
	}

	// 5. Integrity Check
	if err := loader.CheckProductionIPs(cfg); err != nil {
		return err
	}

	// 6. Mandatory Service Validation
	if err := cfg.ValidateMandatoryServices(); err != nil {
		return fmt.Errorf("production: validation failed: %w", err)
	}

	return nil
}

// -----------------------------------------------------------------------------

func (s *ProductionStrategy) Sync(cfg *core.Config) error {
	if s.Client != nil {
		cfg.Logger.Info("Production: Syncing updates to Server...")
		return s.Client.UpdateConfig(cfg)
	}
	return nil
}

// -----------------------------------------------------------------------------

func (s *ProductionStrategy) Set(cfg *core.Config, updates map[string]map[string]string) error {
	if s.Client == nil {
		return fmt.Errorf("production: config server client not initialized")
	}

	// 1. Prepare the full state we WANT to reach (Preview)
	// We don't call cfg.Set yet to maintain server authority.
	nextState := cfg.PreviewSet(updates)
	if nextState == nil {
		return fmt.Errorf("production: failed to calculate next configuration state")
	}

	// 2. Push to server (Authoritative Check)
	cfg.Logger.Info("Production: Pushing authoritative update request to Server...")
	if err := s.Client.UpdateConfigMap(nextState); err != nil {
		return fmt.Errorf("production: server rejected update: %w", err)
	}

	// 3. Success! Now apply locally (Direct apply, no redundant calculation)
	cfg.Apply(nextState)
	return nil
}

// -----------------------------------------------------------------------------

func (s *ProductionStrategy) GetHandler() *network.ConfigProtoHandler {
	if s.Client != nil {
		return s.Client.Handler
	}
	return nil
}
