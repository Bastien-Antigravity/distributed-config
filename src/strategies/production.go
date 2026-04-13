package strategies

import (
	"fmt"

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
// 2. Mem Config:
//   - Behavior: Fetched from Server (GET).
//
// 3. Persistence / Dump:
//   - On Missing File: Generates a SKELETON (Empty) file and FAILS. User must configure it.
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

	// 1. Env Load
	loader.LoadCommonFromEnv(cfg)

	// 2. Server Load
	type ConfigServerCap struct {
		IP   string `json:"ip"`
		Port string `json:"port"`
	}
	var cs ConfigServerCap
	if err := cfg.GetCapability("config_server", &cs); err == nil && cs.IP != "" {
		addr := fmt.Sprintf("%s:%s", cs.IP, cs.Port)
		client, err := network.NewClient(addr, cfg)
		if err == nil {
			s.Client = client
			serverConfig, err := client.GetConfig()
			if err == nil {
				// Merge...
				cfg.Logger.Info("Production: Loaded from Server")
				if serverConfig.Common.Name != "" {
					cfg.Common.Name = serverConfig.Common.Name
				}
			}
		}
	}

	// 3. File Load
	fullPath := loader.ResolveConfigPath("config")
	if err := loader.LoadConfigFromFileSafe(cfg, fullPath); err != nil {
		return err
	}

	// 4. Integrity Check
	if err := loader.CheckProductionIPs(cfg); err != nil {
		return err
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

func (s *ProductionStrategy) GetHandler() *network.ConfigProtoHandler {
	if s.Client != nil {
		return s.Client.Handler
	}
	return nil
}
