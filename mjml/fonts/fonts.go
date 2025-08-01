package fonts

import (
	"fmt"
	"strings"
)

const (
	// DefaultFontStack is the default font family used by MJML components
	DefaultFontStack = "Ubuntu, Helvetica, Arial, sans-serif"
)

// GoogleFontsMapping maps font family names to their Google Fonts URLs
var GoogleFontsMapping = map[string]string{
	"Ubuntu":     "https://fonts.googleapis.com/css?family=Ubuntu:300,400,500,700",
	"Open Sans":  "https://fonts.googleapis.com/css?family=Open+Sans:300,400,500,700",
	"Roboto":     "https://fonts.googleapis.com/css?family=Roboto:300,400,500,700",
	"Lato":       "https://fonts.googleapis.com/css?family=Lato:300,400,500,700",
	"Montserrat": "https://fonts.googleapis.com/css?family=Montserrat:300,400,500,700",
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

// GetGoogleFontURL checks if a font family corresponds to a Google Font and returns its URL
func GetGoogleFontURL(fontFamily string) string {
	// Clean up the font family string - remove quotes and extra whitespace
	fontFamily = strings.Trim(fontFamily, `"' `)

	// Check each Google Font mapping
	for fontName, url := range GoogleFontsMapping {
		// Case-insensitive check and see if the font family contains this font name
		if strings.Contains(strings.ToLower(fontFamily), strings.ToLower(fontName)) {
			return url
		}
	}

	return ""
}

// ConvertFontFamiliesToURLs converts a slice of font families to Google Font URLs
func ConvertFontFamiliesToURLs(fontFamilies []string) []string {
	var urls []string
	seen := make(map[string]bool)

	for _, fontFamily := range fontFamilies {
		if url := GetGoogleFontURL(fontFamily); url != "" {
			if !seen[url] {
				urls = append(urls, url)
				seen[url] = true
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
