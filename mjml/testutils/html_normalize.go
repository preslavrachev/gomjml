package testutils

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

// NormalizeHTMLAttributes normalizes the ordering of HTML tag attributes and the ordering of CSS properties
// within style attributes in the given HTML line. It first sorts the attributes of each HTML tag, then sorts
// the CSS properties inside any style attribute, returning the normalized HTML string.
func NormalizeHTMLAttributes(line string) string {
	// First normalize HTML attribute ordering
	tagRegex := regexp.MustCompile(`<[^>]+>`)
	normalized := tagRegex.ReplaceAllStringFunc(line, sortTagAttributes)

	// Then normalize CSS properties within style attributes
	styleRegex := regexp.MustCompile(`style="([^"]*)"`)
	normalized = styleRegex.ReplaceAllStringFunc(normalized, func(match string) string {
		// Extract style content
		styleContent := strings.TrimPrefix(match, `style="`)
		styleContent = strings.TrimSuffix(styleContent, `"`)

		// Normalize CSS properties
		normalizedCSS := NormalizeCSSProperties(styleContent)

		return fmt.Sprintf(`style="%s"`, normalizedCSS)
	})

	return normalized
}

// sortTagAttributes sorts the attributes within a single HTML tag alphabetically.
// Handles both opening tags (<div class="foo" id="bar">) and self-closing tags (<img src="..." />).
func sortTagAttributes(tag string) string {
	// Remove < and > brackets
	inner := strings.Trim(tag, "<>")

	// Handle self-closing tags by preserving the trailing slash
	isSelfClosing := strings.HasSuffix(inner, "/")
	if isSelfClosing {
		inner = strings.TrimSuffix(inner, "/")
		inner = strings.TrimSpace(inner)
	}

	// Parse the tag content
	parts := parseHTMLTagContent(inner)
	if len(parts) < 2 {
		// No attributes to sort - just tag name
		if isSelfClosing {
			return fmt.Sprintf("<%s />", inner)
		}
		return tag
	}

	tagName := parts[0]
	attributes := parts[1:]

	// Sort attributes alphabetically
	sort.Strings(attributes)

	// Reconstruct the tag
	result := fmt.Sprintf("<%s %s", tagName, strings.Join(attributes, " "))
	if isSelfClosing {
		result += " />"
	} else {
		result += ">"
	}

	return result
}

// parseHTMLTagContent splits tag content into tag name and individual attributes.
// Handles quoted attribute values properly, including spaces within quotes.
func parseHTMLTagContent(content string) []string {
	var parts []string
	var current strings.Builder
	inQuotes := false
	quoteChar := byte(0)

	content = strings.TrimSpace(content)

	for i := 0; i < len(content); i++ {
		char := content[i]

		switch {
		case !inQuotes && (char == '"' || char == '\''):
			// Start of quoted string
			inQuotes = true
			quoteChar = char
			current.WriteByte(char)

		case inQuotes && char == quoteChar:
			// End of quoted string
			inQuotes = false
			quoteChar = 0
			current.WriteByte(char)

		case inQuotes:
			// Inside quoted string - preserve all characters
			current.WriteByte(char)

		case !inQuotes && char == ' ':
			// Space outside quotes - end current part
			if current.Len() > 0 {
				parts = append(parts, current.String())
				current.Reset()
			}
			// Skip consecutive spaces
			for i+1 < len(content) && content[i+1] == ' ' {
				i++
			}

		default:
			// Regular character outside quotes
			current.WriteByte(char)
		}
	}

	// Add final part if any
	if current.Len() > 0 {
		parts = append(parts, current.String())
	}

	return parts
}

// NormalizeCSSProperties takes a CSS style string and normalizes its properties.
// It removes any surrounding quotes, parses the style properties, sorts them alphabetically,
// and reconstructs the style string in a consistent, compact format (no spaces).
// If the original string had a trailing semicolon, it is preserved in the output.
// This function is useful for comparing or standardizing inline CSS styles.
//
// Example:
//
//	Input:  `font-size: 12px; color: red;`
//	Output: `color:red;font-size:12px;`
func NormalizeCSSProperties(styleValue string) string {
	// Remove style="" wrapper if present
	styleValue = strings.Trim(styleValue, `"'`)

	// Preserve trailing semicolon if present
	hasTrailingSemicolon := strings.HasSuffix(styleValue, ";")

	// Use existing robust CSS parser
	props := parseStyleProperties(styleValue)

	// Convert back to normalized string format (compact inline CSS style)
	var sortedProps []string
	for prop, value := range props {
		sortedProps = append(sortedProps, fmt.Sprintf("%s:%s", prop, value))
	}
	sort.Strings(sortedProps)

	result := strings.Join(sortedProps, ";")
	if hasTrailingSemicolon && result != "" {
		result += ";"
	}

	return result
}
