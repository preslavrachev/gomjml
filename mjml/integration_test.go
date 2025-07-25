package mjml

import (
	"os"
	"os/exec"
	"regexp"
	"strings"
	"testing"
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

			// Compare outputs (normalize whitespace for comparison)
			expectedNorm := normalizeHTML(expected)
			actualNorm := normalizeHTML(actual)

			if expectedNorm != actualNorm {
				t.Errorf("Output mismatch for %s\n\nExpected:\n%s\n\nActual:\n%s", tc.name, expected, actual)

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
