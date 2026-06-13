package distributed_config

import (
	"github.com/Bastien-Antigravity/distributed-config/src/facade"
	"github.com/Bastien-Antigravity/distributed-config/src/loader"
	"github.com/Bastien-Antigravity/distributed-config/src/secret"
	"gopkg.in/yaml.v3"
)

// -----------------------------------------------------------------------------

// Config is the main configuration object returned by the library.
// It wraps the core configuration data and provides access to helper methods.
type Config = facade.Config

// New initializes a new configuration instance based on the specified profile.
//
// Profiles:
//   - "production": Connects to Config Server (GET & PUT), Full Synchronization.
//   - "staging":    Connects to Config Server (GET only), No Updates.
//   - "test":       Uses Local Defaults (127.0.0.2) but mimics Production behavior.
//   - "standalone": Local YAML only, No network connection.
func New(profile string) *Config {
	return facade.NewConfig(profile)
}

// ProcessNode is a helper that expands environment variables in a YAML node.
func ProcessNode(n *yaml.Node) {
	loader.ProcessNode(n)
}

// ResolveConfigPath returns the absolute path to the configuration file based on platform search rules.
func ResolveConfigPath(targetName string) string {
	return loader.ResolveConfigPath(targetName)
}

// Decrypt decrypts a single ENC(...) ciphertext string.
func Decrypt(ciphertext string) (string, error) {
	return secret.Decrypt(ciphertext)
}

// ProcessConfigSecrets is a helper that decrypts all ENC(...) blocks in a raw byte slice.
func ProcessConfigSecrets(content []byte) ([]byte, error) {
	return secret.ProcessConfigSecrets(content)
}
