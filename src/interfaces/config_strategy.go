package interfaces

import "github.com/Bastien-Antigravity/distributed-config/src/core"

// ConfigStrategy defines the behavior for different configuration profiles.
// -----------------------------------------------------------------------------

type ConfigStrategy interface {
	// Name returns the strategy name (e.g., "production", "test")
	Name() string

	// Load retrieves the initial configuration.
	// It should handle Env loading, Defaults generation, or Remote fetching.
	Load(cfg *core.Config) error

	// Sync performs any necessary synchronization (e.g., pushing updates to server).
	Sync(cfg *core.Config) error
}
