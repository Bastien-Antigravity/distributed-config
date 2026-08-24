package factory

import (
	"fmt"
	"strings"

	"github.com/Bastien-Antigravity/distributed-config/src/interfaces"
	"github.com/Bastien-Antigravity/distributed-config/src/strategies"
)

// NewStrategy returns the correct ConfigStrategy based on the profile name.
// Supports profiles: standalone (devel, dev, development), test, staging (stage), production (prod).
// -----------------------------------------------------------------------------

func NewStrategy(profile string) (interfaces.ConfigStrategy, error) {
	switch strings.ToLower(strings.TrimSpace(profile)) {
	case "standalone", "devel", "dev", "development":
		return &strategies.StandaloneStrategy{}, nil
	case "test":
		return &strategies.TestStrategy{}, nil
	case "staging", "stage":
		return &strategies.CloudStrategy{Profile: "staging", ReadOnly: true}, nil
	case "production", "prod":
		return &strategies.CloudStrategy{Profile: "production", ReadOnly: false}, nil
	default:
		return nil, fmt.Errorf("unknown profile: '%s'. Available profiles: standalone (devel), test, staging (stage), production (prod)", profile)
	}
}
