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

	// GetHandler returns the network handler if the strategy supports it.
	GetHandler() *network.ConfigProtoHandler

	// Close shuts down the strategy and any associated resources (e.g. network client).
	Close() error
}
