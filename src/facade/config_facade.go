package facade

import (
	"fmt"
	"strings"

	"github.com/Bastien-Antigravity/distributed-config/src/core"
	"github.com/Bastien-Antigravity/distributed-config/src/factory"
	"github.com/Bastien-Antigravity/distributed-config/src/interfaces"
	"github.com/Bastien-Antigravity/distributed-config/src/network"
	"github.com/Bastien-Antigravity/distributed-config/src/utils"
)

// Facade Config Struct
// -----------------------------------------------------------------------------

type Config struct {
	*core.Config
	strategy interfaces.ConfigStrategy
	handler  *network.ConfigProtoHandler

	// Callbacks
	ParentOnLiveConfUpdate func(map[string]map[string]string)
}

// NewConfig initializes the configuration based on the requested Profile logic.
// profile: standalone | test | staging | production
// -----------------------------------------------------------------------------

func NewConfig(profile string) *Config {
	cfgData := &core.Config{}
	initialMap := make(map[string]map[string]string)
	cfgData.LiveConfig.Store(&initialMap)
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

	// 4. Store Handler and Strategy for manual sync/wiring
	configWrapper.handler = strategy.GetHandler()
	configWrapper.strategy = strategy

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

// Set overrides the core.Config.Set to trigger local callbacks and global synchronization.
// -----------------------------------------------------------------------------
func (config *Config) Set(section, key, value string) {
	config.Config.Set(section, key, value)

	// Automatically trigger global synchronization
	if config.strategy != nil {
		if err := config.strategy.Sync(config.Config); err != nil {
			config.Config.Logger.Error("Auto-Sync failed after Set: %v", err)
		}
	}

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
// Sync manually triggers a refresh from the underlying strategy (e.g. Config Server).
// -----------------------------------------------------------------------------
func (config *Config) Sync() error {
	if config.strategy == nil {
		return fmt.Errorf("no strategy associated with this configuration")
	}
	return config.strategy.Sync(config.Config)
}
