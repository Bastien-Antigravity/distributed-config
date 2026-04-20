package factory

import (
	"fmt"

	"github.com/Bastien-Antigravity/distributed-config/src/interfaces"
	"github.com/Bastien-Antigravity/distributed-config/src/strategies"
)

// NewStrategy returns the correct ConfigStrategy based on the profile name.
// -----------------------------------------------------------------------------

func NewStrategy(profile string) (interfaces.ConfigStrategy, error) {
	switch profile {
	case "standalone":
		return &strategies.StandaloneStrategy{}, nil
	case "test":
		return &strategies.TestStrategy{}, nil
	case "staging":
		return &strategies.StagingStrategy{}, nil
	case "production":
		return &strategies.ProductionStrategy{}, nil
	default:
		return nil, fmt.Errorf("unknown profile: '%s'. Available: standalone, test, staging, production", profile)
	}
}
