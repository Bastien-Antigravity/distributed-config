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
		Capabilities: make(map[string]interface{}),
	}
}
