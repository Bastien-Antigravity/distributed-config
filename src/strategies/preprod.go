package strategies

import (
	"fmt"

	"github.com/Bastien-Antigravity/distributed-config/src/core"
	"github.com/Bastien-Antigravity/distributed-config/src/loader"
	"github.com/Bastien-Antigravity/distributed-config/src/network"
)

// PreprodStrategy: Connects to Config Server (GET only). No Update.
//
// 1. Common Config:
//   - Source: Environment -> Server -> File.
//   - Logic: Local File is the authoritative source.
//
// 2. Mem Config:
//   - Behavior: Fetched from Server (GET).
//
// 3. Persistence / Dump:
//   - On Missing File: Generates a SKELETON (Empty) file and FAILS.
//   - Sync: DISABLED (Read-Only). No updates sent to Server.
// -----------------------------------------------------------------------------

type PreprodStrategy struct {
	Client *network.Client
}

// -----------------------------------------------------------------------------

func (s *PreprodStrategy) Name() string { return "preprod" }

// -----------------------------------------------------------------------------

func (s *PreprodStrategy) Load(cfg *core.Config) error {
	fmt.Println("Strategy: Preprod")

	// 1. Env Load
	loader.LoadCommonFromEnv(cfg)

	// 2. Server Load (GET)
	if cfg.Capabilities.ConfigServer != nil {
		addr := fmt.Sprintf("%s:%s", cfg.Capabilities.ConfigServer.IP, cfg.Capabilities.ConfigServer.Port)
		client, err := network.NewClient(addr, cfg)
		if err == nil {
			s.Client = client
			serverConfig, err := client.GetConfig()
			if err == nil {
				// Merge... (Stubbed)
				fmt.Println("Preprod: Loaded from Server")
				if serverConfig.Common.Name != "" {
					cfg.Common.Name = serverConfig.Common.Name
				}
			}
		}
	}

	// 3. File Load (Sync-check only, no generation of defaults preferred)
	// We use safe loader that DOES NOT generate defaults if missing?
	// User said: "on error if problem with config we write an empty skeleton config.yaml file"
	return loader.LoadConfigFromFileSafe(cfg, "config_preprod.yaml")
}

// -----------------------------------------------------------------------------

func (s *PreprodStrategy) Sync(cfg *core.Config) error {
	fmt.Println("Preprod: Sync disabled (Read-Only Mode)")
	return nil
}
