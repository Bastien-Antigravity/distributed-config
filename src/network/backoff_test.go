package network

import (
	"testing"
	"time"
)

func TestBackoff_GetDelay(t *testing.T) {
	b := NewBackoff()
	b.BaseDelay = 100 * time.Millisecond
	b.MaxDelay = 1000 * time.Millisecond
	b.Factor = 2.0
	b.Jitter = 0.0 // Disable jitter for predictable tests

	// Attempt 0: 100ms
	d := b.GetDelay(0)
	if d != 100*time.Millisecond {
		t.Errorf("Expected 100ms, got %v", d)
	}

	// Attempt 1: 200ms
	d = b.GetDelay(1)
	if d != 200*time.Millisecond {
		t.Errorf("Expected 200ms, got %v", d)
	}

	// Attempt 2: 400ms
	d = b.GetDelay(2)
	if d != 400*time.Millisecond {
		t.Errorf("Expected 400ms, got %v", d)
	}

	// Attempt 3: 800ms
	d = b.GetDelay(3)
	if d != 800*time.Millisecond {
		t.Errorf("Expected 800ms, got %v", d)
	}

	// Attempt 4: 1000ms (Capped at MaxDelay)
	d = b.GetDelay(4)
	if d != 1000*time.Millisecond {
		t.Errorf("Expected 1000ms, got %v", d)
	}
}

func TestBackoff_Jitter(t *testing.T) {
	b := NewBackoff()
	b.BaseDelay = 100 * time.Millisecond
	b.Jitter = 0.5 // High jitter

	// Run multiple times and ensure values vary
	delays := make(map[time.Duration]bool)
	for i := 0; i < 100; i++ {
		d := b.GetDelay(0)
		delays[d] = true
		
		// Ensure it's within range [50ms, 150ms]
		if d < 50*time.Millisecond || d > 150*time.Millisecond {
			t.Errorf("Delay %v out of jitter range [50ms, 150ms]", d)
		}
	}

	if len(delays) < 10 {
		t.Errorf("Jitter didn't produce enough variation: only %d distinct delays", len(delays))
	}
}
