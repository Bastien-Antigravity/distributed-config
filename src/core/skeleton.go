package core

// =============================================================================
// ESSENTIAL PROCESS:
// Instantiates a minimal skeleton configuration template for new bootstrap
// environments where no pre-existing configuration file exists.
//
// DATA FLOW:
// 1. Input: None.
// 2. Logic: Constructs a *Config containing placeholder CHANGE_ME strings and empty capabilities.
// 3. Output: Skeleton *Config struct ready for serialization.
//
// KEY PARAMETERS:
// - CommonConfig: Minimal baseline metadata.
// =============================================================================

// -----------------------------------------------------------------------------

func NewSkeletonConfig() *Config {
	return &Config{
		Common: CommonConfig{
			Name:           "CHANGE_ME",
			CommonFilePath: "CHANGE_ME",
			PublicKey:      "CHANGE_ME",
			Reset:          false,
		},
		Capabilities: make(map[string]interface{}),
	}
}
