package facade

import (
	"fmt"
	"strings"

	"github.com/Bastien-Antigravity/distributed-config/src/core"
	"github.com/Bastien-Antigravity/distributed-config/src/factory"
	"github.com/Bastien-Antigravity/distributed-config/src/network"
	"github.com/Bastien-Antigravity/distributed-config/src/utils"
)

// Facade Config Struct
// -----------------------------------------------------------------------------

type Config struct {
	*core.Config
	handler *network.ConfigProtoHandler

	// Callbacks
	ParentOnLiveConfUpdate func(map[string]map[string]string)
}

// NewConfig initializes the configuration based on the requested Profile logic.
// profile: standalone | test | staging | production
// -----------------------------------------------------------------------------

func NewConfig(profile string) *Config {
	cfgData := &core.Config{
		LiveConfig: make(map[string]map[string]string),
	}
	cfgData.Logger = utils.EnsureSafeLogger(nil) // Default to no-op if not explicitly set later

	configWrapper := &Config{
		Config: cfgData,
	}

	cfgData.Logger.Info("Initializing Config with Profile: %s", profile)

	// 1. Get Strategy
	profile = strings.TrimSpace(profile)
	strategy, err := factory.NewStrategy(profile)
	if err != nil {
		fmt.Printf("Critical Error: %v\n", err)
		return configWrapper
	}

	if strategy == nil {
		panic(fmt.Sprintf("distributed-config: Factory returned nil strategy for profile '%s'", profile))
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

	// 4. Store Handler for Callback wiring
	configWrapper.handler = strategy.GetHandler()

	return configWrapper
}

// Callbacks & Helpers
// -----------------------------------------------------------------------------

func (config *Config) OnLiveConfUpdate(onLiveConfUpdateFn func(map[string]map[string]string)) {
	config.ParentOnLiveConfUpdate = onLiveConfUpdateFn
	if config.handler != nil {
		config.handler.SetOnLiveConfUpdate(onLiveConfUpdateFn)
	}
}

// Set overrides the core.Config.Set to trigger local callbacks.
// -----------------------------------------------------------------------------
func (config *Config) Set(section, key, value string) {
	config.Config.Set(section, key, value)
	if config.ParentOnLiveConfUpdate != nil {
		// Create a single-entry update map for the callback
		update := map[string]map[string]string{
			section: {
				key: value,
			},
		}
		config.ParentOnLiveConfUpdate(update)
	}
}
