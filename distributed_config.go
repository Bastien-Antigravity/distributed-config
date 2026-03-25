package distributed_config

import (
	"github.com/Bastien-Antigravity/distributed-config/src/facade"
)

// -----------------------------------------------------------------------------

// Config is the main configuration object returned by the library.
// It wraps the core configuration data and provides access to helper methods.
type Config = facade.Config

// Version is the current version of the distributed-config library.
const Version = "1.4.0"

// New initializes a new configuration instance based on the specified profile.
//
// Profiles:
//   - "production": Connects to Config Server (GET & PUT), Full Synchronization.
//   - "preprod":    Connects to Config Server (GET only), No Updates.
//   - "test":       Uses Local Defaults (127.0.0.2) but mimics Production behavior.
//   - "standalone": Local YAML only, No network connection.
func New(profile string) *Config {
	return facade.NewConfig(profile)
}
