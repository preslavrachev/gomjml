package mjml

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/preslavrachev/gomjml/mjml/components"
	"github.com/preslavrachev/gomjml/mjml/testutils"
)

/*
TestMJMLAgainstExpected runs a suite of integration tests to verify that the MJML rendering
implementation produces HTML output matching the expected results for a variety of MJML input files,
created using the MRML CLI.

// AIDEV-NOTE: If you are unsure about a test output, try the htmlcompare utility for a semantic diff.
// Example:
//   cd mjml/testdata && ../../bin/htmlcompare basic
// Or from project root:
//   ./bin/htmlcompare basic --testdata-dir mjml/testdata

For each test case, it reads the corresponding MJML file from the "testdata" directory, using the test case name
(e.g., "basic") to construct the filename "testdata/basic.mjml". It then renders the MJML to HTML using the Render function,
and compares the output to a pre-generated expected HTML file named "testdata/basic.html".

The mapping is:
  - For test case "foo", the MJML input is at "testdata/foo.mjml"
  - The expected HTML output is at "testdata/foo.html"

On mismatch, the test provides a detailed DOM diff, logs style differences, and writes both
actual and expected outputs to temporary files for debugging purposes.
*/
func TestMJMLAgainstExpected(t *testing.T) {
	// Reset navbar ID counter for deterministic testing
	components.ResetNavbarIDCounter()
	// Reset carousel ID counter for deterministic testing
	components.ResetCarouselIDCounter()
	testCases := []string{
		"mjml",
		"mj-body",
		"basic",
		"comment",
		"with-head",
		"complex-layout",
		"wrapper-basic",
		"wrapper-background",
		"wrapper-fullwidth",
		"wrapper-border",
		"group-footer-test",
		"section-padding-top-zero",
		//"austin-layout-from-mjml-io", // Commented out
		// Austin layout component tests
		"austin-header-section",
		"austin-hero-images",
		"austin-wrapper-basic",
		"austin-text-with-links",
		"austin-buttons",
		"austin-two-column-images",
		"austin-divider",
		"austin-two-column-text",
		"austin-full-width-wrapper",
		//"austin-social-media", // Commented out
		"austin-footer-text",
		"austin-group-component",
		"austin-global-attributes",
		"austin-map-image",
		// MRML reference tests
		"mrml-divider-basic",
		"mrml-text-basic",
		"mrml-button-basic",
		"body-wrapper-section",
		"mj-attributes",
		// MJ-Group tests from MRML
		"mj-group",
		"mj-group-background-color",
		"mj-group-class",
		"mj-group-direction",
		"mj-group-vertical-align",
		"mj-group-width",
		// Simple MJML components from MRML test suite
		"mj-text",
		"mj-text-class",
		"mj-button",
		"mj-button-class",
		"mj-image",
		"mj-image-class",
		"mj-image-src-with-url-params",
		"mj-section-with-columns",
		"mj-section",
		"mj-section-class",
		"mj-column",
		"mj-column-padding",
		"mj-column-class",
		"mj-wrapper",
		// MJ-RAW tests
		"mj-raw",
		"mj-raw-conditional-comment",
		"mj-raw-go-template",
		// MJ-SOCIAL tests
		"mj-social",
		"mj-social-align",
		"mj-social-border-radius",
		"mj-social-class",
		"mj-social-color",
		"mj-social-container-background-color",
		"mj-social-element-ending",
		"mj-social-font-family",
		"mj-social-font",
		"mj-social-icon",
		"mj-social-link",
		"mj-social-mode",
		"mj-social-padding",
		"mj-social-text",
		// MJ-ACCORDION tests
		"mj-accordion",
		"mj-accordion-font-padding",
		"mj-accordion-icon",
		"mj-accordion-other",
		// MJ-NAVBAR tests
		"mj-navbar",
		"mj-navbar-ico",
		"mj-navbar-align-class",
		"mj-navbar-multiple",
		// MJ-HERO tests
		"mj-hero",
		"mj-hero-background-color",
		"mj-hero-background-height",
		"mj-hero-background-position",
		"mj-hero-background-url",
		"mj-hero-background-width",
		"mj-hero-class",
		"mj-hero-height",
		"mj-hero-mode",
		"mj-hero-vertical-align",
		// MJ-SPACER test
		"mj-spacer",
		// MJ-TABLE tests
		"mj-table",
		"mj-table-other",
		"mj-table-table",
		"mj-table-text",
		// MJ-CAROUSEL tests
		"mj-carousel",
		"mj-carousel-align-border-radius-class",
		"mj-carousel-icon",
		"mj-carousel-tb",
		"mj-carousel-thumbnails",
		// Custom test cases
		"notifuse-open-br-tags",
		//"notifuse-full",
	}

	for _, testName := range testCases {
		t.Run(testName, func(t *testing.T) {
			// Generate filename from test name
			filename := getTestdataFilename(testName)

			// Read test MJML file
			mjmlContent, err := os.ReadFile(filename)
			if err != nil {
				t.Fatalf("Failed to read test file %s: %v", filename, err)
			}

			// Get expected output from cached HTML file
			expectedFile := strings.Replace(filename, ".mjml", ".html", 1)
			expectedContent, err := os.ReadFile(expectedFile)
			if err != nil {
				t.Fatalf("Failed to read expected HTML file %s: %v", expectedFile, err)
			}
			expected := string(expectedContent)

			// Get actual output from Go implementation (direct library usage)
			actual, err := Render(string(mjmlContent))
			if err != nil {
				t.Fatalf("Failed to render MJML: %v", err)
			}

			// Compare outputs using DOM tree comparison
			if !compareDOMTrees(expected, actual) {
				// Enhanced DOM-based diff with debugging
				domDiff := createDOMDiff(expected, actual)
				t.Errorf("\n%s", domDiff)

				// Enhanced debugging: analyze style differences with precise element identification
				t.Logf("Style differences for %s:", testName)
				compareStylesPrecise(t, expected, actual)

				writeDebugFiles(testName, expected, actual)
			} else {
				// DOM trees match, but check for self-closing tag serialization differences
				if selfClosingDiff := checkSelfClosingTagDifferences(expected, actual); selfClosingDiff != "" {
					t.Errorf("Self-closing tag serialization differences found:\n%s", selfClosingDiff)
					writeDebugFiles(testName, expected, actual)
				}
			}
		})
	}
}

// getTestdataFilename returns the file path for a test MJML file located in the "testdata" directory,
// using the provided testName as the base filename. The resulting path has the format "testdata/{testName}.mjml".
func getTestdataFilename(testName string) string {
	return fmt.Sprintf("testdata/%s.mjml", testName)
}

// writeDebugFiles writes both expected and actual HTML outputs to temp files for debugging
func writeDebugFiles(testName, expected, actual string) {
	// For debugging: write both outputs to temp files
	os.WriteFile("/tmp/expected_"+testName+".html", []byte(expected), 0o644)
	os.WriteFile("/tmp/actual_"+testName+".html", []byte(actual), 0o644)
}

// TestDirectLibraryUsage demonstrates and tests direct library usage
func TestDirectLibraryUsage(t *testing.T) {
	mjmlInput := `<mjml>
		<mj-body>
			<mj-section>
				<mj-column>
					<mj-text>Direct library test</mj-text>
				</mj-column>
			</mj-section>
		</mj-body>
	</mjml>`

	// Test direct library usage as documented in the restructuring plan
	html, err := Render(mjmlInput)
	if err != nil {
		t.Fatalf("Direct library usage failed: %v", err)
	}

	// Verify basic structure
	if !strings.Contains(html, "<!doctype html>") {
		t.Error("Output should contain DOCTYPE")
	}

	if !strings.Contains(html, "Direct library test") {
		t.Error("Output should contain the text content")
	}

	if !strings.Contains(html, "mj-column-per-100") {
		t.Error("Output should contain responsive CSS classes")
	}
}

// TestComponentCreation tests the component creation pattern from the plan
func TestComponentCreation(t *testing.T) {
	mjmlInput := `<mjml><mj-body><mj-section><mj-column><mj-text>Test</mj-text></mj-column></mj-section></mj-body></mjml>`

	// Parse MJML
	ast, err := ParseMJML(mjmlInput)
	if err != nil {
		t.Fatalf("ParseMJML failed: %v", err)
	}

	// Create component tree
	component, err := NewFromAST(ast)
	if err != nil {
		t.Fatalf("NewFromAST failed: %v", err)
	}

	// Render to HTML
	html, err := RenderComponentString(component)
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	// Verify output
	if !strings.Contains(html, "Test") {
		t.Error("Output should contain test text")
	}
}

// TestCSSNormalization tests the CSS content normalization function
func TestCSSNormalization(t *testing.T) {
	testCases := []struct {
		name     string
		css1     string
		css2     string
		expected bool
	}{
		{
			name:     "identical CSS",
			css1:     ".mj-column-per-100 { width:100% }",
			css2:     ".mj-column-per-100 { width:100% }",
			expected: true,
		},
		{
			name:     "different order CSS rules",
			css1:     ".mj-column-per-100 { width:100% } .mj-column-per-50 { width:50% }",
			css2:     ".mj-column-per-50 { width:50% } .mj-column-per-100 { width:100% }",
			expected: true,
		},
		{
			name:     "different whitespace",
			css1:     ".mj-column-per-100{width:100%}.mj-column-per-50{width:50%}",
			css2:     ".mj-column-per-100 { width: 100% } .mj-column-per-50 { width: 50% }",
			expected: true,
		},
		{
			name:     "different content",
			css1:     ".mj-column-per-100 { width:100% }",
			css2:     ".mj-column-per-100 { width:50% }",
			expected: false,
		},
		{
			name:     "complex media query reordering",
			css1:     "@media only screen { .mj-column-per-100 { width:100% } .mj-column-per-50 { width:50% } }",
			css2:     "@media only screen { .mj-column-per-50 { width:50% } .mj-column-per-100 { width:100% } }",
			expected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			normalized1 := normalizeCSSContent(tc.css1)
			normalized2 := normalizeCSSContent(tc.css2)

			result := normalized1 == normalized2
			if result != tc.expected {
				t.Errorf("Expected %v, got %v", tc.expected, result)
				t.Logf("CSS1: %s", tc.css1)
				t.Logf("CSS2: %s", tc.css2)
				t.Logf("Normalized CSS1: %s", normalized1)
				t.Logf("Normalized CSS2: %s", normalized2)
			}
		})
	}
}

// compareStylesPrecise provides exact element identification for style differences
func compareStylesPrecise(t *testing.T, expected, actual string) {
	expectedDoc, err1 := goquery.NewDocumentFromReader(strings.NewReader(expected))
	actualDoc, err2 := goquery.NewDocumentFromReader(strings.NewReader(actual))

	if err1 != nil || err2 != nil {
		t.Logf("DOM parsing failed: expected=%v, actual=%v", err1, err2)
		return
	}

	// Build ordered lists of styled elements
	var expectedElements []ElementInfo
	var actualElements []ElementInfo

	expectedDoc.Find("[style]").Each(func(i int, el *goquery.Selection) {
		style, _ := el.Attr("style")
		classes, _ := el.Attr("class")
		tagName := goquery.NodeName(el)

		expectedElements = append(expectedElements, ElementInfo{
			Tag:     tagName,
			Classes: classes,
			Style:   style,
			Index:   i,
		})
	})

	actualDoc.Find("[style]").Each(func(i int, el *goquery.Selection) {
		style, _ := el.Attr("style")
		classes, _ := el.Attr("class")
		tagName := goquery.NodeName(el)

		// Extract debug info to identify which MJML component created this element
		debugComponent := ""
		if debugAttr, exists := el.Attr("data-mj-debug-group"); exists && debugAttr == "true" {
			debugComponent = "mj-group"
		} else if debugAttr, exists := el.Attr("data-mj-debug-column"); exists && debugAttr == "true" {
			debugComponent = "mj-column"
		} else if debugAttr, exists := el.Attr("data-mj-debug-section"); exists && debugAttr == "true" {
			debugComponent = "mj-section"
		} else if debugAttr, exists := el.Attr("data-mj-debug-text"); exists && debugAttr == "true" {
			debugComponent = "mj-text"
		} else if debugAttr, exists := el.Attr("data-mj-debug-wrapper"); exists && debugAttr == "true" {
			debugComponent = "mj-wrapper"
		}

		actualElements = append(actualElements, ElementInfo{
			Tag:       tagName,
			Classes:   classes,
			Style:     style,
			Index:     i,
			Component: debugComponent,
		})
	})

	// Compare element by element
	maxLen := max(len(expectedElements), len(actualElements))
	for i := 0; i < maxLen; i++ {
		var expected, actual *ElementInfo
		if i < len(expectedElements) {
			expected = &expectedElements[i]
		}
		if i < len(actualElements) {
			actual = &actualElements[i]
		}

		if expected == nil {
			componentInfo := ""
			if actual.Component != "" {
				componentInfo = fmt.Sprintf(" [created by %s]", actual.Component)
			}
			t.Logf("  Extra element[%d]: <%s class=\"%s\" style=\"%s\">%s",
				i, actual.Tag, actual.Classes, actual.Style, componentInfo)
		} else if actual == nil {
			t.Logf("  Missing element[%d]: <%s class=\"%s\" style=\"%s\">",
				i, expected.Tag, expected.Classes, expected.Style)
		} else if !testutils.StylesEqual(expected.Style, actual.Style) {
			componentInfo := ""
			if actual.Component != "" {
				componentInfo = fmt.Sprintf(" [created by %s]", actual.Component)
			}
			t.Logf("  Style diff element[%d]: <%s class=\"%s\">%s",
				i, actual.Tag, actual.Classes, componentInfo)
			t.Logf("    Expected: style=\"%s\"", expected.Style)
			t.Logf("    Actual:   style=\"%s\"", actual.Style)

			// Show specific property differences
			expectedProps := parseStyleProperties(expected.Style)
			actualProps := parseStyleProperties(actual.Style)

			for prop, expectedVal := range expectedProps {
				if actualVal, exists := actualProps[prop]; !exists {
					t.Logf("    Missing property: %s=%s", prop, expectedVal)
				} else if actualVal != expectedVal {
					t.Logf("    Wrong value: %s=%s (expected %s)", prop, actualVal, expectedVal)
				}
			}

			for prop, actualVal := range actualProps {
				if _, exists := expectedProps[prop]; !exists {
					t.Logf("    Extra property: %s=%s", prop, actualVal)
				}
			}
		}
	}
}

// ElementInfo represents a styled HTML element
type ElementInfo struct {
	Tag       string
	Classes   string
	Style     string
	Index     int
	Component string // Which MJML component created this element (from debug attrs)
}

// parseStyleProperties parses CSS style string into property map
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

		parts := strings.SplitN(decl, ":", 2)
		if len(parts) == 2 {
			prop := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			props[prop] = value
		}
	}

	return props
}

// max returns the maximum of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// StyleDiff represents differences between expected and actual CSS properties
type StyleDiff struct {
	Missing    map[string]string    // prop: expectedValue
	Mismatched map[string][2]string // prop: [expected, actual]
	Extra      map[string]string    // prop: actualValue
}

// IsEmpty returns true if there are no differences
func (d StyleDiff) IsEmpty() bool {
	return len(d.Missing) == 0 && len(d.Mismatched) == 0 && len(d.Extra) == 0
}

// String formats the diff for readable output
func (d StyleDiff) String() string {
	if d.IsEmpty() {
		return ""
	}

	var parts []string

	if len(d.Missing) > 0 {
		var missing []string
		for prop, value := range d.Missing {
			missing = append(missing, fmt.Sprintf("%s=%s", prop, value))
		}
		parts = append(parts, fmt.Sprintf("Missing: %s", strings.Join(missing, ", ")))
	}

	if len(d.Mismatched) > 0 {
		var mismatched []string
		for prop, values := range d.Mismatched {
			mismatched = append(mismatched, fmt.Sprintf("%s=%s→%s", prop, values[0], values[1]))
		}
		parts = append(parts, fmt.Sprintf("Wrong values: %s", strings.Join(mismatched, ", ")))
	}

	if len(d.Extra) > 0 {
		var extra []string
		for prop, value := range d.Extra {
			extra = append(extra, fmt.Sprintf("%s=%s", prop, value))
		}
		parts = append(parts, fmt.Sprintf("Extra: %s", strings.Join(extra, ", ")))
	}

	return strings.Join(parts, " | ")
}

// compareStylePropertiesMaps compares two CSS property maps directly
func compareStylePropertiesMaps(expectedProps, actualProps map[string]string) StyleDiff {
	diff := StyleDiff{
		Missing:    make(map[string]string),
		Mismatched: make(map[string][2]string),
		Extra:      make(map[string]string),
	}

	// Find properties only in expected (missing)
	for prop, expectedValue := range expectedProps {
		if actualValue, exists := actualProps[prop]; !exists {
			diff.Missing[prop] = expectedValue
		} else if actualValue != expectedValue {
			diff.Mismatched[prop] = [2]string{expectedValue, actualValue}
		}
	}

	// Find properties only in actual (extra)
	for prop, actualValue := range actualProps {
		if _, exists := expectedProps[prop]; !exists {
			diff.Extra[prop] = actualValue
		}
	}

	return diff
}

// compareDOMTrees compares two HTML strings using DOM tree comparison
// This approach handles attribute ordering, CSS property ordering, and whitespace normalization
func compareDOMTrees(expected, actual string) bool {
	expectedDoc, err := goquery.NewDocumentFromReader(strings.NewReader(expected))
	if err != nil {
		return false
	}

	actualDoc, err := goquery.NewDocumentFromReader(strings.NewReader(actual))
	if err != nil {
		return false
	}

	return compareNodes(expectedDoc.Selection, actualDoc.Selection)
}

// compareNodes recursively compares two goquery selections (DOM subtrees)
func compareNodes(expected, actual *goquery.Selection) bool {
	// Compare number of nodes
	if expected.Length() != actual.Length() {
		return false
	}

	// Compare each node pair
	equal := true
	expected.Each(func(i int, expectedNode *goquery.Selection) {
		if i >= actual.Length() {
			equal = false
			return
		}

		actualNode := actual.Eq(i)

		// Compare node types and tag names
		expectedTag := goquery.NodeName(expectedNode)
		actualTag := goquery.NodeName(actualNode)
		if expectedTag != actualTag {
			equal = false
			return
		}

		// For text nodes, compare text content (normalized)
		if expectedTag == "#text" {
			expectedText := strings.TrimSpace(expectedNode.Text())
			actualText := strings.TrimSpace(actualNode.Text())
			if expectedText != actualText {
				equal = false
				return
			}
			return
		}

		// For element nodes, compare attributes
		if !compareAttributes(expectedNode, actualNode) {
			equal = false
			return
		}

		// Recursively compare children
		expectedChildren := expectedNode.Children()
		actualChildren := actualNode.Children()
		if !compareNodes(expectedChildren, actualChildren) {
			equal = false
			return
		}

		// Compare text content for elements that might have mixed content
		expectedText := strings.TrimSpace(expectedNode.Contents().Not("*").Text())
		actualText := strings.TrimSpace(actualNode.Contents().Not("*").Text())

		// Special handling for style tags - check for specific CSS issues first
		if expectedTag == "style" {
			// Check for Firefox-specific .moz-text-html prefix issues
			if hasFirefoxCSSIssue(expectedText, actualText) {
				equal = false
				return
			}
			// Then apply general CSS normalization for ordering issues
			if normalizeCSSContent(expectedText) != normalizeCSSContent(actualText) {
				equal = false
				return
			}
		} else if expectedText != actualText {
			equal = false
			return
		}
	})

	return equal
}

// compareAttributes compares attributes between two nodes, normalizing style attributes
func compareAttributes(expected, actual *goquery.Selection) bool {
	// Get all attributes from both nodes
	expectedAttrs := make(map[string]string)
	actualAttrs := make(map[string]string)

	// Extract expected attributes
	if expected.Length() > 0 {
		node := expected.Get(0)
		for _, attr := range node.Attr {
			if attr.Key == "style" {
				expectedAttrs[attr.Key] = normalizeStyleAttribute(attr.Val)
			} else if !strings.HasPrefix(attr.Key, "data-mj-debug") { // Skip debug attributes
				expectedAttrs[attr.Key] = attr.Val
			}
		}
	}

	// Extract actual attributes
	if actual.Length() > 0 {
		node := actual.Get(0)
		for _, attr := range node.Attr {
			if attr.Key == "style" {
				actualAttrs[attr.Key] = normalizeStyleAttribute(attr.Val)
			} else if !strings.HasPrefix(attr.Key, "data-mj-debug") { // Skip debug attributes
				actualAttrs[attr.Key] = attr.Val
			}
		}
	}

	// Compare attribute maps
	if len(expectedAttrs) != len(actualAttrs) {
		return false
	}

	for key, expectedVal := range expectedAttrs {
		actualVal, exists := actualAttrs[key]
		if !exists || expectedVal != actualVal {
			return false
		}
	}

	return true
}

// normalizeStyleAttribute normalizes CSS style attributes by sorting properties
func normalizeStyleAttribute(style string) string {
	if style == "" {
		return ""
	}

	// Parse CSS properties
	props := parseStyleProperties(style)

	// Sort properties by key for consistent comparison
	var keys []string
	for key := range props {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	// Rebuild style string with sorted properties
	var parts []string
	for _, key := range keys {
		if value := strings.TrimSpace(props[key]); value != "" {
			parts = append(parts, key+": "+value)
		}
	}

	result := strings.Join(parts, "; ")
	if result != "" && !strings.HasSuffix(result, ";") {
		result += ";"
	}

	return result
}

// createDOMDiff creates a detailed diff report using DOM analysis
func createDOMDiff(expected, actual string) string {
	// ANSI color codes
	red := "\033[31m"
	reset := "\033[0m"
	bold := "\033[1m"

	expectedDoc, err1 := goquery.NewDocumentFromReader(strings.NewReader(expected))
	actualDoc, err2 := goquery.NewDocumentFromReader(strings.NewReader(actual))

	if err1 != nil || err2 != nil {
		return fmt.Sprintf("DOM parsing failed: expected=%v, actual=%v", err1, err2)
	}

	var diffs []string

	// Compare document structures
	expectedHtml := expectedDoc.Find("html")
	actualHtml := actualDoc.Find("html")

	if expectedHtml.Length() != actualHtml.Length() {
		diffs = append(diffs, fmt.Sprintf("%sStructure mismatch:%s Expected %d html elements, got %d",
			bold, reset, expectedHtml.Length(), actualHtml.Length()))
	}

	// Compare specific elements that commonly differ
	// 50 most common HTML tags for comparison
	compareTags := []string{
		"a", "abbr", "address", "area", "article", "aside", "audio", "b", "base", "bdi",
		"bdo", "blockquote", "body", "br", "button", "canvas", "caption", "cite", "code", "col",
		"colgroup", "data", "datalist", "dd", "del", "details", "dfn", "dialog", "div", "dl",
		"dt", "em", "embed", "fieldset", "figcaption", "figure", "footer", "form", "h1", "h2",
		"h3", "h4", "h5", "h6", "head", "header", "hr", "html", "i", "iframe", "img",
		"input", "ins", "kbd", "label", "legend", "li", "link", "main", "map", "mark",
		"meta", "meter", "nav", "noscript", "object", "ol", "optgroup", "option", "output", "p",
		"param", "picture", "pre", "progress", "q", "rb", "rp", "rt", "rtc", "ruby",
		"s", "samp", "script", "section", "select", "small", "source", "span", "strong", "style",
		"sub", "summary", "sup", "svg", "table", "tbody", "td", "template", "textarea", "tfoot",
		"th", "thead", "time", "title", "tr", "track", "u", "ul", "var", "video",
	}
	for _, tag := range compareTags {
		expectedCount := expectedDoc.Find(tag).Length()
		actualCount := actualDoc.Find(tag).Length()
		if expectedCount != actualCount {
			diffs = append(diffs, fmt.Sprintf("%s%s count mismatch:%s Expected %d, got %d",
				bold, tag, reset, expectedCount, actualCount))
		}
	}

	// Compare style attributes
	styleComparison := compareAllStyleAttributes(expectedDoc, actualDoc)
	if styleComparison != "" {
		diffs = append(diffs, styleComparison)
	}

	// Compare debug attributes to identify problematic MJML components
	debugComparison := analyzeDebugAttributes(actualDoc)
	if debugComparison != "" {
		diffs = append(diffs, debugComparison)
	}

	if len(diffs) == 0 {
		return "DOM structures match but content differs. Check text content and attribute values."
	}

	return fmt.Sprintf("%sDOM Differences:%s\n%s%s%s",
		bold, reset,
		red, strings.Join(diffs, "\n"), reset)
}

// compareAllStyleAttributes compares all style attributes in the documents
func compareAllStyleAttributes(expectedDoc, actualDoc *goquery.Document) string {
	var diffs []string

	// Build maps of elements by tag name for proper position tracking
	expectedElements := make(map[string][]*goquery.Selection)
	actualElements := make(map[string][]*goquery.Selection)

	// Collect expected styled elements by tag
	expectedDoc.Find("[style]").Each(func(i int, el *goquery.Selection) {
		tag := goquery.NodeName(el)
		expectedElements[tag] = append(expectedElements[tag], el)
	})

	// Collect actual styled elements by tag
	actualDoc.Find("[style]").Each(func(i int, el *goquery.Selection) {
		tag := goquery.NodeName(el)
		actualElements[tag] = append(actualElements[tag], el)
	})

	// Compare each tag type
	for tag, expectedList := range expectedElements {
		actualList, exists := actualElements[tag]
		if !exists {
			diffs = append(diffs, fmt.Sprintf("  Missing all styled %s elements (expected %d)", tag, len(expectedList)))
			continue
		}

		if len(expectedList) != len(actualList) {
			diffs = append(
				diffs,
				fmt.Sprintf(
					"  %s element count mismatch: expected %d, actual %d",
					tag,
					len(expectedList),
					len(actualList),
				),
			)
		}

		// Aggregate all CSS properties for this tag type
		expectedProps := make(map[string]string)
		actualProps := make(map[string]string)

		// Collect all expected properties for this tag type
		for _, el := range expectedList {
			if style, exists := el.Attr("style"); exists {
				tagProps := parseStyleProperties(style)
				for prop, value := range tagProps {
					expectedProps[prop] = value
				}
			}
		}

		// Collect all actual properties for this tag type
		for _, el := range actualList {
			if style, exists := el.Attr("style"); exists {
				tagProps := parseStyleProperties(style)
				for prop, value := range tagProps {
					actualProps[prop] = value
				}
			}
		}

		// Compare aggregated properties for this tag type
		if len(expectedProps) > 0 || len(actualProps) > 0 {
			styleDiff := compareStylePropertiesMaps(expectedProps, actualProps)
			if !styleDiff.IsEmpty() {
				diffs = append(diffs, fmt.Sprintf("  %s elements: %s", tag, styleDiff.String()))
			}
		}
	}

	// Check for actual elements that don't exist in expected
	for tag, actualList := range actualElements {
		if _, exists := expectedElements[tag]; !exists {
			diffs = append(diffs, fmt.Sprintf("  Unexpected styled %s elements (found %d)", tag, len(actualList)))
		}
	}

	if len(diffs) == 0 {
		return ""
	}

	return "Style attribute differences:\n" + strings.Join(diffs, "\n")
}

// analyzeDebugAttributes analyzes debug attributes to identify which MJML components are present
func analyzeDebugAttributes(actualDoc *goquery.Document) string {
	var analysis []string

	// Count debug attributes by component type
	debugCounts := make(map[string]int)

	actualDoc.Find("[data-mj-debug-text]").Each(func(i int, s *goquery.Selection) {
		debugCounts["text"]++
	})

	actualDoc.Find("[data-mj-debug-button]").Each(func(i int, s *goquery.Selection) {
		debugCounts["button"]++
	})

	actualDoc.Find("[data-mj-debug-image]").Each(func(i int, s *goquery.Selection) {
		debugCounts["image"]++
	})

	actualDoc.Find("[data-mj-debug-column]").Each(func(i int, s *goquery.Selection) {
		debugCounts["column"]++
	})

	actualDoc.Find("[data-mj-debug-section]").Each(func(i int, s *goquery.Selection) {
		debugCounts["section"]++
	})

	actualDoc.Find("[data-mj-debug-wrapper]").Each(func(i int, s *goquery.Selection) {
		debugCounts["wrapper"]++
	})

	actualDoc.Find("[data-mj-debug-divider]").Each(func(i int, s *goquery.Selection) {
		debugCounts["divider"]++
	})

	actualDoc.Find("[data-mj-debug-social-element]").Each(func(i int, s *goquery.Selection) {
		debugCounts["social-element"]++
	})

	if len(debugCounts) > 0 {
		analysis = append(analysis, "MJML Components found in actual output:")
		for component, count := range debugCounts {
			analysis = append(analysis, fmt.Sprintf("  - %s: %d instances", component, count))
		}

		// Show MJML tag context for better debugging
		tagInfo := getMJMLTagInfo(actualDoc)
		if len(tagInfo) > 0 {
			analysis = append(analysis, "MJML Tags referenced:")
			for tag, count := range tagInfo {
				analysis = append(analysis, fmt.Sprintf("  - <%s>: %d instances", tag, count))
			}
		}

		// Identify likely problematic components based on common failure patterns
		if debugCounts["social-element"] > 0 && debugCounts["divider"] > 0 {
			analysis = append(analysis, "  ⚠️  Social and divider components often require missing dependencies")
		}
		if debugCounts["button"] > 0 {
			analysis = append(analysis, "  ⚠️  Button components may have MSO rendering differences")
		}
		if debugCounts["wrapper"] > 0 {
			analysis = append(analysis, "  ⚠️  Wrapper components may have background/border style issues")
		}
	}

	if len(analysis) == 0 {
		return ""
	}

	return strings.Join(analysis, "\n")
}

// getMJMLTagInfo extracts MJML tag information from debug attributes
func getMJMLTagInfo(doc *goquery.Document) map[string]int {
	tagCounts := make(map[string]int)

	doc.Find("[data-mj-tag]").Each(func(i int, s *goquery.Selection) {
		if tag, exists := s.Attr("data-mj-tag"); exists && tag != "" {
			tagCounts[tag]++
		}
	})

	return tagCounts
}

// hasFirefoxCSSIssue checks for specific Firefox CSS issues like missing .moz-text-html prefixes
func hasFirefoxCSSIssue(expected, actual string) bool {
	// Only check if this looks like a Firefox-specific style tag
	if !strings.Contains(expected, ".moz-text-html") {
		return false // Not a Firefox CSS style, no issue
	}

	// Simple heuristic: if expected has ".moz-text-html" but actual is missing some instances
	expectedCount := strings.Count(expected, ".moz-text-html")
	actualCount := strings.Count(actual, ".moz-text-html")

	// If actual has fewer .moz-text-html prefixes than expected, it's likely an issue
	return actualCount < expectedCount
}

// normalizeCSSContent normalizes CSS content for comparison by removing whitespace and sorting characters
func normalizeCSSContent(css string) string {
	// Remove all whitespace and newlines
	normalized := strings.ReplaceAll(css, " ", "")
	normalized = strings.ReplaceAll(normalized, "\n", "")
	normalized = strings.ReplaceAll(normalized, "\t", "")
	normalized = strings.ReplaceAll(normalized, "\r", "")

	// Convert to slice of runes, sort, and convert back
	runes := []rune(normalized)
	sort.Slice(runes, func(i, j int) bool {
		return runes[i] < runes[j]
	})

	return string(runes)
}

// checkSelfClosingTagDifferences detects differences in self-closing tag serialization
// between expected and actual HTML that would be missed by DOM comparison
func checkSelfClosingTagDifferences(expected, actual string) string {
	// HTML5 void elements that should be self-closing
	voidTags := []string{"br", "hr", "img", "input", "meta", "link", "area", "base", "col", "embed", "source", "track", "wbr"}

	var differences []string

	for _, tag := range voidTags {
		// Count different serialization patterns for this tag
		expectedUnclosed := countTagPattern(expected, fmt.Sprintf("<%s>", tag))
		actualUnclosed := countTagPattern(actual, fmt.Sprintf("<%s>", tag))

		expectedSelfClosed := countTagPattern(expected, fmt.Sprintf("<%s/>", tag)) + countTagPattern(expected, fmt.Sprintf("<%s />", tag))
		actualSelfClosed := countTagPattern(actual, fmt.Sprintf("<%s/>", tag)) + countTagPattern(actual, fmt.Sprintf("<%s />", tag))

		// Check for differences in serialization
		if expectedUnclosed != actualUnclosed || expectedSelfClosed != actualSelfClosed {
			differences = append(differences,
				fmt.Sprintf("<%s> tag serialization mismatch:\n  Expected: %d unclosed + %d self-closed\n  Actual:   %d unclosed + %d self-closed",
					tag, expectedUnclosed, expectedSelfClosed, actualUnclosed, actualSelfClosed))
		}
	}

	if len(differences) > 0 {
		return strings.Join(differences, "\n")
	}

	return ""
}

// countTagPattern counts occurrences of a specific tag pattern in HTML
func countTagPattern(html, pattern string) int {
	return strings.Count(strings.ToLower(html), strings.ToLower(pattern))
}
