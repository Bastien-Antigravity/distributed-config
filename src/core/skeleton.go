package core

// NewSkeletonConfig returns an empty/zero-value Config struct.
// Used for generating "missing.yaml" or skeletons without polluting data.
// -----------------------------------------------------------------------------

func NewSkeletonConfig() *Config {
	return &Config{
		Common: CommonConfig{
			Name:  "CHANGE_ME",
			Reset: false,
		},
		// Empty capabilities or minimal comments if using a better YAML encoder
		Capabilities: Capabilities{},
	}
}
