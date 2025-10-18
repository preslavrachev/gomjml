package testmode

import (
	"sync"
	"sync/atomic"
)

var (
	enabled atomic.Bool
	mu      sync.Mutex
)

// Enable enables deterministic ID generation for integration tests.
// This is a ONE-WAY operation - once enabled, it cannot be disabled in production.
// ONLY use this in test code, never in production.
func Enable() {
	enabled.Store(true)
}

// IsEnabled returns true if test mode is currently enabled.
func IsEnabled() bool {
	return enabled.Load()
}

// Controller provides exclusive access to test mode for unit testing.
// Only obtainable via LockForTesting.
type Controller struct {
	locked bool
}

// Enable sets test mode to enabled.
// Must be called while holding the Controller from LockForTesting.
func (c *Controller) Enable() {
	if !c.locked {
		panic("testmode: Enable called without holding lock")
	}
	enabled.Store(true)
}

// Disable sets test mode to disabled.
// Must be called while holding the Controller from LockForTesting.
// This is ONLY for unit tests - production code cannot disable test mode.
func (c *Controller) Disable() {
	if !c.locked {
		panic("testmode: Disable called without holding lock")
	}
	enabled.Store(false)
}

// Release releases the exclusive lock on test mode.
func (c *Controller) Release() {
	if !c.locked {
		return
	}
	c.locked = false
	mu.Unlock()
}

// LockForTesting locks test mode for exclusive unit testing access.
// Returns a Controller that must be Released when done (use defer).
// ONLY for use in unit tests that verify test mode behavior.
//
// Example:
//
//	ctrl := testmode.LockForTesting()
//	defer ctrl.Release()
//
//	ctrl.Disable()
//	// test random behavior
//
//	ctrl.Enable()
//	// test deterministic behavior
func LockForTesting() *Controller {
	mu.Lock()
	return &Controller{locked: true}
}
