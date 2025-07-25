package mjml

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

// TestMJMLAgainstMRML compares Go implementation output with MRML (Rust) output
func TestMJMLAgainstMRML(t *testing.T) {
	testCases := []struct {
		name     string
		filename string
	}{
		{"basic", "testdata/basic.mjml"},
		{"with-head", "testdata/with-head.mjml"},
		{"complex-layout", "testdata/complex-layout.mjml"},
		{"wrapper-basic", "testdata/wrapper-basic.mjml"},
		{"wrapper-background", "testdata/wrapper-background.mjml"},
		{"wrapper-fullwidth", "testdata/wrapper-fullwidth.mjml"},
		{"wrapper-border", "testdata/wrapper-border.mjml"},
		{"group-footer-test", "testdata/group-footer-test.mjml"},
		{"section-padding-top-zero", "testdata/section-padding-top-zero.mjml"},
		{"Austin layout from the MJML.io site", "testdata/austin-layout-from-mjml-io.mjml"},
		// Austin layout component tests
		{"austin-header-section", "testdata/austin-header-section.mjml"},
		{"austin-hero-images", "testdata/austin-hero-images.mjml"},
		{"austin-wrapper-basic", "testdata/austin-wrapper-basic.mjml"},
		{"austin-text-with-links", "testdata/austin-text-with-links.mjml"},
		{"austin-buttons", "testdata/austin-buttons.mjml"},
		{"austin-two-column-images", "testdata/austin-two-column-images.mjml"},
		{"austin-divider", "testdata/austin-divider.mjml"},
		{"austin-two-column-text", "testdata/austin-two-column-text.mjml"},
		{"austin-full-width-wrapper", "testdata/austin-full-width-wrapper.mjml"},
		{"austin-social-media", "testdata/austin-social-media.mjml"},
		{"austin-footer-text", "testdata/austin-footer-text.mjml"},
		{"austin-group-component", "testdata/austin-group-component.mjml"},
		{"austin-global-attributes", "testdata/austin-global-attributes.mjml"},
		{"austin-map-image", "testdata/austin-map-image.mjml"},
		// MRML reference tests
		{"mrml-divider-basic", "testdata/mrml-divider-basic.mjml"},
		{"mrml-text-basic", "testdata/mrml-text-basic.mjml"},
		{"mrml-button-basic", "testdata/mrml-button-basic.mjml"},
		{"body-wrapper-section", "testdata/body-wrapper-section.mjml"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Read test MJML file
			mjmlContent, err := os.ReadFile(tc.filename)
			if err != nil {
				t.Fatalf("Failed to read test file %s: %v", tc.filename, err)
			}

			// Get expected output from MRML (Rust implementation)
			expected, err := runMRML(string(mjmlContent))
			if err != nil {
				t.Fatalf("Failed to run MRML: %v", err)
			}

			// Get actual output from Go implementation (direct library usage)
			actual, err := Render(string(mjmlContent))
			if err != nil {
				t.Fatalf("Failed to render MJML: %v", err)
			}

			// Compare outputs using DOM tree comparison
			if !compareDOMTrees(expected, actual) {
				// Enhanced DOM-based diff with debugging
				diff := createDOMDiff(expected, actual)
				t.Errorf("\n%s", diff)

				// Enhanced debugging: analyze style differences
				t.Logf("Style differences for %s:", tc.name)
				compareStyles(t, expected, actual)

				// For debugging: write both outputs to temp files
				os.WriteFile("/tmp/expected_"+tc.name+".html", []byte(expected), 0o644)
				os.WriteFile("/tmp/actual_"+tc.name+".html", []byte(actual), 0o644)
			}
		})
	}
}

// runMRML calls the MRML (Rust) implementation to get expected output
func runMRML(mjmlInput string) (string, error) {
	// Create temporary file for input
	tmpFile, err := os.CreateTemp("", "test_*.mjml")
	if err != nil {
		return "", err
	}
	defer os.Remove(tmpFile.Name())

	// Write MJML input to temp file
	if _, err := tmpFile.WriteString(mjmlInput); err != nil {
		tmpFile.Close()
		return "", err
	}
	tmpFile.Close()

	// Run mrml command with correct syntax
	cmd := exec.Command("mrml", tmpFile.Name(), "render")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return string(output), nil
}

// normalizeHTML normalizes HTML for comparison by removing extra whitespace
func normalizeHTML(html string) string {
	// Remove leading/trailing whitespace
	html = strings.TrimSpace(html)

	// Normalize line endings
	html = strings.ReplaceAll(html, "\r\n", "\n")
	html = strings.ReplaceAll(html, "\r", "\n")

	// Remove extra whitespace between tags (but preserve content whitespace)
	lines := strings.Split(html, "\n")
	var normalizedLines []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			normalizedLines = append(normalizedLines, line)
		}
	}

	return strings.Join(normalizedLines, "\n")
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
	html, err := component.Render()
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	// Verify output
	if !strings.Contains(html, "Test") {
		t.Error("Output should contain test text")
	}
}

// createSimpleDiff creates a character-level diff that shows exactly where they differ
func createSimpleDiff(expected, actual string) string {
	// ANSI color codes
	red := "\033[31m"
	green := "\033[32m"
	reset := "\033[0m"
	bold := "\033[1m"

	// Normalize strings for comparison
	expectedClean := strings.TrimSpace(expected)
	actualClean := strings.TrimSpace(actual)

	// Find first character difference
	minLen := len(expectedClean)
	if len(actualClean) < minLen {
		minLen = len(actualClean)
	}

	diffPos := -1
	for i := 0; i < minLen; i++ {
		if expectedClean[i] != actualClean[i] {
			diffPos = i
			break
		}
	}

	// If no character differences in common length, difference is at the end
	if diffPos == -1 && len(expectedClean) != len(actualClean) {
		diffPos = minLen
	}

	if diffPos == -1 {
		return "No differences found"
	}

	// Show context around the difference (50 chars before, 100 chars after)
	contextBefore := 50
	contextAfter := 100

	start := diffPos - contextBefore
	if start < 0 {
		start = 0
	}

	// Get expected snippet
	expectedEnd := diffPos + contextAfter
	if expectedEnd > len(expectedClean) {
		expectedEnd = len(expectedClean)
	}
	expectedSnippet := expectedClean[start:expectedEnd]

	// Get actual snippet
	actualEnd := diffPos + contextAfter
	if actualEnd > len(actualClean) {
		actualEnd = len(actualClean)
	}
	actualSnippet := actualClean[start:actualEnd]

	// Mark the difference position within the snippet
	markerPos := diffPos - start

	// Create visual markers
	expectedMarker := ""
	actualMarker := ""

	if markerPos < len(expectedSnippet) {
		expectedMarker = expectedSnippet[:markerPos] + bold + red + string(
			expectedSnippet[markerPos],
		) + reset + expectedSnippet[markerPos+1:]
	} else {
		expectedMarker = expectedSnippet + bold + red + "EOF" + reset
	}

	if markerPos < len(actualSnippet) {
		actualMarker = actualSnippet[:markerPos] + bold + green + string(
			actualSnippet[markerPos],
		) + reset + actualSnippet[markerPos+1:]
	} else {
		actualMarker = actualSnippet + bold + green + "EOF" + reset
	}

	return fmt.Sprintf("DIFF at position %d:\n- MRML (expected): %s%s%s\n+ gomjml (actual): %s%s%s",
		diffPos,
		red, expectedMarker, reset,
		green, actualMarker, reset)
}

// compareStyles analyzes and compares CSS styles between expected and actual output
func compareStyles(t *testing.T, expected, actual string) {
	styleRegex := regexp.MustCompile(`style="([^"]*)"`)

	expectedStyles := extractStyles(expected, styleRegex)
	actualStyles := extractStyles(actual, styleRegex)

	// Compare number of styled elements
	if len(expectedStyles) != len(actualStyles) {
		t.Logf("  Style count mismatch: expected %d, actual %d", len(expectedStyles), len(actualStyles))
	}

	// Compare individual styles
	maxLen := len(expectedStyles)
	if len(actualStyles) > maxLen {
		maxLen = len(actualStyles)
	}

	for i := 0; i < maxLen; i++ {
		var expectedStyle, actualStyle string
		if i < len(expectedStyles) {
			expectedStyle = expectedStyles[i]
		}
		if i < len(actualStyles) {
			actualStyle = actualStyles[i]
		}

		if expectedStyle != actualStyle {
			t.Logf("  Style %d mismatch:", i+1)
			t.Logf("    Expected: %s", expectedStyle)
			t.Logf("    Actual:   %s", actualStyle)

			// Analyze individual CSS properties
			compareStyleProperties(t, expectedStyle, actualStyle)
		}
	}
}

// extractStyles extracts all style attributes from HTML
func extractStyles(html string, regex *regexp.Regexp) []string {
	matches := regex.FindAllStringSubmatch(html, -1)
	styles := make([]string, len(matches))
	for i, match := range matches {
		if len(match) > 1 {
			styles[i] = match[1]
		}
	}
	return styles
}

// compareStyleProperties compares individual CSS properties within a style attribute
func compareStyleProperties(t *testing.T, expected, actual string) {
	expectedProps := parseStyleProperties(expected)
	actualProps := parseStyleProperties(actual)

	// Find properties only in expected
	for prop, value := range expectedProps {
		if actualValue, exists := actualProps[prop]; !exists {
			t.Logf("      Missing property: %s: %s", prop, value)
		} else if actualValue != value {
			t.Logf("      Property mismatch %s: expected '%s', actual '%s'", prop, value, actualValue)
		}
	}

	// Find properties only in actual
	for prop, value := range actualProps {
		if _, exists := expectedProps[prop]; !exists {
			t.Logf("      Extra property: %s: %s", prop, value)
		}
	}
}

// parseStyleProperties parses a CSS style string into a map of properties
func parseStyleProperties(style string) map[string]string {
	props := make(map[string]string)
	if style == "" {
		return props
	}

	declarations := strings.Split(style, ";")
	for _, decl := range declarations {
		decl = strings.TrimSpace(decl)
		if decl == "" {
			continue
		}

		parts := strings.SplitN(decl, ":", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			props[key] = value
		}
	}

	return props
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
		if expectedText != actualText {
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
	compareTags := []string{"head", "body", "table", "td", "div", "span"}
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

	expectedDoc.Find("[style]").Each(func(i int, expectedEl *goquery.Selection) {
		expectedStyle, _ := expectedEl.Attr("style")
		expectedTag := goquery.NodeName(expectedEl)

		// Find corresponding element in actual document
		actualEl := actualDoc.Find(expectedTag).Eq(i)
		if actualEl.Length() == 0 {
			diffs = append(diffs, fmt.Sprintf("  Missing styled %s element at position %d", expectedTag, i))
			return
		}

		actualStyle, exists := actualEl.Attr("style")
		if !exists {
			diffs = append(diffs, fmt.Sprintf("  %s[%d] missing style attribute", expectedTag, i))
			return
		}

		normalizedExpected := normalizeStyleAttribute(expectedStyle)
		normalizedActual := normalizeStyleAttribute(actualStyle)

		if normalizedExpected != normalizedActual {
			diffs = append(diffs, fmt.Sprintf("  %s[%d] style differs:", expectedTag, i))
			diffs = append(diffs, fmt.Sprintf("    Expected: %s", normalizedExpected))
			diffs = append(diffs, fmt.Sprintf("    Actual:   %s", normalizedActual))
		}
	})

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
