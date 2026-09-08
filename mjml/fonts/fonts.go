package fonts

import (
	"fmt"
	"sort"
	"strings"
)

const (
	// DefaultFontStack is the default font family used by MJML components
	DefaultFontStack = "Ubuntu, Helvetica, Arial, sans-serif"
)

// GoogleFontsMapping maps font family names to their Google Fonts URLs. It is
// the single authoritative source for both names and URLs: callers may add,
// remove, or override entries and GetGoogleFontURL/ConvertFontFamiliesToURLs
// will observe the change.
var GoogleFontsMapping = map[string]string{
	"Ubuntu":     "https://fonts.googleapis.com/css?family=Ubuntu:300,400,500,700",
	"Open Sans":  "https://fonts.googleapis.com/css?family=Open+Sans:300,400,500,700",
	"Roboto":     "https://fonts.googleapis.com/css?family=Roboto:300,400,500,700",
	"Lato":       "https://fonts.googleapis.com/css?family=Lato:300,400,500,700",
	"Montserrat": "https://fonts.googleapis.com/css?family=Montserrat:300,400,500,700",
}

// googleFontCanonicalOrder mirrors MJML 4.15.3's font detection order (Open
// Sans, Lato, Roboto, Ubuntu), with Montserrat kept after the canonical
// entries since it is not part of upstream's scanned set. Font lookups use
// this order, falling back to sorted order for any names added to
// GoogleFontsMapping beyond these, so import order stays deterministic
// without depending on Go map iteration.
var googleFontCanonicalOrder = []string{"Open Sans", "Lato", "Roboto", "Ubuntu", "Montserrat"}

// fontLookupOrder returns the names in GoogleFontsMapping to check, in
// deterministic order: the canonical MJML order first (skipping any name
// removed from the map), followed by any additional map entries in sorted
// order. This keeps GoogleFontsMapping the single source of truth for both
// which fonts exist and the order they are matched in.
func fontLookupOrder() []string {
	order := make([]string, 0, len(GoogleFontsMapping))
	seen := make(map[string]bool, len(googleFontCanonicalOrder))
	for _, name := range googleFontCanonicalOrder {
		if _, ok := GoogleFontsMapping[name]; ok {
			order = append(order, name)
			seen[name] = true
		}
	}

	var extra []string
	for name := range GoogleFontsMapping {
		if !seen[name] {
			extra = append(extra, name)
		}
	}
	sort.Strings(extra)

	return append(order, extra...)
}

// DetectDefaultFonts checks if components use default fonts that need importing
// This handles MJML's behavior of importing fonts based on component defaults, not just rendered text
func DetectDefaultFonts(hasTextComponents, hasSocialComponents, hasButtonComponents bool) []string {
	var fontsToImport []string

	// MJML automatically imports Ubuntu font when components with text content are present
	// This matches MRML's behavior - it imports fonts based on component presence, not content scanning
	if hasTextComponents || hasSocialComponents || hasButtonComponents {
		// Check if Ubuntu font should be imported (default font for most text-based components)
		if url := GetGoogleFontURL(DefaultFontStack); url != "" {
			fontsToImport = append(fontsToImport, url)
		}
	}

	return fontsToImport
}

// GetGoogleFontURL checks if a font family corresponds to a Google Font and returns its URL.
// When a font family string matches multiple mapping entries (e.g. a stack
// listing two Google fonts), fontLookupOrder's stable order decides the
// winner rather than Go's random map iteration order.
func GetGoogleFontURL(fontFamily string) string {
	cleaned := strings.ToLower(strings.Trim(fontFamily, `"' `))
	for _, name := range fontLookupOrder() {
		if strings.Contains(cleaned, strings.ToLower(name)) {
			return GoogleFontsMapping[name]
		}
	}
	return ""
}

// ConvertFontFamiliesToURLs converts a set of font families to their Google
// Font URLs, deduplicated and ordered by fontLookupOrder rather than by the
// order families were encountered. Deduplication is by URL, not by name, so
// two names mapped to the same URL (e.g. a mutable-mapping alias) still emit
// only one entry; empty URLs are skipped.
func ConvertFontFamiliesToURLs(fontFamilies []string) []string {
	var urls []string
	seenURLs := make(map[string]bool)
	for _, name := range fontLookupOrder() {
		url := GoogleFontsMapping[name]
		if url == "" || seenURLs[url] {
			continue
		}
		for _, fontFamily := range fontFamilies {
			cleaned := strings.ToLower(strings.Trim(fontFamily, `"' `))
			if strings.Contains(cleaned, strings.ToLower(name)) {
				urls = append(urls, url)
				seenURLs[url] = true
				break
			}
		}
	}
	return urls
}

// BuildFontsTags generates HTML for font imports (similar to MJML.io's buildFontsTags)
func BuildFontsTags(fontsToImport []string) string {
	if len(fontsToImport) == 0 {
		return ""
	}

	var result strings.Builder

	// Generate conditional comment for non-Outlook clients (match MRML's exact format)
	result.WriteString("<!--[if !mso]><!-->")

	// Generate <link> tags (no newlines between elements)
	for _, url := range fontsToImport {
		result.WriteString(fmt.Sprintf(`<link href="%s" rel="stylesheet" type="text/css">`, url))
	}

	// Generate <style> tag with @import statements (inline format to match MRML)
	result.WriteString("<style type=\"text/css\">")
	for _, url := range fontsToImport {
		result.WriteString(fmt.Sprintf("@import url(%s);", url))
	}
	result.WriteString("</style>")

	result.WriteString("<!--<![endif]-->")

	return result.String()
}
