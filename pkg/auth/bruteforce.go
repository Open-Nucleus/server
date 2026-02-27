package auth

import (
	"sync"
	"time"
)

// BruteForceGuard tracks failed authentication attempts per device.
type BruteForceGuard struct {
	mu       sync.Mutex
	failures map[string]*failureWindow
	maxFails int
	window   time.Duration
}

type failureWindow struct {
	attempts []time.Time
}

// NewBruteForceGuard creates a new guard with the specified limits.
func NewBruteForceGuard(maxFails int, window time.Duration) *BruteForceGuard {
	return &BruteForceGuard{
		failures: make(map[string]*failureWindow),
		maxFails: maxFails,
		window:   window,
	}
}

// RecordFailure records a failed authentication attempt for a device.
func (g *BruteForceGuard) RecordFailure(deviceID string) {
	g.mu.Lock()
	defer g.mu.Unlock()

	fw, ok := g.failures[deviceID]
	if !ok {
		fw = &failureWindow{}
		g.failures[deviceID] = fw
	}
	fw.attempts = append(fw.attempts, time.Now())
}

// IsBlocked returns true if the device has exceeded the failure threshold.
func (g *BruteForceGuard) IsBlocked(deviceID string) bool {
	g.mu.Lock()
	defer g.mu.Unlock()

	fw, ok := g.failures[deviceID]
	if !ok {
		return false
	}

	cutoff := time.Now().Add(-g.window)
	g.pruneExpired(fw, cutoff)

	return len(fw.attempts) >= g.maxFails
}

// Reset clears the failure record for a device (e.g., after successful auth).
func (g *BruteForceGuard) Reset(deviceID string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	delete(g.failures, deviceID)
}

// pruneExpired removes attempts older than the cutoff. Must be called under lock.
func (g *BruteForceGuard) pruneExpired(fw *failureWindow, cutoff time.Time) {
	valid := fw.attempts[:0]
	for _, t := range fw.attempts {
		if t.After(cutoff) {
			valid = append(valid, t)
		}
	}
	fw.attempts = valid
}
