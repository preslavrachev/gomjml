package testutils

import (
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// StylesEqual compares two CSS style strings semantically, ignoring property order
// and whitespace differences. It parses both styles into property maps and compares
// them for equality.
//
// Returns true if both styles represent the same set of CSS properties and values.
//
// Examples:
//
//	StylesEqual("color: red; font-size: 12px;", "font-size: 12px; color: red;") // true
//	StylesEqual("padding: 10px", "padding:10px") // true
//	StylesEqual("", "") // true
//	StylesEqual("color: red", "color: blue") // false
func StylesEqual(style1, style2 string) bool {
	if style1 == style2 {
		return true
	}

	props1 := parseStyleProperties(style1)
	props2 := parseStyleProperties(style2)

	if len(props1) != len(props2) {
		return false
	}

	for prop, value1 := range props1 {
		if value2, exists := props2[prop]; !exists || value1 != value2 {
			return false
		}
	}

	return true
}

// parseStyleProperties parses a CSS style string into a map of property-value pairs.
// Handles standard CSS declaration format: "property: value; property2: value2;"
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

		// Split by colon to separate property and value
		parts := strings.SplitN(decl, ":", 2)
		if len(parts) != 2 {
			continue
		}

		prop := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		if prop != "" && value != "" {
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

// StyleComparisonResult contains detailed style comparison results
type StyleComparisonResult struct {
	HasDifferences bool
	Elements       []ElementComparison
	ParseError     error
}

// ElementComparison represents style differences for a single element
type ElementComparison struct {
	Index     int
	Tag       string
	Classes   string
	Component string
	Status    ElementStatus
	Expected  string
	Actual    string
	StyleDiff StyleDiff
}

// ElementStatus indicates the comparison status
type ElementStatus int

const (
	ElementIdentical ElementStatus = iota
	ElementMissing
	ElementExtra
	ElementDifferent
)

// CompareStylesPrecise provides exact element identification for style differences
func CompareStylesPrecise(expected, actual string) StyleComparisonResult {
	expectedDoc, err1 := goquery.NewDocumentFromReader(strings.NewReader(expected))
	actualDoc, err2 := goquery.NewDocumentFromReader(strings.NewReader(actual))

	if err1 != nil || err2 != nil {
		return StyleComparisonResult{
			HasDifferences: true,
			ParseError:     fmt.Errorf("DOM parsing failed: expected=%v, actual=%v", err1, err2),
		}
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

	// Compare element by element and build results
	result := StyleComparisonResult{
		HasDifferences: false,
		Elements:       make([]ElementComparison, 0),
	}

	maxLen := max(len(expectedElements), len(actualElements))
	for i := 0; i < maxLen; i++ {
		var expected, actual *ElementInfo
		if i < len(expectedElements) {
			expected = &expectedElements[i]
		}
		if i < len(actualElements) {
			actual = &actualElements[i]
		}

		comparison := ElementComparison{Index: i}

		if expected == nil {
			// Extra element in actual
			comparison.Status = ElementExtra
			comparison.Tag = actual.Tag
			comparison.Classes = actual.Classes
			comparison.Component = actual.Component
			comparison.Actual = actual.Style
			result.HasDifferences = true
		} else if actual == nil {
			// Missing element in actual
			comparison.Status = ElementMissing
			comparison.Tag = expected.Tag
			comparison.Classes = expected.Classes
			comparison.Expected = expected.Style
			result.HasDifferences = true
		} else if !StylesEqual(expected.Style, actual.Style) {
			// Element exists but styles differ
			comparison.Status = ElementDifferent
			comparison.Tag = actual.Tag
			comparison.Classes = actual.Classes
			comparison.Component = actual.Component
			comparison.Expected = expected.Style
			comparison.Actual = actual.Style

			// Calculate detailed style differences
			expectedProps := parseStyleProperties(expected.Style)
			actualProps := parseStyleProperties(actual.Style)
			comparison.StyleDiff = compareStylePropertiesMaps(expectedProps, actualProps)
			result.HasDifferences = true
		} else {
			// Styles are identical
			comparison.Status = ElementIdentical
			comparison.Tag = actual.Tag
			comparison.Classes = actual.Classes
			comparison.Component = actual.Component
			comparison.Expected = expected.Style
			comparison.Actual = actual.Style
		}

		result.Elements = append(result.Elements, comparison)
	}

	return result
}

// ElementInfo represents a styled HTML element
type ElementInfo struct {
	Tag       string
	Classes   string
	Style     string
	Index     int
	Component string // Which MJML component created this element (from debug attrs)
}
