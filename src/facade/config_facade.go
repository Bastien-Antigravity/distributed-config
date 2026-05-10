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

func (config *Config) OnRegistryUpdate(onRegistryUpdateFn func(map[string][]string)) {
	if config.handler != nil {
		config.handler.SetOnRegistryUpdate(onRegistryUpdateFn)
	}
}

// Set overrides the core.Config.Set to trigger local callbacks and global synchronization.
// -----------------------------------------------------------------------------
func (config *Config) Set(updates map[string]map[string]string) error {
	if config.strategy != nil {
		if err := config.strategy.Set(config.Config, updates); err != nil {
			config.Logger.Error("Strategy.Set failed: %v", err)
			return err // Abort local callback on failure and return error
		}

		// Single Source of Truth Eventing:
		// If pushing to a central server, we rely on the Watch() listener to
		// catch the server's BROADCAST_SYNC to trigger observers safely.
		name := config.strategy.Name()
		if name == "production" || name == "test" {
			return nil
		}
	} else {
		// Fallback for cases where strategy is missing (should not happen in normal usage)
		config.Config.Set(updates)
	}

	// Trigger local callback for UI consistency (Standalone & Staging)
	if config.ParentOnLiveConfUpdate != nil {
		config.ParentOnLiveConfUpdate(*config.LiveConfig.Load())
	}
	return nil
}

// SetSingle is a helper for updating a single configuration value.
func (config *Config) SetSingle(section, key, value string) error {
	return config.Set(map[string]map[string]string{
		section: {key: value},
	})
}

// Sync manually triggers a refresh from the underlying strategy (e.g. Config Server).
// -----------------------------------------------------------------------------
func (config *Config) Sync() error {
	if config.strategy == nil {
		return fmt.Errorf("no strategy associated with this configuration")
	}
	return config.strategy.Sync(config.Config)
}

// ShareConfig overrides core.Config.ShareConfig to ensure updates are synchronized
// according to the current strategy (e.g., pushed to Config Server in Production).
// -----------------------------------------------------------------------------
func (config *Config) ShareConfig(payload interface{}) error {
	updates, err := core.ParseSharePayload(payload)
	if err != nil {
		return err
	}
	return config.Set(updates)
}
