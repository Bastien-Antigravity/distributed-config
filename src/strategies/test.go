package strategies

// =============================================================================
// ESSENTIAL PROCESS:
// Test environment strategy enforcing strict isolated loopback bindings (127.0.0.2)
// while replicating full production network synchronization and server push semantics.
//
// DATA FLOW:
// 1. Input: Test YAML files, test defaults (127.0.0.2), and mock/local config server.
// 2. Logic: Enforces test loopback integrity; syncs deltas to test server.
// 3. Output: Initialized *core.Config with active test network synchronization client.
//
// KEY PARAMETERS:
// - TestIP: Strict loopback address requirement (127.0.0.2).
// =============================================================================

import (
	"github.com/Bastien-Antigravity/distributed-config/src/core"
	"github.com/Bastien-Antigravity/distributed-config/src/network"
)

// -----------------------------------------------------------------------------

// TestStrategy: Local YAML. Uses "Test Defaults" (127.0.0.2).
// Runs like Production (Connects to Server & Syncs), requiring a local YAML file
// and enforcing 127.0.0.2 loopback IPs when CHECK_TEST_IPS=true.
// -----------------------------------------------------------------------------

type TestStrategy struct {
	Client *network.Client
	cloud  *CloudStrategy
}

// -----------------------------------------------------------------------------

func (s *TestStrategy) delegate() *CloudStrategy {
	if s.cloud == nil {
		s.cloud = &CloudStrategy{
			Profile:     "test",
			ReadOnly:    false,
			RequireFile: true,
		}
	}
	s.cloud.Client = s.Client
	return s.cloud
}

// -----------------------------------------------------------------------------

func (s *TestStrategy) Name() string { return "test" }

// -----------------------------------------------------------------------------

func (s *TestStrategy) Load(cfg *core.Config) error {
	d := s.delegate()
	err := d.Load(cfg)
	s.Client = d.Client
	return err
}

// -----------------------------------------------------------------------------

func (s *TestStrategy) Sync(cfg *core.Config) error {
	return s.delegate().Sync(cfg)
}

// -----------------------------------------------------------------------------

func (s *TestStrategy) GetHandler() *network.ConfigProtoHandler {
	return s.delegate().GetHandler()
}

// -----------------------------------------------------------------------------

func (s *TestStrategy) Set(cfg *core.Config, updates map[string]map[string]string) error {
	return s.delegate().Set(cfg, updates)
}

// -----------------------------------------------------------------------------

// SetWithOCC pushes updates with an expected version to validate OCC.
func (s *TestStrategy) SetWithOCC(cfg *core.Config, updates map[string]map[string]string, expectedVersion uint64) error {
	return s.delegate().SetWithOCC(cfg, updates, expectedVersion)
}

// -----------------------------------------------------------------------------

// DeleteKeys sends key deletion request to the server.
func (s *TestStrategy) DeleteKeys(cfg *core.Config, deletions map[string][]string) error {
	return s.delegate().DeleteKeys(cfg, deletions)
}

// -----------------------------------------------------------------------------

// DeleteKeysWithOCC sends key deletion request to the server with OCC validation.
func (s *TestStrategy) DeleteKeysWithOCC(cfg *core.Config, deletions map[string][]string, expectedVersion uint64) error {
	return s.delegate().DeleteKeysWithOCC(cfg, deletions, expectedVersion)
}

// -----------------------------------------------------------------------------

func (s *TestStrategy) Close() error {
	err := s.delegate().Close()
	s.Client = nil
	return err
}
