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
	fmt.Println("Strategy: Production")

	// 1. Env Load
	loader.LoadCommonFromEnv(cfg)

	// 2. Server Load
	if cfg.Capabilities.ConfigServer != nil {
		addr := fmt.Sprintf("%s:%s", cfg.Capabilities.ConfigServer.IP, cfg.Capabilities.ConfigServer.Port)

		client, err := network.NewClient(addr, cfg)
		if err == nil {
			s.Client = client
			serverConfig, err := client.GetConfig()
			if err == nil {
				// Merge...
				fmt.Println("Production: Loaded from Server")
				if serverConfig.Common.Name != "" {
					cfg.Common.Name = serverConfig.Common.Name
				}
			}
		}
	}

	// 3. File Load
	// 3. File Load
	if err := loader.LoadConfigFromFileSafe(cfg, "config.yaml"); err != nil {
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
		fmt.Println("Production: Syncing updates to Server...")
		return s.Client.UpdateConfig(cfg)
	}
	return nil
}
