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

// canonicalFonts mirrors MJML 4.15.3's font detection order (Open Sans,
// Lato, Roboto, Ubuntu), with Montserrat kept after the canonical entries
// since it is not part of upstream's scanned set. Names and their lowercase
// forms are fixed at package init so lookups never re-derive them.
var canonicalFonts = [...]struct {
	name  string
	lower string
}{
	{"Open Sans", "open sans"},
	{"Lato", "lato"},
	{"Roboto", "roboto"},
	{"Ubuntu", "ubuntu"},
	{"Montserrat", "montserrat"},
}

// isCanonicalName reports whether name is one of canonicalFonts. Scanning
// the fixed 5-entry array is cheaper than maintaining a seen-set.
func isCanonicalName(name string) bool {
	for _, f := range canonicalFonts {
		if f.name == name {
			return true
		}
	}
	return false
}

// sortedExtraNames returns the GoogleFontsMapping keys that aren't part of
// canonicalFonts, in sorted order. Called only when such entries exist.
func sortedExtraNames(canonicalCount int) []string {
	extra := make([]string, 0, len(GoogleFontsMapping)-canonicalCount)
	for name := range GoogleFontsMapping {
		if !isCanonicalName(name) {
			extra = append(extra, name)
		}
	}
	sort.Strings(extra)
	return extra
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
// listing two Google fonts), the canonical MJML order decides the winner,
// followed by any non-canonical GoogleFontsMapping entries in sorted order,
// rather than Go's random map iteration order. The canonical-only case (no
// custom entries in GoogleFontsMapping) makes no ordering allocations; the
// cleaned-string copy below still allocates.
func GetGoogleFontURL(fontFamily string) string {
	cleaned := strings.ToLower(strings.Trim(fontFamily, `"' `))

	canonicalCount := 0
	for _, f := range canonicalFonts {
		url, ok := GoogleFontsMapping[f.name]
		if !ok {
			continue
		}
		canonicalCount++
		if strings.Contains(cleaned, f.lower) {
			return url
		}
	}

	if len(GoogleFontsMapping) == canonicalCount {
		return ""
	}
	for _, name := range sortedExtraNames(canonicalCount) {
		if strings.Contains(cleaned, strings.ToLower(name)) {
			return GoogleFontsMapping[name]
		}
	}
	return ""
}

// ConvertFontFamiliesToURLs converts a set of font families to their Google
// Font URLs, deduplicated and ordered by canonical MJML order first, then
// any non-canonical GoogleFontsMapping entries in sorted order - rather than
// by the order families were encountered. Deduplication is by URL, not by
// name, so two names mapped to the same URL (e.g. a mutable-mapping alias)
// still emit only one entry; empty URLs are skipped.
func ConvertFontFamiliesToURLs(fontFamilies []string) []string {
	cleanedFamilies := make([]string, len(fontFamilies))
	for i, fontFamily := range fontFamilies {
		cleanedFamilies[i] = strings.ToLower(strings.Trim(fontFamily, `"' `))
	}

	var urls []string
	seenURLs := make(map[string]bool)
	canonicalCount := 0

	for _, f := range canonicalFonts {
		url, ok := GoogleFontsMapping[f.name]
		if !ok {
			continue
		}
		canonicalCount++
		if url == "" || seenURLs[url] {
			continue
		}
		for _, cleaned := range cleanedFamilies {
			if strings.Contains(cleaned, f.lower) {
				urls = append(urls, url)
				seenURLs[url] = true
				break
			}
		}
	}

	if len(GoogleFontsMapping) == canonicalCount {
		return urls
	}
	// Duplicated rather than factored into a shared closure: a closure over
	// urls/seenURLs/cleanedFamilies would allocate, defeating the point.
	for _, name := range sortedExtraNames(canonicalCount) {
		url := GoogleFontsMapping[name]
		if url == "" || seenURLs[url] {
			continue
		}
		lower := strings.ToLower(name)
		for _, cleaned := range cleanedFamilies {
			if strings.Contains(cleaned, lower) {
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
