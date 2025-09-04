package testutils

import (
	"strings"
)

// StylesEqual compares two CSS style strings semantically, ignoring property order
// and whitespace differences. It parses both styles into property maps and compares
// them for equality.
//
// Returns true if both styles represent the same set of CSS properties and values.
//
// Examples:
//
//	StylesEqual("color: red; font-size: 12px;", "font-size: 12px; color: red;") // true
//	StylesEqual("padding: 10px", "padding:10px") // true
//	StylesEqual("", "") // true
//	StylesEqual("color: red", "color: blue") // false
func StylesEqual(style1, style2 string) bool {
	if style1 == style2 {
		return true
	}

	props1 := parseStyleProperties(style1)
	props2 := parseStyleProperties(style2)

	if len(props1) != len(props2) {
		return false
	}

	for prop, value1 := range props1 {
		if value2, exists := props2[prop]; !exists || value1 != value2 {
			return false
		}
	}

	return true
}

// parseStyleProperties parses a CSS style string into a map of property-value pairs.
// Handles standard CSS declaration format: "property: value; property2: value2;"
func parseStyleProperties(style string) map[string]string {
	props := make(map[string]string)
	if style == "" {
		return props
	}

	// Split by semicolon and parse each property
	declarations := strings.Split(style, ";")
	for _, decl := range declarations {
		decl = strings.TrimSpace(decl)
		if decl == "" {
			continue
		}

		// Split by colon to separate property and value
		parts := strings.SplitN(decl, ":", 2)
		if len(parts) != 2 {
			continue
		}

		prop := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		if prop != "" && value != "" {
			props[prop] = value
		}
	}

	return props
}
