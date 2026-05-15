package network

import (
	"math"
	"math/rand"
	"time"
)

// Backoff implements jittered exponential backoff.
type Backoff struct {
	BaseDelay time.Duration
	MaxDelay  time.Duration
	Factor    float64
	Jitter    float64
}

// NewBackoff creates a new Backoff instance with standard defaults.
func NewBackoff() *Backoff {
	return &Backoff{
		BaseDelay: 100 * time.Millisecond,
		MaxDelay:  5 * time.Second,
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
