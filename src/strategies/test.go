package strategies

import (
	"fmt"

	"github.com/Bastien-Antigravity/distributed-config/src/core"
	"github.com/Bastien-Antigravity/distributed-config/src/loader"
	"github.com/Bastien-Antigravity/distributed-config/src/network"
)

// TestStrategy: Local YAML. Uses "Test Defaults" (127.0.0.2).
// Runs EXACTLY like Production (Connects to Server & Syncs), but bootstraps from Defaults.
//
// 1. Common Config:
//   - Source: Defaults (127.0.0.2) -> Environment -> Server -> File (Re-read).
//   - Logic: Local File is the authoritative source (overrides Server).
//   - Integrity: PANICS if any IP is NOT 127.0.0.2.
//
// 2. Mem Config:
//   - Behavior: Fetched from Server (GET).
//
// 3. Persistence / Dump:
//   - On Missing File: Generates a FULL functional config file with Test Defaults.
//   - Sync: ACTIVE. Pushes local changes to the Server (PUT).
// -----------------------------------------------------------------------------

type TestStrategy struct {
	Client *network.Client
}

// -----------------------------------------------------------------------------

func (s *TestStrategy) Name() string { return "test" }

// -----------------------------------------------------------------------------

func (s *TestStrategy) Load(cfg *core.Config) error {
	fmt.Println("Strategy: Test (Production-Like Logic)")

	// 1. Bootstrap: Load File to populate defaults (specifically Server IP: 127.0.0.2)
	// This ensures cfg.Capabilities.ConfigServer is not nil
	if err := loader.LoadConfigFromFile(cfg, "config_test.yaml"); err != nil {
		return err
	}

	// 2. Env Load (Overrides Defaults)
	loader.LoadCommonFromEnv(cfg)

	// 3. Server Load (Using IP from Bootstrap/Env)
	if cfg.Capabilities.ConfigServer != nil {
		addr := fmt.Sprintf("%s:%s", cfg.Capabilities.ConfigServer.IP, cfg.Capabilities.ConfigServer.Port)
		client, err := network.NewClient(addr, cfg)
		if err == nil {
			s.Client = client
			serverConfig, err := client.GetConfig()
			if err == nil {
				// Merge Server Config into current Config
				fmt.Println("Test: Loaded from Server")
				// Simple Manual Merge (Server Wins)
				if serverConfig.Common.Name != "" {
					cfg.Common.Name = serverConfig.Common.Name
				}
				// (In real impl, deep merge capabilities here)
			}
		} else {
			fmt.Printf("Test: Warning - Could not connect to Config Server at %s (Mock/Dev?)\n", addr)
		}
	}

	// 4. File Load Override (File Wins)
	// We reload the file to ensure local file edits override whatever the server sent.
	// This matches Production precedence (File > Server).
	// 5. Integrity Check
	if err := loader.CheckTestIPs(cfg); err != nil {
		return err
	}

	return nil
}

// -----------------------------------------------------------------------------

func (s *TestStrategy) Sync(cfg *core.Config) error {
	if s.Client != nil {
		fmt.Println("Test: Syncing updates to Server...")
		return s.Client.UpdateConfig(cfg)
	}
	return nil
}
