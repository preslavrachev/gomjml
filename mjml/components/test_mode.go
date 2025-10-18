package components

import (
	"github.com/preslavrachev/gomjml/mjml/testmode"
)

// EnableTestMode enables deterministic ID generation for integration tests.
// This is a ONE-WAY operation - once enabled, it cannot be disabled.
// ONLY use this in test code, never in production.
func EnableTestMode() {
	testmode.Enable()
}

func isTestMode() bool {
	return testmode.IsEnabled()
}

// resetNavbarTestIndex resets the navbar test ID counter.
// ONLY for use in unit tests when switching between test modes.
func resetNavbarTestIndex() {
	navbarTestIndex.Store(0)
}

// resetCarouselTestIndex resets the carousel test ID counter.
// ONLY for use in unit tests when switching between test modes.
func resetCarouselTestIndex() {
	carouselTestIndex.Store(0)
}
