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
				// Merge... (Stubbed)
				fmt.Println("Preprod: Loaded from Server")
				if serverConfig.Common.Name != "" {
					cfg.Common.Name = serverConfig.Common.Name
				}
			}
		}
	}

	// 3. File Load
	fullPath := loader.ResolveConfigPath("config_preprod")
	return loader.LoadConfigFromFileSafe(cfg, fullPath)
}

// -----------------------------------------------------------------------------

func (s *PreprodStrategy) Sync(cfg *core.Config) error {
	fmt.Println("Preprod: Sync disabled (Read-Only Mode)")
	return nil
}

// -----------------------------------------------------------------------------

func (s *PreprodStrategy) GetHandler() *network.ConfigProtoHandler {
	if s.Client != nil {
		return s.Client.Handler
	}
	return nil
}
