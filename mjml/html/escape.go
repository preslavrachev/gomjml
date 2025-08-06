package html

import (
	"strings"
)

// EscapeXMLAttr escapes special characters in XML/HTML attribute values.
// It handles the standard XML entities that could break attribute parsing:
// - " -> &quot; (double quotes)
// - ' -> &#39; (single quotes, using numeric entity for better compatibility)
// - & -> &amp; (ampersands)
// - < -> &lt; (less than)
// - > -> &gt; (greater than)
//
// This function is specifically designed for use in HTML/MJML attribute values
// to prevent XML injection attacks.
func EscapeXMLAttr(s string) string {
	// Use strings.Replacer for efficient multiple replacements
	replacer := strings.NewReplacer(
		"&", "&amp;", // Must be first to avoid double-escaping
		"\"", "&quot;", // Double quotes
		"'", "&#39;", // Single quotes (numeric entity for compatibility)
		"<", "&lt;", // Less than
		">", "&gt;", // Greater than
	)
	return replacer.Replace(s)
}
