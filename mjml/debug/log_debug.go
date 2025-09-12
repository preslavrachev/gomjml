//go:build debug

// Package debug provides logging functionality for development and troubleshooting.
// This file contains debug build versions with actual logging implementation.
package debug

import (
	"fmt"
	"os"
	"strings"
	"time"
)

// Enabled reports whether debug logging is enabled.
// When built with the "debug" tag, this returns true so callers
// can guard expensive debug data construction.
func Enabled() bool { return true }

// DebugLog logs a debug message with component, phase, and formatted message.
// Format: [COMPONENT:phase] message: formatted_args
func DebugLog(component, phase, message string, args ...interface{}) {
	timestamp := time.Now().Format("15:04:05.000")
	formattedMessage := message
	if len(args) > 0 {
		formattedMessage = fmt.Sprintf(message, args...)
	}

	fmt.Fprintf(os.Stderr, "[%s] [%s:%s] %s\n",
		timestamp, component, phase, formattedMessage)
}

// DebugLogWithData logs a debug message with structured data.
// Format: [COMPONENT:phase] message: key1=value1 key2=value2
func DebugLogWithData(component, phase, message string, data map[string]interface{}) {
	timestamp := time.Now().Format("15:04:05.000")

	var dataStr strings.Builder
	if len(data) > 0 {
		dataStr.WriteString(": ")
		first := true
		for key, value := range data {
			if !first {
				dataStr.WriteString(" ")
			}
			dataStr.WriteString(fmt.Sprintf("%s=%v", key, value))
			first = false
		}
	}

	fmt.Fprintf(os.Stderr, "[%s] [%s:%s] %s%s\n",
		timestamp, component, phase, message, dataStr.String())
}

// DebugLogTiming logs timing information for performance analysis.
// Format: [COMPONENT:phase] message: duration=123ms
func DebugLogTiming(component, phase, message string, durationMs int64) {
	timestamp := time.Now().Format("15:04:05.000")
	fmt.Fprintf(os.Stderr, "[%s] [%s:%s] %s: duration=%dms\n",
		timestamp, component, phase, message, durationMs)
}

// DebugLogError logs error conditions during rendering.
// Format: [COMPONENT:phase] ERROR: message: error=actual_error
func DebugLogError(component, phase, message string, err error) {
	timestamp := time.Now().Format("15:04:05.000")
	fmt.Fprintf(os.Stderr, "[%s] [%s:%s] ERROR: %s: error=%v\n",
		timestamp, component, phase, message, err)
}
