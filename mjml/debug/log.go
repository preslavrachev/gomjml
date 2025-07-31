//go:build !debug

// Package debug provides logging functionality for development and troubleshooting.
// This file contains production build versions (no-op functions) that get optimized away.
package debug

// DebugLog logs a debug message with component, phase, and formatted message.
// In production builds, this function is a no-op and gets inlined/optimized away.
func DebugLog(component, phase, message string, args ...interface{}) {
	// No-op in production build
}

// DebugLogWithData logs a debug message with structured data.
// In production builds, this function is a no-op and gets inlined/optimized away.
func DebugLogWithData(component, phase, message string, data map[string]interface{}) {
	// No-op in production build
}

// DebugLogTiming logs timing information for performance analysis.
// In production builds, this function is a no-op and gets inlined/optimized away.
func DebugLogTiming(component, phase, message string, durationMs int64) {
	// No-op in production build
}

// DebugLogError logs error conditions during rendering.
// In production builds, this function is a no-op and gets inlined/optimized away.
func DebugLogError(component, phase, message string, err error) {
	// No-op in production build
}
