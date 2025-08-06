package mjml

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/preslavrachev/gomjml/mjml/options"
)

// TestMJMLAgainstExpected compares Go implementation output with pre-generated expected HTML
func TestMJMLAgainstExpected(t *testing.T) {
	testCases := []struct {
		name              string
		filename          string
		testMJMLRoundTrip bool // Set to true to enable MJML round-trip test for this specific test case
	}{
		{name: "basic", filename: "testdata/basic.mjml", testMJMLRoundTrip: false},
		{name: "with-head", filename: "testdata/with-head.mjml", testMJMLRoundTrip: false},
		{name: "complex-layout", filename: "testdata/complex-layout.mjml", testMJMLRoundTrip: false},
		{name: "wrapper-basic", filename: "testdata/wrapper-basic.mjml", testMJMLRoundTrip: false},
		{name: "wrapper-background", filename: "testdata/wrapper-background.mjml", testMJMLRoundTrip: false},
		{name: "wrapper-fullwidth", filename: "testdata/wrapper-fullwidth.mjml", testMJMLRoundTrip: false},
		{name: "wrapper-border", filename: "testdata/wrapper-border.mjml", testMJMLRoundTrip: false},
		{name: "group-footer-test", filename: "testdata/group-footer-test.mjml", testMJMLRoundTrip: false},
		{name: "section-padding-top-zero", filename: "testdata/section-padding-top-zero.mjml", testMJMLRoundTrip: false},
		//{name: "Austin layout from the MJML.io site", filename: "testdata/austin-layout-from-mjml-io.mjml", testMJMLRoundTrip: false},
		// Austin layout component tests
		{name: "austin-header-section", filename: "testdata/austin-header-section.mjml", testMJMLRoundTrip: false},
		{name: "austin-hero-images", filename: "testdata/austin-hero-images.mjml", testMJMLRoundTrip: false},
		{name: "austin-wrapper-basic", filename: "testdata/austin-wrapper-basic.mjml", testMJMLRoundTrip: false},
		{name: "austin-text-with-links", filename: "testdata/austin-text-with-links.mjml", testMJMLRoundTrip: false},
		{name: "austin-buttons", filename: "testdata/austin-buttons.mjml", testMJMLRoundTrip: false},
		{name: "austin-two-column-images", filename: "testdata/austin-two-column-images.mjml", testMJMLRoundTrip: false},
		{name: "austin-divider", filename: "testdata/austin-divider.mjml", testMJMLRoundTrip: false},
		{name: "austin-two-column-text", filename: "testdata/austin-two-column-text.mjml", testMJMLRoundTrip: false},
		{name: "austin-full-width-wrapper", filename: "testdata/austin-full-width-wrapper.mjml", testMJMLRoundTrip: false},
		//{name: "austin-social-media", filename: "testdata/austin-social-media.mjml", testMJMLRoundTrip: false},
		{name: "austin-footer-text", filename: "testdata/austin-footer-text.mjml", testMJMLRoundTrip: false},
		{name: "austin-group-component", filename: "testdata/austin-group-component.mjml", testMJMLRoundTrip: false},
		{name: "austin-global-attributes", filename: "testdata/austin-global-attributes.mjml", testMJMLRoundTrip: false},
		{name: "austin-map-image", filename: "testdata/austin-map-image.mjml", testMJMLRoundTrip: false},
		// MRML reference tests
		{name: "mrml-divider-basic", filename: "testdata/mrml-divider-basic.mjml", testMJMLRoundTrip: false},
		{name: "mrml-text-basic", filename: "testdata/mrml-text-basic.mjml", testMJMLRoundTrip: false},
		{name: "mrml-button-basic", filename: "testdata/mrml-button-basic.mjml", testMJMLRoundTrip: false},
		{name: "body-wrapper-section", filename: "testdata/body-wrapper-section.mjml", testMJMLRoundTrip: false},
		// MJ-Group tests from MRML
		{name: "mj-group", filename: "testdata/mj-group.mjml", testMJMLRoundTrip: false},
		{name: "mj-group-background-color", filename: "testdata/mj-group-background-color.mjml", testMJMLRoundTrip: false},
		{name: "mj-group-class", filename: "testdata/mj-group-class.mjml", testMJMLRoundTrip: false},
		{name: "mj-group-direction", filename: "testdata/mj-group-direction.mjml", testMJMLRoundTrip: false},
		{name: "mj-group-vertical-align", filename: "testdata/mj-group-vertical-align.mjml", testMJMLRoundTrip: false},
		{name: "mj-group-width", filename: "testdata/mj-group-width.mjml", testMJMLRoundTrip: false},
		// Simple MJML components from MRML test suite
		{name: "mj-text", filename: "testdata/mj-text.mjml", testMJMLRoundTrip: false},
		{name: "mj-text-class", filename: "testdata/mj-text-class.mjml", testMJMLRoundTrip: false},
		{name: "mj-button", filename: "testdata/mj-button.mjml", testMJMLRoundTrip: false},
		{name: "mj-button-class", filename: "testdata/mj-button-class.mjml", testMJMLRoundTrip: false},
		{name: "mj-image", filename: "testdata/mj-image.mjml", testMJMLRoundTrip: false},
		{name: "mj-image-class", filename: "testdata/mj-image-class.mjml", testMJMLRoundTrip: false},
		{name: "mj-section-with-columns", filename: "testdata/mj-section-with-columns.mjml", testMJMLRoundTrip: false},
		{name: "mj-section", filename: "testdata/mj-section.mjml", testMJMLRoundTrip: false},
		{name: "mj-section-class", filename: "testdata/mj-section-class.mjml", testMJMLRoundTrip: false},
		{name: "mj-column", filename: "testdata/mj-column.mjml", testMJMLRoundTrip: false},
		{name: "mj-column-padding", filename: "testdata/mj-column-padding.mjml", testMJMLRoundTrip: false},
		{name: "mj-column-class", filename: "testdata/mj-column-class.mjml", testMJMLRoundTrip: false},
		{name: "mj-wrapper", filename: "testdata/mj-wrapper.mjml", testMJMLRoundTrip: false},
		// MJ-RAW tests
		{name: "mj-raw", filename: "testdata/mj-raw.mjml", testMJMLRoundTrip: false},
		{name: "mj-raw-conditional-comment", filename: "testdata/mj-raw-conditional-comment.mjml", testMJMLRoundTrip: false},
		{name: "mj-raw-go-template", filename: "testdata/mj-raw-go-template.mjml", testMJMLRoundTrip: false},
		// MJ-SOCIAL tests
		{name: "mj-social", filename: "testdata/mj-social.mjml", testMJMLRoundTrip: false},
		{name: "mj-social-align", filename: "testdata/mj-social-align.mjml", testMJMLRoundTrip: false},
		{name: "mj-social-border-radius", filename: "testdata/mj-social-border-radius.mjml", testMJMLRoundTrip: false},
		{name: "mj-social-class", filename: "testdata/mj-social-class.mjml", testMJMLRoundTrip: false},
		{name: "mj-social-color", filename: "testdata/mj-social-color.mjml", testMJMLRoundTrip: false},
		{name: "mj-social-container-background-color", filename: "testdata/mj-social-container-background-color.mjml", testMJMLRoundTrip: false},
		{name: "mj-social-element-ending", filename: "testdata/mj-social-element-ending.mjml", testMJMLRoundTrip: false},
		{name: "mj-social-font-family", filename: "testdata/mj-social-font-family.mjml", testMJMLRoundTrip: false},
		{name: "mj-social-font", filename: "testdata/mj-social-font.mjml", testMJMLRoundTrip: false},
		{name: "mj-social-icon", filename: "testdata/mj-social-icon.mjml", testMJMLRoundTrip: false},
		{name: "mj-social-link", filename: "testdata/mj-social-link.mjml", testMJMLRoundTrip: false},
		{name: "mj-social-mode", filename: "testdata/mj-social-mode.mjml", testMJMLRoundTrip: false},
		{name: "mj-social-padding", filename: "testdata/mj-social-padding.mjml", testMJMLRoundTrip: false},
		{name: "mj-social-text", filename: "testdata/mj-social-text.mjml", testMJMLRoundTrip: false},
		// MJ-ACCORDION tests (commented out - need implementation)
		{name: "mj-accordion", filename: "testdata/mj-accordion.mjml", testMJMLRoundTrip: false},
		{name: "mj-accordion-font-padding", filename: "testdata/mj-accordion-font-padding.mjml", testMJMLRoundTrip: false},
		{name: "mj-accordion-icon", filename: "testdata/mj-accordion-icon.mjml", testMJMLRoundTrip: false},
		{name: "mj-accordion-other", filename: "testdata/mj-accordion-other.mjml", testMJMLRoundTrip: false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Read test MJML file
			mjmlContent, err := os.ReadFile(tc.filename)
			if err != nil {
				t.Fatalf("Failed to read test file %s: %v", tc.filename, err)
			}

			// Get expected output from cached HTML file
			expectedFile := strings.Replace(tc.filename, ".mjml", ".html", 1)
			expectedContent, err := os.ReadFile(expectedFile)
			if err != nil {
				// Handle special case for conditional comments
				if tc.name == "mj-raw-conditional-comment" {
					t.Logf(
						"No expected HTML file for %s due to conditional comments, checking that our implementation works",
						tc.name,
					)

					// Just verify our implementation doesn't crash and produces some output
					actual, err := RenderHTML(string(mjmlContent))
					if err != nil {
						t.Fatalf("Failed to render MJML: %v", err)
					}

					if len(actual) == 0 {
						t.Fatal("Expected non-empty output")
					}

					// Verify it contains the expected raw content
					if !strings.Contains(actual, `<div>mso</div>`) {
						t.Error("Expected output to contain <div>mso</div>")
					}
					if !strings.Contains(actual, `<span>general</span>`) {
						t.Error("Expected output to contain <span>general</span>")
					}
					if !strings.Contains(actual, `<img src="bananas" alt=""`) {
						t.Error("Expected output to contain img tag with bananas src")
					}

					return
				}
				t.Fatalf("Failed to read expected HTML file %s: %v", expectedFile, err)
			}
			expected := string(expectedContent)

			// Get actual output from Go implementation (direct library usage)
			actual, err := RenderHTML(string(mjmlContent))
			if err != nil {
				t.Fatalf("Failed to render MJML: %v", err)
			}

			// Compare outputs using DOM tree comparison
			if !compareDOMTrees(expected, actual) {
				// Enhanced DOM-based diff with debugging
				domDiff := createDOMDiff(expected, actual)
				t.Errorf("\n%s", domDiff)

				// Enhanced debugging: analyze style differences with precise element identification
				t.Logf("Style differences for %s:", tc.name)
				compareStylesPrecise(t, expected, actual)

				// For debugging: write both outputs to temp files
				os.WriteFile("/tmp/expected_"+tc.name+".html", []byte(expected), 0o644)
				os.WriteFile("/tmp/actual_"+tc.name+".html", []byte(actual), 0o644)
			}

			// MJML round-trip test (if enabled for this test case)
			if tc.testMJMLRoundTrip {
				t.Run(tc.name+"_mjml_roundtrip", func(t *testing.T) {
					runMJMLRoundTripTest(t, string(mjmlContent), tc.name)
				})
			}
		})
	}
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
	html, err := RenderHTML(mjmlInput)
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

	// RenderHTML to HTML
	html, err := RenderComponentString(component)
	if err != nil {
		t.Fatalf("RenderHTML failed: %v", err)
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
		} else if expected.Style != actual.Style {
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

// runMJMLRoundTripTest tests MJML round-trip: MJML -> AST -> MJML and compares with original
func runMJMLRoundTripTest(t *testing.T, originalMJML, testName string) {
	// Parse original MJML to AST
	ast, err := ParseMJML(originalMJML)
	if err != nil {
		t.Fatalf("Failed to parse MJML for round-trip test: %v", err)
	}

	// Render AST back to MJML
	renderedMJML, err := RenderFromAST(ast, WithOutputFormat(options.OutputMJML))
	if err != nil {
		t.Fatalf("Failed to render AST to MJML: %v", err)
	}

	// Compare original and rendered MJML
	if !compareMJMLContent(originalMJML, renderedMJML) {
		t.Errorf("MJML round-trip failed for %s", testName)
		t.Logf("Original MJML length: %d", len(originalMJML))
		t.Logf("Rendered MJML length: %d", len(renderedMJML))

		// Write both to temp files for debugging
		os.WriteFile("/tmp/original_"+testName+".mjml", []byte(originalMJML), 0o644)
		os.WriteFile("/tmp/rendered_"+testName+".mjml", []byte(renderedMJML), 0o644)

		// Show first difference
		showMJMLDiff(t, originalMJML, renderedMJML, testName)
	}
}

// compareMJMLContent compares two MJML strings, normalizing whitespace and structure
func compareMJMLContent(original, rendered string) bool {
	// Basic normalization - remove extra whitespace, normalize line endings
	normalizeWhitespace := func(s string) string {
		// Normalize line endings
		s = strings.ReplaceAll(s, "\r\n", "\n")
		s = strings.ReplaceAll(s, "\r", "\n")

		// Split into lines and trim each
		lines := strings.Split(s, "\n")
		var normalizedLines []string
		for _, line := range lines {
			trimmed := strings.TrimSpace(line)
			if trimmed != "" {
				normalizedLines = append(normalizedLines, trimmed)
			}
		}

		return strings.Join(normalizedLines, "\n")
	}

	normalizedOriginal := normalizeWhitespace(original)
	normalizedRendered := normalizeWhitespace(rendered)

	return normalizedOriginal == normalizedRendered
}

// showMJMLDiff shows the first significant difference between original and rendered MJML
func showMJMLDiff(t *testing.T, original, rendered, testName string) {
	originalLines := strings.Split(original, "\n")
	renderedLines := strings.Split(rendered, "\n")

	maxLines := max(len(originalLines), len(renderedLines))

	for i := 0; i < maxLines; i++ {
		var origLine, rendLine string
		if i < len(originalLines) {
			origLine = strings.TrimSpace(originalLines[i])
		}
		if i < len(renderedLines) {
			rendLine = strings.TrimSpace(renderedLines[i])
		}

		if origLine != rendLine {
			t.Logf("First difference at line %d:", i+1)
			t.Logf("  Original: %q", origLine)
			t.Logf("  Rendered: %q", rendLine)
			break
		}
	}
}
