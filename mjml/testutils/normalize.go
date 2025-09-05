package testutils

import (
	"regexp"
	"sort"
	"strings"
)

var (
	whitespaceBetweenTagsRe = regexp.MustCompile(`>\s+<`)
	multipleWhitespaceRe    = regexp.MustCompile(`\s+`)
)

// NormalizeForComparison normalizes HTML content to eliminate whitespace and formatting
// differences that don't affect the semantic content. This is useful for comparing
// generated HTML against reference files where minor formatting differences should be ignored.
//
// The function performs the following normalizations:
// 1. Removes whitespace between HTML tags (e.g., "> <" becomes "><")
// 2. Normalizes text content within tags by trimming whitespace
// 3. Converts multiple whitespace characters to single spaces
// 4. Sorts CSS properties within style attributes for consistent ordering
// 5. Normalizes spacing around colons in CSS properties
func NormalizeForComparison(html string) string {
	// Remove whitespace between tags
	normalized := whitespaceBetweenTagsRe.ReplaceAllString(html, "><")

	// Normalize text content within tags (preserve &nbsp; and other entities)
	normalized = regexp.MustCompile(`>(\s+)([^<]*?)(\s+)<`).ReplaceAllStringFunc(normalized, func(match string) string {
		// Extract content between > and <
		content := match[1 : len(match)-1] // Remove > and <
		content = strings.TrimSpace(content)
		if content == "" {
			return "><"
		}
		return ">" + content + "<"
	})

	// Normalize multiple whitespace to single spaces
	normalized = multipleWhitespaceRe.ReplaceAllString(normalized, " ")

	// Handle double quoted style attributes
	normalized = regexp.MustCompile(`style="([^"]*)"`).ReplaceAllStringFunc(normalized, func(match string) string {
		styleContent := match[7 : len(match)-1] // Remove style=" and "
		if styleContent == "" {
			return `style=""`
		}
		return `style="` + normalizeCSSProperties(styleContent) + `"`
	})

	// Handle single quoted style attributes
	normalized = regexp.MustCompile(`style='([^']*)'`).ReplaceAllStringFunc(normalized, func(match string) string {
		styleContent := match[7 : len(match)-1] // Remove style=' and '
		if styleContent == "" {
			return `style=''`
		}
		return `style='` + normalizeCSSProperties(styleContent) + `'`
	})

	return strings.TrimSpace(normalized)
}

// Helper function to normalize CSS properties
func normalizeCSSProperties(cssContent string) string {
	// Parse and sort CSS properties
	properties := strings.Split(cssContent, ";")
	var sortedProps []string
	for _, prop := range properties {
		if prop = strings.TrimSpace(prop); prop != "" {
			// Normalize spaces around colons
			if colonIdx := strings.Index(prop, ":"); colonIdx != -1 {
				key := strings.TrimSpace(prop[:colonIdx])
				value := strings.TrimSpace(prop[colonIdx+1:])
				sortedProps = append(sortedProps, key+":"+value)
			}
		}
	}
	sort.Strings(sortedProps)
	return strings.Join(sortedProps, ";")
}
