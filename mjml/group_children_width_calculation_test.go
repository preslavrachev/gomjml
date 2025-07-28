package mjml

import (
	"fmt"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/preslavrachev/gomjml/mjml/components"
)

func TestGroupChildrenWidthCalculation(t *testing.T) {
	for columnCount := 1; columnCount <= 10; columnCount++ {
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
			renderResult, err := RenderWithAST(mjmlInput)
			if err != nil {
				t.Fatalf("Failed to render MJML: %v", err)
			}
			htmlOutput := renderResult.HTML

			// Extract expected CSS class from the AST instead of manually calculating
			expectedClasses := extractColumnClassesFromAST(renderResult.AST)
			if len(expectedClasses) == 0 {
				t.Fatalf("No CSS classes found in AST for %d columns", columnCount)
			}
			expectedCSSClass := expectedClasses[0] // Should be only one class for a single group

			// Calculate expected percentage for validation
			// expectedPercentage := 100.0 / float64(columnCount)

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
			actualOccurrences := strings.Count(
				htmlOutput,
				fmt.Sprintf(`class="mj-outlook-group-fix %s"`, expectedCSSClass),
			)

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

			// // CRITICAL: Validate inline style width values (the actual layout-controlling CSS)
			// // Use 'g' format to avoid trailing zeros (e.g., 12.5 instead of 12.500000000000000)
			// expectedWidthPercent := fmt.Sprintf("width:%s%%", strconv.FormatFloat(expectedPercentage, 'g', -1, 64))

			// columnDivs := doc.Find(fmt.Sprintf("div.%s", expectedCSSClass))
			// if columnDivs.Length() != columnCount {
			// 	t.Errorf("Expected %d column divs with class '%s', found %d",
			// 		columnCount, expectedCSSClass, columnDivs.Length())
			// }

			// // Validate each column div has the correct inline width style
			// columnsWithWrongWidth := 0
			// columnDivs.Each(func(i int, s *goquery.Selection) {
			// 	styleAttr, exists := s.Attr("style")
			// 	if !exists {
			// 		t.Errorf("Column div %d missing style attribute", i)
			// 		return
			// 	}

			// 	// Check if the style contains the expected width percentage
			// 	if !strings.Contains(styleAttr, expectedWidthPercent) {
			// 		columnsWithWrongWidth++
			// 		t.Errorf("Column div %d has incorrect width in style attribute. Expected '%s' in: %s",
			// 			i, expectedWidthPercent, styleAttr)
			// 	}
			// })

			// if columnsWithWrongWidth > 0 {
			// 	t.Errorf("Found %d columns with incorrect width styling out of %d total columns",
			// 		columnsWithWrongWidth, columnCount)
			// }
		})
	}
}

// Helper function for Go versions that don't have built-in min
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// extractColumnClassesFromAST extracts the CSS class names that will be generated for columns in groups
func extractColumnClassesFromAST(ast *MJMLNode) []string {
	var classes []string

	// Find the body and traverse to find groups
	body := ast.FindFirstChild("mj-body")
	if body == nil {
		return classes
	}

	// Look for sections
	for _, section := range body.FindAllChildren("mj-section") {
		// Look for groups within sections
		for _, group := range section.FindAllChildren("mj-group") {
			// Get column children in this group
			columns := group.FindAllChildren("mj-column")
			columnCount := len(columns)

			if columnCount > 0 {
				// Calculate percentage per column (same as group component does)
				percentagePerColumn := 100.0 / float64(columnCount)

				for _, columnNode := range columns {
					// Create a temporary column component from the AST node
					columnComponent, err := CreateComponent(columnNode, &RenderOpts{})
					if err != nil {
						continue
					}

					// Cast to MJColumnComponent and simulate what group does
					if columnComp, ok := columnComponent.(*components.MJColumnComponent); ok {
						// Set the width attribute like the group component does
						if columnComp.GetAttribute("width") == nil {
							// Use the same logic as group component
							percentageWidth := fmt.Sprintf("%.15f%%", percentagePerColumn)
							percentageWidth = strings.TrimRight(percentageWidth, "0")
							percentageWidth = strings.TrimRight(percentageWidth, ".")
							if !strings.HasSuffix(percentageWidth, "%") {
								percentageWidth += "%"
							}
							columnComp.Attrs["width"] = percentageWidth
						}

						// Now get the class name
						className, _ := columnComp.GetColumnClass()
						if className != "" {
							classes = append(classes, className)
						}
					}
				}
			}
		}
	}

	return classes
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
