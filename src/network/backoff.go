package network

import (
	"math"
	"math/rand"
	"strconv"
	"time"

	"github.com/Bastien-Antigravity/distributed-config/src/core"
)

// Backoff implements exponential backoff with jitter
type Backoff struct {
	BaseDelay time.Duration
	MaxDelay  time.Duration
	Factor    float64
	Jitter    float64
}

// NewBackoff creates a new backoff strategy using parameters from the config
func NewBackoff(cfg *core.Config) *Backoff {
	base := 100 * time.Millisecond
	if cfg != nil && cfg.Common.RetryBaseMS != "" {
		if val, err := strconv.Atoi(cfg.Common.RetryBaseMS); err == nil && val > 0 {
			base = time.Duration(val) * time.Millisecond
		}
	}

	max := 5 * time.Second
	if cfg != nil && cfg.Common.RetryMaxSec != "" {
		if val, err := strconv.Atoi(cfg.Common.RetryMaxSec); err == nil && val > 0 {
			max = time.Duration(val) * time.Second
		}
	}

	return &Backoff{
		BaseDelay: base,
		MaxDelay:  max,
		Factor:    2.0,
		Jitter:    0.1,
	}
}

// GetDelay calculates the delay for the current attempt.
func (b *Backoff) GetDelay(attempt int) time.Duration {
	if attempt < 0 {
		attempt = 0
	}

	delay := float64(b.BaseDelay) * math.Pow(b.Factor, float64(attempt))

	// Apply Jitter
	if b.Jitter > 0 {
		jitterRange := delay * b.Jitter
		randomJitter := (rand.Float64() * 2 * jitterRange) - jitterRange
		delay += randomJitter
	}

	// Cap at MaxDelay
	if delay > float64(b.MaxDelay) {
		delay = float64(b.MaxDelay)
	}

	return time.Duration(delay)
}
