package mjml

import (
	"fmt"
	"math/rand"
	"strings"
	"testing"
	"time"

	"github.com/PuerkitoBio/goquery"
)

func TestGroupChildrenWidthCalculation(t *testing.T) {
	// Seed random number generator
	rand.Seed(time.Now().UnixNano())

	// Test with 7 columns to check precise decimal CSS class generation

	columnCount := rand.Intn(20) + 1

	t.Run(fmt.Sprintf("with_%d_columns", columnCount), func(t *testing.T) {
		// Build MJML string with the specified number of columns
		var mjmlBuilder strings.Builder
		mjmlBuilder.WriteString(`<mjml><mj-body><mj-section><mj-group>`)

		for j := 1; j <= columnCount; j++ {
			mjmlBuilder.WriteString(fmt.Sprintf(`<mj-column><mj-text>Column %d</mj-text></mj-column>`, j))
		}

		mjmlBuilder.WriteString(`</mj-group></mj-section></mj-body></mjml>`)
		mjmlInput := mjmlBuilder.String()

		// Render the MJML
		htmlOutput, err := Render(mjmlInput)
		if err != nil {
			t.Fatalf("Failed to render MJML: %v", err)
		}

		// Calculate expected percentage per column with decimal precision
		expectedPercentage := 100.0 / float64(columnCount)

		// Generate the precise CSS class name with decimal digits
		// Format: mj-column-per-{integer}-{decimal_digits}
		integerPart := int(expectedPercentage)
		decimalPart := expectedPercentage - float64(integerPart)

		var expectedCSSClass string
		if decimalPart == 0 {
			// No decimal part (e.g., 2 columns = 50%)
			expectedCSSClass = fmt.Sprintf("mj-column-per-%d", integerPart)
		} else {
			// With decimal part (e.g., 7 columns = 14.285714285714286%)
			decimalString := fmt.Sprintf("%.15f", decimalPart)[2:] // Remove "0."
			decimalString = strings.TrimRight(decimalString, "0")  // Remove trailing zeros
			expectedCSSClass = fmt.Sprintf("mj-column-per-%d-%s", integerPart, decimalString)
		}

		t.Logf("Testing with %d columns, expecting %.2f%% per column (%s)",
			columnCount, expectedPercentage, expectedCSSClass)

		// Verify CSS class appears in the HTML elements
		if !strings.Contains(htmlOutput, expectedCSSClass) {
			// Extract actual CSS classes that were found
			actualClasses := extractCSSClasses(htmlOutput)
			t.Errorf(
				"Expected CSS class '%s' not found in HTML output. Found classes: %v",
				expectedCSSClass,
				actualClasses,
			)
			t.Logf("HTML output snippet: %s", htmlOutput[:min(1000, len(htmlOutput))])
		}

		// Count occurrences of the CSS class in div elements
		expectedOccurrences := columnCount // Each column should have this class
		actualOccurrences := strings.Count(htmlOutput, fmt.Sprintf(`class="mj-outlook-group-fix %s"`, expectedCSSClass))

		if actualOccurrences != expectedOccurrences {
			t.Errorf("Expected %d occurrences of CSS class '%s' in div elements, found %d",
				expectedOccurrences, expectedCSSClass, actualOccurrences)
		}

		// Check that at least one <style> block contains the expected CSS class
		// Use goquery to parse the HTML and select <style> blocks
		doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlOutput))
		if err != nil {
			t.Fatalf("Failed to parse HTML output with goquery: %v", err)
		}
		foundInStyle := false
		doc.Find("style").Each(func(i int, s *goquery.Selection) {
			styleText := s.Text()
			if strings.Contains(styleText, expectedCSSClass) {
				foundInStyle = true
			}
		})
		if !foundInStyle {
			t.Errorf("Expected CSS class '%s' not found in any <style> block", expectedCSSClass)
		}
	})
}

// Helper function for Go versions that don't have built-in min
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// extractCSSClasses finds all mj-column-per-* classes in the HTML output
func extractCSSClasses(html string) []string {
	var classes []string

	// Look for class="..." patterns and extract mj-column-per-* classes
	lines := strings.Split(html, "\n")
	for _, line := range lines {
		if strings.Contains(line, "class=") && strings.Contains(line, "mj-column-per-") {
			// Extract the class attribute value
			start := strings.Index(line, `class="`)
			if start != -1 {
				start += 7 // Move past 'class="'
				end := strings.Index(line[start:], `"`)
				if end != -1 {
					classAttr := line[start : start+end]
					// Split by spaces and find mj-column-per-* classes
					for _, class := range strings.Fields(classAttr) {
						if strings.HasPrefix(class, "mj-column-per-") {
							classes = append(classes, class)
						}
					}
				}
			}
		}
	}

	return classes
}
