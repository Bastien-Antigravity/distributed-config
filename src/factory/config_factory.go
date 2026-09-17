package factory

// =============================================================================
// ESSENTIAL PROCESS:
// Strategy factory instantiating the appropriate ConfigStrategy implementation
// based on profile aliases (standalone, test, staging, production).
//
// DATA FLOW:
// 1. Input: Profile name string (e.g. "standalone", "prod", "staging", "test").
// 2. Logic: Maps canonical names and aliases to concrete Strategy structs.
// 3. Output: ConfigStrategy interface instance or error on invalid profile name.
//
// KEY PARAMETERS:
// - profile: Environment identifier selecting runtime configuration strategy.
// =============================================================================

import (
	"fmt"
	"strings"

	"github.com/Bastien-Antigravity/distributed-config/src/interfaces"
	"github.com/Bastien-Antigravity/distributed-config/src/strategies"
)

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
