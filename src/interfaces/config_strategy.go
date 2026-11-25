package interfaces

// =============================================================================
// ESSENTIAL PROCESS:
// Defines the ConfigStrategy lifecycle contract governing profile behaviors
// across standalone, test, staging, and production environments.
//
// DATA FLOW:
// 1. Input: Core Config pointers and delta update maps.
// 2. Logic: Strategy implementation dictates synchronization, caching, and network protocols.
// 3. Output: Loaded and synchronized configuration state.
//
// KEY PARAMETERS:
// - ConfigStrategy: Primary interface implemented by profile strategy structures.
// =============================================================================

import (
	"github.com/Bastien-Antigravity/distributed-config/src/core"
	"github.com/Bastien-Antigravity/distributed-config/src/network"
)

// -----------------------------------------------------------------------------

type ConfigStrategy interface {
	// Name returns the strategy name (e.g., "production", "test")
	Name() string

	// Load retrieves the initial configuration.
	// It should handle Env loading, Defaults generation, or Remote fetching.
	Load(cfg *core.Config) error

	// Sync performs any necessary synchronization (e.g., pushing updates to server).
	Sync(cfg *core.Config) error

	// Set handles local configuration updates according to the strategy's consistency model.
	Set(cfg *core.Config, updates map[string]map[string]string) error

	// SetWithOCC pushes updates with optimistic concurrency control.
	SetWithOCC(cfg *core.Config, updates map[string]map[string]string, expectedVersion uint64) error

	// DeleteKeys removes specified section keys from live configuration and syncs according to strategy.
	DeleteKeys(cfg *core.Config, deletions map[string][]string) error

	// DeleteKeysWithOCC removes specified section keys with optimistic concurrency control.
	DeleteKeysWithOCC(cfg *core.Config, deletions map[string][]string, expectedVersion uint64) error

	// GetHandler returns the network handler if the strategy supports it.
	GetHandler() *network.ConfigProtoHandler

	// Close shuts down the strategy and any associated resources (e.g. network client).
	Close() error
}

// -----------------------------------------------------------------------------

// ConfigStrategyOCC and ConfigStrategyDeleter are retained as type aliases for backwards compatibility.
type ConfigStrategyOCC = ConfigStrategy
type ConfigStrategyDeleter = ConfigStrategy
