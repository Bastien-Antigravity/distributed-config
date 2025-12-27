package facade

import (
	"fmt"

	"github.com/Bastien-Antigravity/distributed-config/src/core"
	"github.com/Bastien-Antigravity/distributed-config/src/factory"
	"github.com/Bastien-Antigravity/distributed-config/src/network"
)

// Facade Config Struct
// -----------------------------------------------------------------------------

type Config struct {
	*core.Config
	handler *network.ConfigProtoHandler
	client  *network.Client

	// Callbacks
	ParentOnMemConfUpdate func(map[string]map[string]string)
}

// NewConfig initializes the configuration based on the requested Profile logic.
// profile: standalone | test | preprod | production
// -----------------------------------------------------------------------------

func NewConfig(profile string) *Config {
	cfgData := &core.Config{
		MemConfig: make(map[string]map[string]string),
	}
	configWrapper := &Config{
		Config: cfgData,
	}

	fmt.Printf("Initializing Config with Profile: %s\n", profile)

	// 1. Get Strategy
	strategy, err := factory.NewStrategy(profile)
	if err != nil {
		fmt.Printf("Critical Error: %v\n", err)
		// Return empty wrapper or panic?
		// For now, let's print huge error and return, caller will likely fail or see empty config.
		return configWrapper
	}

	// 2. Load
	if err := strategy.Load(configWrapper.Config); err != nil {
		fmt.Printf("Config Load Error: %v\n", err)
		// Strategy might have generated a skeleton or missing file.
		return configWrapper
	}

	// 3. Sync
	if err := strategy.Sync(configWrapper.Config); err != nil {
		fmt.Printf("Config Sync Warning: %v\n", err)
	}

	return configWrapper
}

// Callbacks & Helpers
// -----------------------------------------------------------------------------

func (config *Config) OnMemConfUpdate(onMemConfUpdateFn func(map[string]map[string]string)) {
	config.ParentOnMemConfUpdate = onMemConfUpdateFn
}
