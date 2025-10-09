package mjml

import (
	"errors"
	"fmt"
	"os"
	"regexp"
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

	type testCase struct {
		name       string
		errHandler func(error) error
	}

	testCases := []testCase{
		// "mjml", -- MJML Badly formatted, must return an error instead
		{name: "mj-body"},
		{name: "mj-body-background-color"},
		{name: "mj-body-class"},
		{name: "mj-body-width"},
		{name: "basic"},
		{name: "comment"},
		{name: "with-head"},
		// {name: "complex-layout"},
		{name: "wrapper-basic"},
		{name: "wrapper-background"},
		{name: "wrapper-fullwidth"},
		{name: "wrapper-border"},
		{name: "group-footer-test"},
		{name: "section-bg-vml-color"},
		{name: "section-fullwidth-background-image"},
		{name: "section-fullwidth-bg-transparent"},
		{name: "section-padding-top-zero"},
		// //{name: "austin-layout-from-mjml-io"}, // Commented out
		// // Austin layout component tests
		{name: "austin-header-section"},
		{name: "austin-hero-images"},
		{name: "austin-wrapper-basic"},
		{name: "austin-text-with-links"},
		{name: "austin-buttons"},
		{name: "austin-two-column-images"},
		{name: "austin-divider"},
		{name: "mj-divider"},
		{name: "mj-divider-alignment"},
		{name: "mj-divider-border"},
		{name: "mj-divider-class"},
		{name: "mj-divider-container-background-color"},
		{name: "mj-divider-in-mj-text"},
		{name: "mj-divider-padding"},
		{name: "mj-divider-width"},
		{name: "mj-divider-container-background-transparent"},
		{name: "austin-two-column-text"},
		{name: "austin-full-width-wrapper"},
		{name: "austin-social-media"},
		{name: "austin-footer-text"},
		{name: "austin-group-component"},
		{name: "austin-global-attributes"},
		{name: "austin-map-image"},
		// // MRML reference tests
		{name: "mrml-divider-basic"},
		{name: "mrml-text-basic"},
		{name: "mrml-button-basic"},
		{name: "body-wrapper-section"},
		{name: "mj-attributes"},
		// // MJ-Group tests from MRML
		{name: "mj-group"},
		{name: "mj-group-background-color"},
		{name: "mj-group-class"},
		{name: "mj-group-mso-wrapper-raw"},
		{name: "mj-group-direction"},
		{name: "mj-group-vertical-align"},
		{name: "mj-group-width"},
		// Simple MJML components from MRML test suite
		{name: "mj-button"},
		{name: "mj-button-align"},
		{name: "mj-button-background"},
		{name: "mj-button-border"},
		{name: "mj-button-border-radius"},
		{name: "mj-button-class"},
		{name: "mj-button-color"},
		{name: "mj-button-container-background-color"},
		{name: "mj-button-example"},
		{name: "mj-button-font-family"},
		{name: "mj-button-font-size"},
		{name: "mj-button-font-style"},
		{name: "mj-button-font-weight"},
		{name: "mj-button-height"},
		{name: "mj-button-href"},
		{name: "mj-button-inner-padding"},
		{name: "mj-button-line-height"},
		{name: "mj-button-padding"},
		{name: "mj-button-text-decoration"},
		{name: "mj-button-text-transform"},
		{name: "mj-button-vertical-align"},
		{name: "mj-button-width"},
		{name: "mj-button-global-attributes"},
		{name: "mj-image"},
		{name: "mj-image-align"},
		{name: "mj-image-border"},
		{name: "mj-image-border-radius"},
		{name: "mj-image-container-background-color"},
		{name: "mj-image-fluid-on-mobile"},
		{name: "mj-image-height"},
		{name: "mj-image-href"},
		{name: "mj-image-padding"},
		{name: "mj-image-rel"},
		// {name: "mj-image-title"},
		{name: "mj-image-class"},
		// {name: "mj-image-src-with-url-params"},
		// {name: "mj-section"},
		{name: "mj-section-background-vml"},
		// {name: "mj-section-background-color"},
		{name: "mj-section-background-url"},
		// {name: "mj-section-background-url-full"},
		{name: "mj-section-body-width"},
		// {name: "mj-section-border"},
		{name: "mj-section-border-radius"},
		// {name: "mj-section-direction"},
		{name: "mj-section-full-width"},
		// {name: "mj-section-padding"},
		// {name: "mj-section-text-align"},
		{name: "mj-section-bg-cover-no-repeat"},
		// {name: "mj-section-global-attributes"},
		{name: "mj-section-width"},
		// {name: "mj-section-with-columns"},
		{name: "mj-section-class"},
		// {name: "mj-column"},
		{name: "mj-column-background-color"},
		// {name: "mj-column-border"},
		{name: "mj-column-border-issue-466"},
		// {name: "mj-column-border-radius"},
		{name: "mj-column-inner-background-color"},
		// {name: "mj-column-vertical-align"},
		{name: "mj-column-padding"},
		// {name: "mj-column-class"},
		// {name: "mj-column-global-attributes"},
		{name: "mj-wrapper"},
		// {name: "mj-wrapper-border"},
		// {name: "mj-wrapper-border-radius"},
		// {name: "mj-wrapper-multiple-sections"},
		{name: "mj-wrapper-other"},
		// {name: "mj-wrapper-padding"},
		// // MJ-Text tests
		// {name: "mj-text"},
		{name: "mj-text-align"},
		{name: "mj-text-color"},
		{name: "mj-text-container-background-color"},
		{name: "mj-text-decoration"},
		// {name: "mj-text-example"},
		{name: "mj-text-font-family"},
		// {name: "mj-text-font-size"},
		// {name: "mj-text-font-style"},
		{name: "mj-text-font-weight"},
		// {name: "mj-text-class"},
		// // MJ-RAW tests
		// {name: "mj-raw"},
		{name: "mj-raw-conditional-comment"},
		// {name: "mj-raw-head"}, // MJML says file badly formatted
		{name: "mj-raw-go-template"},
		// // MJ-SOCIAL tests
		// {name: "mj-social"},
		{name: "mj-social-anchors"},
		// {name: "mj-social-align"},
		{name: "mj-social-border-radius"},
		// {name: "mj-social-class"},
		{name: "mj-social-color"},
		// {name: "mj-social-complex-styling"},
		// {name: "mj-social-container-background-color"},
		{name: "mj-social-element-ending"},
		{name: "mj-social-font-family"},
		// {name: "mj-social-font"},
		{name: "mj-social-icon"},
		// {name: "mj-social-link"},
		// {name: "mj-social-mode"},
		{name: "mj-social-notifuse"},
		// {name: "mj-social-padding"},
		// {name: "mj-social-structure-basic"},
		{name: "mj-social-text"},
		{name: "mj-social-text-wrapper"},
		{name: "mj-social-no-ubuntu-fonts-overridden"},
		{name: "mj-social-ubuntu-fonts-with-text-content"},
		// {name: "mj-social-ubuntu-fonts-icons-only-fallback"},
		// // MJ-ACCORDION tests
		{name: "mj-accordion"},
		// {name: "mj-accordion-font-padding"},
		{name: "mj-accordion-icon"},
		{name: "mj-accordion-other"},
		// // MJ-NAVBAR tests
		{name: "mj-navbar"},
		// {name: "mj-navbar-ico"},
		{name: "mj-navbar-align-class"},
		// {name: "mj-navbar-multiple"},
		// // MJ-HERO tests
		{name: "mj-hero"},
		{name: "mj-hero-background-color"},
		{name: "mj-hero-background-height"},
		// {name: "mj-hero-background-position"},
		// {name: "mj-hero-background-url"},
		{name: "mj-hero-background-width"},
		// {name: "mj-hero-class"},
		// {name: "mj-hero-height"},
		{name: "mj-hero-width", errHandler: func(err error) error {
			expectedErr := ErrInvalidAttribute("mj-hero", "width", 3)
			if err.Error() == expectedErr.Error() {
				return nil
			}
			return expectedErr
		}},
		// {name: "mj-hero-mode"},
		// {name: "mj-hero-vertical-align"},
		// // MJ-SPACER test
		// {name: "mj-spacer"},
		// // MJ-TABLE tests
		{name: "mj-table"},
		// {name: "mj-table-other"},
		{name: "mj-table-table"},
		// {name: "mj-table-text"},
		// // MJ-CAROUSEL tests
		{name: "mj-carousel"},
		{name: "mj-carousel-align-border-radius-class"},
		{name: "mj-carousel-icon"},
		{name: "mj-carousel-tb"},
		{name: "mj-carousel-thumbnails"},
		// // Custom test cases
		{name: "notifuse-open-br-tags"},
		// {name: "notifuse-full"}, //-- NOTE: HTML has been compiled with the MJML compiler already.
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Generate filename from test name
			filename := getTestdataFilename(tc.name)

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

			// FAIL LOUD AND CLEAR: Check for empty expected HTML file
			if len(strings.TrimSpace(expected)) == 0 {
				t.Fatalf("❌ EMPTY EXPECTED HTML FILE: %s\n"+
					"🚨 The expected HTML file is completely empty! This indicates:\n"+
					"   - Missing reference implementation output\n"+
					"   - Failed HTML generation during test setup\n"+
					"   - Incomplete test case preparation\n"+
					"📝 Action required: Generate valid expected HTML content for this test case\n"+
					"💡 Hint: Use the reference MJML implementation to generate expected output",
					expectedFile)
			}

			// Get actual output from Go implementation (direct library usage)
			actual, err := Render(string(mjmlContent))
			if err != nil {
				handled := false
				if tc.errHandler != nil {
					var mjmlErr Error
					if errors.As(err, &mjmlErr) {
						if tc.errHandler(mjmlErr) == nil {
							handled = true
						} else {
							t.Fatalf("Error did not match expectation: %v", err)
						}
					}
				}
				if handled {
					return
				}
				// Unexpected error or no error handler - fail the test
				t.Fatalf("Failed to render MJML: %v", err)
			}

			// If we expected an error but got none, fail
			if tc.errHandler != nil {
				t.Fatalf("Expected the following error: %s, but got none", tc.errHandler(errors.New("no error")))
			}

			// Collect ALL difference types instead of early returns for comprehensive analysis
			var allDifferences []string

			// Normalize legacy MJML reference quirks (eg. MRML style conditionals) so that
			// comparisons operate on semantically equivalent markup. This keeps the testdata
			// fixtures stable while letting the Go renderer follow the upstream MJML output.
			normalizedExpected := normalizeMJMLReference(expected)
			normalizedActual := normalizeMJMLReference(actual)

			// Check for MSO table attribute differences FIRST (before DOM comparison)
			// because MSO conditionals are not part of DOM and will be ignored by DOM comparison
			msoTableDiff := checkMSOTableAttributeDifferences(normalizedExpected, normalizedActual)
			if msoTableDiff != "" {
				allDifferences = append(allDifferences, "MSO table attribute differences found:\n"+msoTableDiff)
			}

			// Check for MSO conditional comment differences
			msoDiff := checkMSOConditionalDifferences(normalizedExpected, normalizedActual)
			if msoDiff != "" {
				allDifferences = append(allDifferences, "MSO conditional comment differences found:\n"+msoDiff)
			}

			// Compare outputs using DOM tree comparison
			domTreesMatch := compareDOMTrees(normalizedExpected, normalizedActual)
			if !domTreesMatch {
				// Check for HTML entity encoding differences
				entityDiff := checkHTMLEntityDifferences(normalizedExpected, normalizedActual)
				if entityDiff != "" {
					allDifferences = append(allDifferences, "HTML entity encoding differences found:\n"+entityDiff)
				}

				// Check for VML attribute differences
				vmlDiff := checkVMLAttributeDifferences(normalizedExpected, normalizedActual)
				if vmlDiff != "" {
					allDifferences = append(allDifferences, "VML attribute differences found:\n"+vmlDiff)
				}

				// Check for background CSS property differences
				bgDiff := checkBackgroundPropertyDifferences(normalizedExpected, normalizedActual)
				if bgDiff != "" {
					allDifferences = append(allDifferences, "Background CSS property differences found:\n"+bgDiff)
				}

				// Enhanced DOM-based diff with debugging
				domDiff := createDOMDiff(normalizedExpected, normalizedActual)
				if domDiff != "" {
					allDifferences = append(allDifferences, "DOM structure differences:\n"+domDiff)
				}

				// Enhanced debugging: analyze style differences with precise element identification
				// AIDEV-NOTE: Only log style differences when they actually exist to reduce noise
				styleResult := testutils.CompareStylesPrecise(normalizedExpected, normalizedActual)
				if styleResult.ParseError != nil {
					allDifferences = append(allDifferences, fmt.Sprintf("DOM parsing failed: %v", styleResult.ParseError))
				} else if styleResult.HasDifferences {
					var styleDiffs []string
					styleDiffs = append(styleDiffs, fmt.Sprintf("Style differences for %s:", tc.name))
					for _, element := range styleResult.Elements {
						switch element.Status {
						case testutils.ElementExtra:
							componentInfo := ""
							if element.Component != "" {
								componentInfo = fmt.Sprintf(" [created by %s]", element.Component)
							}
							styleDiffs = append(styleDiffs, fmt.Sprintf("  Extra element[%d]: <%s class=\"%s\" style=\"%s\">%s",
								element.Index, element.Tag, element.Classes, element.Actual, componentInfo))
						case testutils.ElementMissing:
							styleDiffs = append(styleDiffs, fmt.Sprintf("  Missing element[%d]: <%s class=\"%s\" style=\"%s\">",
								element.Index, element.Tag, element.Classes, element.Expected))
						case testutils.ElementDifferent:
							componentInfo := ""
							if element.Component != "" {
								componentInfo = fmt.Sprintf(" [created by %s]", element.Component)
							}
							styleDiffs = append(styleDiffs, fmt.Sprintf("  Style diff element[%d]: <%s class=\"%s\">%s",
								element.Index, element.Tag, element.Classes, componentInfo))
							styleDiffs = append(styleDiffs, fmt.Sprintf("    Expected: style=\"%s\"", element.Expected))
							styleDiffs = append(styleDiffs, fmt.Sprintf("    Actual:   style=\"%s\"", element.Actual))
							if !element.StyleDiff.IsEmpty() {
								styleDiffs = append(styleDiffs, fmt.Sprintf("    %s", element.StyleDiff.String()))
							}
						}
					}
					if len(styleDiffs) > 0 {
						allDifferences = append(allDifferences, strings.Join(styleDiffs, "\n"))
					}
				}
			}

			// Check for self-closing tag serialization differences regardless of DOM tree match
			selfClosingDiff := checkSelfClosingTagDifferences(normalizedExpected, normalizedActual)
			if selfClosingDiff != "" {
				allDifferences = append(allDifferences, "Self-closing tag serialization differences found:\n"+selfClosingDiff)
			}

			// Report ALL collected differences
			if len(allDifferences) > 0 {
				writeDebugFiles(tc.name, expected, actual)
				t.Errorf("\n=== COMPREHENSIVE DIFFERENCE ANALYSIS ===\n%s\n===========================================",
					strings.Join(allDifferences, "\n\n"))
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

	// Also persist the normalized versions used during comparison for easier diffing
	normalizedExpected := normalizeMJMLReference(expected)
	normalizedActual := normalizeMJMLReference(actual)
	os.WriteFile("/tmp/normalized_expected_"+testName+".html", []byte(normalizedExpected), 0o644)
	os.WriteFile("/tmp/normalized_actual_"+testName+".html", []byte(normalizedActual), 0o644)
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
			switch {
			case attr.Key == "style":
				expectedAttrs[attr.Key] = normalizeStyleAttribute(attr.Val)
			case attr.Key == "class":
				expectedAttrs[attr.Key] = normalizeClassAttribute(attr.Val)
			case !strings.HasPrefix(attr.Key, "data-mj-debug"):
				expectedAttrs[attr.Key] = attr.Val
			}
		}
	}

	// Extract actual attributes
	if actual.Length() > 0 {
		node := actual.Get(0)
		for _, attr := range node.Attr {
			switch {
			case attr.Key == "style":
				actualAttrs[attr.Key] = normalizeStyleAttribute(attr.Val)
			case attr.Key == "class":
				actualAttrs[attr.Key] = normalizeClassAttribute(attr.Val)
			case !strings.HasPrefix(attr.Key, "data-mj-debug"):
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

func normalizeClassAttribute(class string) string {
	if class == "" {
		return ""
	}

	parts := strings.Fields(class)
	sort.Strings(parts)
	return strings.Join(parts, " ")
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

// createDOMDiff compares two HTML DOM strings and returns a formatted string describing their differences.
// It parses both expected and actual HTML strings, compares their structures, counts of common HTML tags,
// style attributes, and debug attributes. Differences are highlighted using ANSI color codes for readability.
// If no structural differences are found, it suggests checking text content and attribute values.
// Returns a human-readable summary of DOM differences or parsing errors.
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

	// If no structural or style differences were found, but the test still failed,
	// it means the DOM trees match in structure and attributes, but there may be
	// differences in text content, attribute values, or other subtle issues.
	// (This function is only called when compareDOMTrees returned false, so we know the test failed.)
	if len(diffs) == 0 {
		// First, check if the difference is just whitespace/formatting
		if testutils.NormalizeForComparison(expected) == testutils.NormalizeForComparison(actual) {
			return "" // No meaningful differences - whitespace/formatting only
		}

		// Last resort: compare character-sorted strings to detect reordering
		// With massive HTML strings, collision chance is astronomically low
		if sortStringChars(expected) == sortStringChars(actual) {
			return "DOM structures match and content is identical when sorted. Likely ordering-only differences."
		}
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

// sortStringChars sorts all characters in a string alphabetically
// Used to detect if two strings have identical content but different ordering
func sortStringChars(s string) string {
	chars := strings.Split(s, "")
	sort.Strings(chars)
	return strings.Join(chars, "")
}

// checkHTMLEntityDifferences detects differences in HTML entity encoding
// that would be normalized away by DOM parsing but are still meaningful
func checkHTMLEntityDifferences(expected, actual string) string {
	// Common HTML entity patterns that might differ
	entityPairs := []struct {
		encoded string
		decoded string
		name    string
	}{
		{"&amp;", "&", "ampersand"},
		{"&lt;", "<", "less-than"},
		{"&gt;", ">", "greater-than"},
		{"&quot;", "\"", "quote"},
		{"&#x27;", "'", "apostrophe"},
		{"&#39;", "'", "apostrophe-numeric"},
	}

	var differences []string

	for _, pair := range entityPairs {
		expectedCount := strings.Count(expected, pair.encoded)
		actualCount := strings.Count(actual, pair.encoded)

		expectedDecodedCount := strings.Count(expected, pair.decoded)
		actualDecodedCount := strings.Count(actual, pair.decoded)

		// If one uses encoded form and other uses decoded form
		if expectedCount != actualCount {
			if expectedCount > 0 && actualCount == 0 && actualDecodedCount > 0 {
				differences = append(differences,
					fmt.Sprintf("Expected uses encoded %s (%s) %d times, actual uses decoded (%s) %d times",
						pair.name, pair.encoded, expectedCount, pair.decoded, actualDecodedCount))
			} else if actualCount > 0 && expectedCount == 0 && expectedDecodedCount > 0 {
				differences = append(differences,
					fmt.Sprintf("Actual uses encoded %s (%s) %d times, expected uses decoded (%s) %d times",
						pair.name, pair.encoded, actualCount, pair.decoded, expectedDecodedCount))
			} else {
				differences = append(differences,
					fmt.Sprintf("%s encoding mismatch: expected %d encoded, actual %d encoded",
						pair.name, expectedCount, actualCount))
			}
		}
	}

	if len(differences) > 0 {
		return strings.Join(differences, "\n")
	}
	return ""
}

// checkMSOConditionalDifferences detects differences in MSO conditional comments
// that would be normalized away by DOM parsing but are still meaningful for email rendering
func checkMSOConditionalDifferences(expected, actual string) string {
	// Common MSO conditional patterns that might differ
	msoPatterns := []struct {
		pattern string
		name    string
	}{
		{"<!--[if mso]>", "mso-opening"},
		{"<!--[if !mso]><!-->", "not-mso-opening"},
		{"<!--<![endif]-->", "endif"},
		{"<!--[if mso | IE]>", "mso-or-ie-opening"},
		{"<!--[if !mso | IE]><!-->", "not-mso-or-ie-opening"},
		{"<!--[if IE] mso |>", "ie-mso-opening"}, // Added missing pattern
		{"<!--[if lte mso 11]>", "mso-lte-11-opening"},
		{"<![endif]-->", "simple-endif"},
	}

	var differences []string

	// Check for count differences first
	for _, pattern := range msoPatterns {
		expectedCount := strings.Count(expected, pattern.pattern)
		actualCount := strings.Count(actual, pattern.pattern)

		if expectedCount != actualCount {
			differences = append(differences,
				fmt.Sprintf("MSO conditional %s mismatch: expected %d, actual %d",
					pattern.name, expectedCount, actualCount))
		}
	}

	// Check for sequence/ordering differences by comparing MSO blocks
	if len(differences) == 0 {
		// Extract sequences of MSO conditionals + HTML elements for comparison
		expectedSequence := extractMSOSequences(expected)
		actualSequence := extractMSOSequences(actual)

		if len(expectedSequence) != len(actualSequence) {
			differences = append(differences,
				fmt.Sprintf("MSO sequence count mismatch: expected %d blocks, actual %d blocks",
					len(expectedSequence), len(actualSequence)))
		} else {
			// Compare each sequence block
			for i, expectedBlock := range expectedSequence {
				if i < len(actualSequence) && expectedBlock != actualSequence[i] {
					differences = append(differences,
						fmt.Sprintf("MSO sequence differs at block %d:\n  Expected: %s\n  Actual: %s",
							i, expectedBlock, actualSequence[i]))
				}
			}
		}
	}

	if len(differences) > 0 {
		return strings.Join(differences, "\n")
	}
	return ""
}

// checkVMLAttributeDifferences detects differences in VML attributes that are critical for email rendering
func checkVMLAttributeDifferences(expected, actual string) string {
	vmlPatterns := []struct {
		pattern string
		name    string
	}{
		{`position="([^"]*)"`, "position"},
		{`origin="([^"]*)"`, "origin"},
		{`color="([^"]*)"`, "color"},
		{`size="([^"]*)"`, "size"},
		{`type="([^"]*)"`, "type"},
		{`aspect="([^"]*)"`, "aspect"},
	}

	var differences []string

	for _, pattern := range vmlPatterns {
		// Count different values for this VML attribute
		expectedMatches := findRegexMatches(expected, pattern.pattern)
		actualMatches := findRegexMatches(actual, pattern.pattern)

		if len(expectedMatches) != len(actualMatches) {
			differences = append(differences,
				fmt.Sprintf("VML %s attribute count mismatch: expected %d, actual %d",
					pattern.name, len(expectedMatches), len(actualMatches)))
		} else {
			// Check for value differences
			for i, expectedVal := range expectedMatches {
				if i < len(actualMatches) && expectedVal != actualMatches[i] {
					differences = append(differences,
						fmt.Sprintf("VML %s attribute value mismatch: expected '%s', actual '%s'",
							pattern.name, expectedVal, actualMatches[i]))
				}
			}
		}
	}

	if len(differences) > 0 {
		return strings.Join(differences, "\n")
	}
	return ""
}

// checkMSOTableAttributeDifferences detects differences in MSO table attributes
func checkMSOTableAttributeDifferences(expected, actual string) string {
	msoTableAttrs := []string{"bgcolor", "width", "align", "cellpadding", "cellspacing"}

	var differences []string

	for _, attr := range msoTableAttrs {
		// Look for MSO conditional table attributes - simplified pattern
		// Only match <table ... bgcolor="..."> inside MSO conditional blocks
		expectedPattern := fmt.Sprintf(`<!--\[if mso.*?<table[^>]*%s="([^"]*)"`, attr)
		actualPattern := expectedPattern

		expectedMatches := findRegexMatches(expected, expectedPattern)
		actualMatches := findRegexMatches(actual, actualPattern)

		if len(expectedMatches) != len(actualMatches) {
			differences = append(differences,
				fmt.Sprintf("MSO table %s attribute count mismatch: expected %d, actual %d",
					attr, len(expectedMatches), len(actualMatches)))
		} else {
			// Check for value differences when counts match
			for i, expectedVal := range expectedMatches {
				if i < len(actualMatches) && expectedVal != actualMatches[i] {
					differences = append(differences,
						fmt.Sprintf("MSO table %s attribute value mismatch: expected '%s', actual '%s'",
							attr, expectedVal, actualMatches[i]))
				}
			}
		}
	}

	if len(differences) > 0 {
		return strings.Join(differences, "\n")
	}
	return ""
}

// checkBackgroundPropertyDifferences detects differences in CSS background properties
func checkBackgroundPropertyDifferences(expected, actual string) string {
	bgProps := []string{"background", "background-color", "background-image", "background-position", "background-size", "background-repeat"}

	var differences []string

	for _, prop := range bgProps {
		// Look for style attributes containing this background property
		pattern := fmt.Sprintf(`style="[^"]*%s:\s*([^;"]*)`, prop)

		expectedMatches := findRegexMatches(expected, pattern)
		actualMatches := findRegexMatches(actual, pattern)

		// Count unique values for each property
		expectedValues := make(map[string]int)
		actualValues := make(map[string]int)

		for _, match := range expectedMatches {
			expectedValues[match]++
		}
		for _, match := range actualMatches {
			actualValues[match]++
		}

		// Compare value distributions
		for value, expectedCount := range expectedValues {
			if actualCount := actualValues[value]; actualCount != expectedCount {
				differences = append(differences,
					fmt.Sprintf("CSS %s value '%s' count mismatch: expected %d, actual %d",
						prop, value, expectedCount, actualCount))
			}
		}

		// Check for extra values in actual
		for value, actualCount := range actualValues {
			if _, exists := expectedValues[value]; !exists {
				differences = append(differences,
					fmt.Sprintf("CSS %s has unexpected value '%s' (count: %d)",
						prop, value, actualCount))
			}
		}
	}

	if len(differences) > 0 {
		return strings.Join(differences, "\n")
	}
	return ""
}

// findRegexMatches finds all matches for a regex pattern and returns the first capture group
func findRegexMatches(text, pattern string) []string {
	re := regexp.MustCompile(pattern)
	matches := re.FindAllStringSubmatch(text, -1)

	var results []string
	for _, match := range matches {
		if len(match) > 1 {
			results = append(results, match[1])
		}
	}
	return results
}

// normalizeMJMLReference cleans up legacy MRML fixtures so comparisons focus on
// semantic differences rather than serialization quirks. It removes empty style
// tags and merges split MSO conditional blocks that MJML now emits as a single
// table/td wrapper.
var (
	mustacheAfterClosingPattern  = regexp.MustCompile(`(-->|>)\s+(\{\{)`)
	mustacheBeforeOpeningPattern = regexp.MustCompile(`(\}\})\s+(<!--|<)`)
)

func normalizeMJMLReference(html string) string {
	normalized := html

	// Merge split MSO conditionals of the form:
	// <!--[if mso | IE]><table ...><tr><![endif]-->\n<!-- [if mso | IE]><td ...><![endif]-->
	// into the modern MJML style: <!--[if mso | IE]><table ...><tr><td ...><![endif]-->
	msoSplitPattern := regexp.MustCompile(`<!--\[if mso \| IE\]><table([^>]*)><tr><!\[endif\]-->\s*<!--\[if mso \| IE\]><td([^>]*)><!\[endif\]-->`)
	normalized = msoSplitPattern.ReplaceAllStringFunc(normalized, func(match string) string {
		submatches := msoSplitPattern.FindStringSubmatch(match)
		if len(submatches) != 3 {
			return match
		}

		tableAttrs := submatches[1]
		tdAttrs := strings.TrimSpace(submatches[2])

		if !strings.Contains(tdAttrs, "class=") {
			if tdAttrs == "" {
				tdAttrs = `class=""`
			} else {
				tdAttrs = `class="" ` + tdAttrs
			}
		}

		if tdAttrs != "" {
			tdAttrs = " " + tdAttrs
		}

		// Preserve attribute spacing but ensure we add a space before the closing angle
		// bracket to match MJML's serialized output.
		return fmt.Sprintf("<!--[if mso | IE]><table%s><tr><td%s ><![endif]-->", tableAttrs, tdAttrs)
	})

	// Merge split closing sequences: <!--[if mso | IE]></td><![endif]--><!--[if mso | IE]></tr></table><![endif]-->
	msoClosePattern := regexp.MustCompile(`<!--\[if mso \| IE\]>\s*</td>\s*<!\[endif\]-->\s*<!--\[if mso \| IE\]>\s*</tr>\s*</table>\s*<!\[endif\]-->`)
	normalized = msoClosePattern.ReplaceAllString(normalized, "<!--[if mso | IE]></td></tr></table><![endif]-->")

	// Ensure MSO tables include the empty class attribute MJML injects
	// so comparisons are done against the same attribute set.
	msoTableOpenPattern := regexp.MustCompile(`<!--\[if mso \| IE\]><table([^>]*)>`)
	normalized = msoTableOpenPattern.ReplaceAllStringFunc(normalized, func(match string) string {
		if strings.Contains(match, "class=") {
			return match
		}

		if idx := strings.Index(match, `cellspacing="0"`); idx != -1 {
			insertPos := idx + len(`cellspacing="0"`)
			return match[:insertPos] + ` class=""` + match[insertPos:]
		}

		if idx := strings.Index(match, `role="presentation"`); idx != -1 {
			return match[:idx] + ` class="" ` + match[idx:]
		}

		return strings.Replace(match, "<table", "<table class=\"\"", 1)
	})

	// Remove empty <style> tags that MRML used to inject but MJML omits.
	emptyStylePattern := regexp.MustCompile(`(?is)<style[^>]*>\s*</style>`)
	normalized = emptyStylePattern.ReplaceAllString(normalized, "")

	// Ensure root wrapper div contains the accessibility attributes MJML outputs
	rootDivPattern := regexp.MustCompile(`<body([^>]*)><div([^>]*)>`)
	normalized = rootDivPattern.ReplaceAllStringFunc(normalized, func(match string) string {
		submatches := rootDivPattern.FindStringSubmatch(match)
		if len(submatches) != 3 {
			return match
		}

		bodyAttrs := submatches[1]
		divAttrs := submatches[2]

		attrRe := regexp.MustCompile(`([a-zA-Z0-9:-]+)="([^"]*)"`)
		matches := attrRe.FindAllStringSubmatch(divAttrs, -1)
		attrMap := make(map[string]string, len(matches))
		keys := make([]string, 0, len(matches))
		for _, m := range matches {
			attrMap[m[1]] = m[2]
			keys = append(keys, m[1])
		}

		if _, exists := attrMap["aria-roledescription"]; !exists {
			attrMap["aria-roledescription"] = "email"
			keys = append(keys, "aria-roledescription")
		}
		if _, exists := attrMap["role"]; !exists {
			attrMap["role"] = "article"
			keys = append(keys, "role")
		}

		sort.Strings(keys)

		var b strings.Builder
		for _, key := range keys {
			b.WriteString(" ")
			b.WriteString(key)
			b.WriteString(`="`)
			b.WriteString(attrMap[key])
			b.WriteString(`"`)
		}

		return fmt.Sprintf("<body%s><div%s>", bodyAttrs, b.String())
	})

	// Normalize moustache templating markers so trailing whitespace produced by
	// legacy MRML fixtures doesn't cause mismatches. MJML trims raw content,
	// so remove any whitespace directly surrounding templating blocks when
	// they abut HTML tags or conditional comments.
	normalized = mustacheAfterClosingPattern.ReplaceAllString(normalized, "$1$2")
	normalized = mustacheBeforeOpeningPattern.ReplaceAllString(normalized, "$1$2")

	// Normalize viewport meta spacing differences (remove spaces after commas)
	viewportPattern := regexp.MustCompile(`(<meta[^>]*name="viewport"[^>]*content=")([^"]*)(")`)
	normalized = viewportPattern.ReplaceAllStringFunc(normalized, func(match string) string {
		submatches := viewportPattern.FindStringSubmatch(match)
		if len(submatches) != 4 {
			return match
		}

		cleaned := strings.ReplaceAll(submatches[2], ", ", ",")
		return submatches[1] + cleaned + submatches[3]
	})

	return normalized
}

// extractMSOSequences extracts sequences of MSO conditional comments and adjacent HTML elements
// for comparison of ordering differences that DOM parsing would normalize away
func extractMSOSequences(html string) []string {
	// Pattern to match MSO conditional blocks with their surrounding content
	// This captures MSO conditionals and the next few HTML elements following them
	re := regexp.MustCompile(`<!--\[if[^>]*>[\s\S]*?<!\[endif\]-->`)
	matches := re.FindAllString(html, -1)

	var sequences []string
	whitespace := regexp.MustCompile(`\s+`)
	for _, match := range matches {
		// Normalize whitespace for comparison and canonicalize attribute ordering
		normalized := whitespace.ReplaceAllString(strings.TrimSpace(match), " ")
		normalized = canonicalizeMSOBlock(normalized)
		if normalized != "" {
			sequences = append(sequences, normalized)
		}
	}

	return sequences
}

// canonicalizeMSOBlock sorts attributes inside MSO table/td tags so that
// serialization differences (like attribute ordering) do not trigger
// false-positive mismatches.
func canonicalizeMSOBlock(block string) string {
	canonical := canonicalizeTagAttributes(block, "table")
	canonical = canonicalizeTagAttributes(canonical, "td")
	return canonical
}

func canonicalizeTagAttributes(block, tag string) string {
	pattern := fmt.Sprintf(`<%s([^>]*)>`, tag)
	re := regexp.MustCompile(pattern)

	return re.ReplaceAllStringFunc(block, func(tagStr string) string {
		attrRe := regexp.MustCompile(`([a-zA-Z0-9:-]+)="([^"]*)"`)
		matches := attrRe.FindAllStringSubmatch(tagStr, -1)
		if len(matches) == 0 {
			return tagStr
		}

		attrMap := make(map[string]string, len(matches))
		keys := make([]string, 0, len(matches))
		for _, m := range matches {
			attrMap[m[1]] = m[2]
			keys = append(keys, m[1])
		}
		sort.Strings(keys)

		var b strings.Builder
		b.WriteString("<")
		b.WriteString(tag)
		for _, key := range keys {
			b.WriteString(" ")
			b.WriteString(key)
			b.WriteString(`="`)
			b.WriteString(attrMap[key])
			b.WriteString(`"`)
		}
		b.WriteString(">")
		return b.String()
	})
}
