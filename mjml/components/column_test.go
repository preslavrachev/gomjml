package components

import (
	"encoding/xml"
	"testing"

	"github.com/preslavrachev/gomjml/mjml/options"
	"github.com/preslavrachev/gomjml/parser"
)

// TestGetColumnClass tests the GetColumnClass method which generates CSS class names
// based on column width calculations. The siblings count represents the total number
// of columns within the same mj-section, used for automatic width distribution when
// no explicit width is specified.
//
// Example MJML structure:
//
//	<mj-section>
//	  <mj-column>Column 1</mj-column>  // siblings=3, gets 33.33% width
//	  <mj-column>Column 2</mj-column>  // siblings=3, gets 33.33% width
//	  <mj-column>Column 3</mj-column>  // siblings=3, gets 33.33% width
//	</mj-section>
//
// Width calculation: 100% / (siblings - rawSiblings) = auto width per column
// Class format: mj-column-per-{width} where dots become dashes
//
// CRITICAL: Precision must match MRML's Rust f32 (32-bit float) behavior!
// Go's default float64 produces different results for fractional percentages:
//   - Go float64: 100/3 = 33.333333333333336 -> "mj-column-per-33-333333333333336"
//   - Rust f32:   100/3 = 33.333332          -> "mj-column-per-33-333332"
//
// Solution: Convert to float32 and use %g formatting to match MRML exactly.
func TestGetColumnClass(t *testing.T) {
	tests := []struct {
		name          string
		width         string // explicit width attribute (empty = auto-calculate)
		siblings      int    // total number of columns in the section
		rawSiblings   int    // number of raw content siblings (text nodes, etc.)
		expectedClass string // expected CSS class name
	}{
		{
			name:          "1 column - 100%",
			width:         "",
			siblings:      1,
			rawSiblings:   0,
			expectedClass: "mj-column-per-100",
		},
		{
			name:          "2 columns - 50%",
			width:         "",
			siblings:      2,
			rawSiblings:   0,
			expectedClass: "mj-column-per-50",
		},
		{
			name:          "3 columns - 33.333332%",
			width:         "",
			siblings:      3,
			rawSiblings:   0,
			expectedClass: "mj-column-per-33-333332",
		},
		{
			name:          "4 columns - 25%",
			width:         "",
			siblings:      4,
			rawSiblings:   0,
			expectedClass: "mj-column-per-25",
		},
		{
			name:          "5 columns - 20%",
			width:         "",
			siblings:      5,
			rawSiblings:   0,
			expectedClass: "mj-column-per-20",
		},
		{
			name:          "6 columns - 16.666666%",
			width:         "",
			siblings:      6,
			rawSiblings:   0,
			expectedClass: "mj-column-per-16-666666",
		},
		{
			name:          "7 columns - 14.285714%",
			width:         "",
			siblings:      7,
			rawSiblings:   0,
			expectedClass: "mj-column-per-14-285714",
		},
		{
			name:          "8 columns - 12.5%",
			width:         "",
			siblings:      8,
			rawSiblings:   0,
			expectedClass: "mj-column-per-12-5",
		},
		{
			name:          "9 columns - 11.111111%",
			width:         "",
			siblings:      9,
			rawSiblings:   0,
			expectedClass: "mj-column-per-11-111111",
		},
		{
			name:          "10 columns - 10%",
			width:         "",
			siblings:      10,
			rawSiblings:   0,
			expectedClass: "mj-column-per-10",
		},
		// Explicit width test cases - tests our strconv.FormatFloat fix
		{
			name:          "Explicit 33.333% width",
			width:         "33.333%",
			siblings:      1, // siblings don't matter for explicit width
			rawSiblings:   0,
			expectedClass: "mj-column-per-33-333",
		},
		{
			name:          "Explicit 25.5% width",
			width:         "25.5%",
			siblings:      1,
			rawSiblings:   0,
			expectedClass: "mj-column-per-25-5",
		},
		{
			name:          "Explicit 16.666666% width (f32 precision test)",
			width:         "16.666666%",
			siblings:      1,
			rawSiblings:   0,
			expectedClass: "mj-column-per-16-666666",
		},
		{
			name:          "Explicit 12.25% width",
			width:         "12.25%",
			siblings:      1,
			rawSiblings:   0,
			expectedClass: "mj-column-per-12-25",
		},
		{
			name:          "Explicit 200px width",
			width:         "200px",
			siblings:      1,
			rawSiblings:   0,
			expectedClass: "mj-column-px-200",
		},
		{
			name:          "Explicit 150px width",
			width:         "150px",
			siblings:      1,
			rawSiblings:   0,
			expectedClass: "mj-column-px-150",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock MJMLNode
			node := &parser.MJMLNode{
				Attrs: []xml.Attr{},
			}

			// Add width attribute if specified
			if tt.width != "" {
				node.Attrs = append(node.Attrs, xml.Attr{
					Name:  xml.Name{Local: "width"},
					Value: tt.width,
				})
			}

			// Create component
			component := NewMJColumnComponent(node, &options.RenderOpts{})

			// Set siblings to simulate the column count scenario
			component.SetSiblings(tt.siblings)
			component.SetRawSiblings(tt.rawSiblings)

			// Test GetColumnClass
			className, _ := component.GetColumnClass()

			if className != tt.expectedClass {
				t.Errorf("GetColumnClass() = %q, expected %q", className, tt.expectedClass)
			}
		})
	}
}
