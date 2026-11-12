package fonts

import (
	"fmt"
	"slices"
	"sort"
	"strings"
)

const (
	// DefaultFontStack is the default font family used by MJML components
	DefaultFontStack = "Ubuntu, Helvetica, Arial, sans-serif"
)

// GoogleFontsMapping maps font family names to their Google Fonts URLs. It is
// the single authoritative source for both names and URLs: callers may add,
// remove, or override entries and GetGoogleFontURL/ConvertFontFamiliesToURLs/
// ImportURLs will observe the change. As ImportURLs imports the mapping's
// entries before the mj-font names that aren't in it, an added entry is
// imported ahead of them.
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
	if name, ok := matchingName(fontFamily); ok {
		return GoogleFontsMapping[name]
	}
	return ""
}

// matchingName returns the GoogleFontsMapping name GetGoogleFontURL resolves a
// font family to: the first entry, in detection order, whose name the family
// contains, ignoring case.
func matchingName(fontFamily string) (string, bool) {
	cleaned := strings.ToLower(strings.Trim(fontFamily, `"' `))

	canonicalCount := 0
	for _, f := range canonicalFonts {
		if _, ok := GoogleFontsMapping[f.name]; !ok {
			continue
		}
		canonicalCount++
		if strings.Contains(cleaned, f.lower) {
			return f.name, true
		}
	}

	if len(GoogleFontsMapping) == canonicalCount {
		return "", false
	}
	for _, name := range sortedExtraNames(canonicalCount) {
		if strings.Contains(cleaned, strings.ToLower(name)) {
			return name, true
		}
	}
	return "", false
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

// Declaration is a font declared by an mj-font tag.
type Declaration struct {
	Name string
	Href string
}

// ImportURLs returns the stylesheet URLs a template imports, ordered and resolved as MJML's
// font registry: the GoogleFontsMapping entries in detection order (canonical names, then any
// other entries in sorted order), followed by the declared names that aren't in it, in the
// order they are first declared. A declaration overwrites the URL of the entry with exactly its
// name in place ("lato" doesn't overwrite "Lato"), and a name declared several times keeps its
// last href. Declarations with an empty href are ignored.
//
// A GoogleFontsMapping entry is imported when one of fontFamilies contains its name, ignoring
// case, or, when importDefault is set, when it is the entry GetGoogleFontURL resolves
// DefaultFontStack to. A declared name is imported whether or not fontFamilies uses it: MJML
// imports it only when used, but fontFamilies doesn't see font-family set in raw HTML
// content, which MJML's scan of the rendered output does. URLs are deduplicated and empty
// URLs skipped.
//
// Entries are matched by name, so overwriting one leaves other names mapped to the same URL
// importing it.
func ImportURLs(fontFamilies []string, declarations []Declaration, importDefault bool) []string {
	// lastHrefs maps each declared name to its last href; newNames lists the declared names
	// that aren't in GoogleFontsMapping, in the order they are first declared.
	var lastHrefs map[string]string
	var newNames []string
	for _, d := range declarations {
		if d.Href == "" {
			continue
		}
		if lastHrefs == nil {
			lastHrefs = make(map[string]string, len(declarations))
		}
		if _, seen := lastHrefs[d.Name]; !seen {
			if _, inMapping := GoogleFontsMapping[d.Name]; !inMapping {
				newNames = append(newNames, d.Name)
			}
		}
		lastHrefs[d.Name] = d.Href
	}

	cleanedFamilies := make([]string, len(fontFamilies))
	for i, fontFamily := range fontFamilies {
		cleanedFamilies[i] = strings.ToLower(strings.Trim(fontFamily, `"' `))
	}
	defaultName := ""
	if importDefault {
		defaultName, _ = matchingName(DefaultFontStack)
	}

	var urls []string
	canonicalCount := 0
	for _, f := range canonicalFonts {
		if _, ok := GoogleFontsMapping[f.name]; !ok {
			continue
		}
		canonicalCount++
		urls = appendEntryImport(urls, f.name, f.lower, cleanedFamilies, lastHrefs, defaultName)
	}
	if len(GoogleFontsMapping) != canonicalCount {
		for _, name := range sortedExtraNames(canonicalCount) {
			urls = appendEntryImport(urls, name, strings.ToLower(name), cleanedFamilies, lastHrefs, defaultName)
		}
	}
	for _, name := range newNames {
		urls = appendUniqueURL(urls, lastHrefs[name])
	}
	return urls
}

// appendEntryImport appends the import of the GoogleFontsMapping entry name, if any: the href
// declared for it, whether or not it is used, else its URL when one of cleanedFamilies
// contains lower or name is defaultName.
func appendEntryImport(urls []string, name, lower string, cleanedFamilies []string, lastHrefs map[string]string, defaultName string) []string {
	if href, declared := lastHrefs[name]; declared {
		return appendUniqueURL(urls, href)
	}
	if (defaultName != "" && name == defaultName) || containsAny(cleanedFamilies, lower) {
		return appendUniqueURL(urls, GoogleFontsMapping[name])
	}
	return urls
}

// containsAny reports whether any of cleanedFamilies contains lower.
func containsAny(cleanedFamilies []string, lower string) bool {
	for _, cleaned := range cleanedFamilies {
		if strings.Contains(cleaned, lower) {
			return true
		}
	}
	return false
}

// appendUniqueURL appends url to urls unless it is empty or already there.
func appendUniqueURL(urls []string, url string) []string {
	if url == "" || slices.Contains(urls, url) {
		return urls
	}
	return append(urls, url)
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
