package testutils

import (
	"fmt"
	"strings"
	"testing"
)

func TestStylesEqual(t *testing.T) {
	tests := []struct {
		name   string
		style1 string
		style2 string
		want   bool
	}{
		{
			name:   "identical strings",
			style1: "color: red; font-size: 12px;",
			style2: "color: red; font-size: 12px;",
			want:   true,
		},
		{
			name:   "different order same properties",
			style1: "color: red; font-size: 12px;",
			style2: "font-size: 12px; color: red;",
			want:   true,
		},
		{
			name:   "different whitespace",
			style1: "color:red;font-size:12px;",
			style2: "color: red; font-size: 12px;",
			want:   true,
		},
		{
			name:   "extra whitespace and semicolons",
			style1: "color: red;  font-size: 12px;;",
			style2: "font-size: 12px; color: red",
			want:   true,
		},
		{
			name:   "empty strings",
			style1: "",
			style2: "",
			want:   true,
		},
		{
			name:   "one empty one with content",
			style1: "",
			style2: "color: red;",
			want:   false,
		},
		{
			name:   "different values",
			style1: "color: red;",
			style2: "color: blue;",
			want:   false,
		},
		{
			name:   "different properties",
			style1: "color: red;",
			style2: "background: red;",
			want:   false,
		},
		{
			name:   "subset properties",
			style1: "color: red; font-size: 12px;",
			style2: "color: red;",
			want:   false,
		},
		{
			name:   "complex real-world example",
			style1: "font-size:0px;padding:20px;word-break:break-word;",
			style2: "font-size:0px;word-break:break-word;padding:20px;",
			want:   true,
		},
		{
			name:   "border radius different order",
			style1: "border:0;border-radius:10px;display:block;outline:none;text-decoration:none;height:auto;width:100%;font-size:13px;",
			style2: "border:0;display:block;outline:none;text-decoration:none;height:auto;width:100%;font-size:13px;border-radius:10px;",
			want:   true,
		},
		{
			name:   "malformed CSS ignored",
			style1: "color: red; invalid-no-colon; font-size: 12px;",
			style2: "color: red; font-size: 12px;",
			want:   true,
		},
		{
			name:   "properties with multiple colons",
			style1: "background: url('http://example.com/image.png');",
			style2: "background: url('http://example.com/image.png');",
			want:   true,
		},
		{
			name:   "case sensitive property names",
			style1: "Color: red;",
			style2: "color: red;",
			want:   false,
		},
		{
			name:   "case sensitive property values",
			style1: "color: RED;",
			style2: "color: red;",
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := StylesEqual(tt.style1, tt.style2)
			if got != tt.want {
				t.Errorf("StylesEqual(%q, %q) = %v, want %v", tt.style1, tt.style2, got, tt.want)
			}
		})
	}
}

func TestParseStyleProperties(t *testing.T) {
	tests := []struct {
		name  string
		style string
		want  map[string]string
	}{
		{
			name:  "empty string",
			style: "",
			want:  map[string]string{},
		},
		{
			name:  "single property",
			style: "color: red;",
			want:  map[string]string{"color": "red"},
		},
		{
			name:  "multiple properties",
			style: "color: red; font-size: 12px;",
			want:  map[string]string{"color": "red", "font-size": "12px"},
		},
		{
			name:  "no trailing semicolon",
			style: "color: red",
			want:  map[string]string{"color": "red"},
		},
		{
			name:  "extra whitespace",
			style: "  color : red ;  font-size : 12px  ;  ",
			want:  map[string]string{"color": "red", "font-size": "12px"},
		},
		{
			name:  "empty declarations",
			style: "color: red;; ; font-size: 12px;",
			want:  map[string]string{"color": "red", "font-size": "12px"},
		},
		{
			name:  "malformed declarations ignored",
			style: "color: red; invalid-no-colon; font-size: 12px;",
			want:  map[string]string{"color": "red", "font-size": "12px"},
		},
		{
			name:  "property with multiple colons",
			style: "background: url('http://example.com/image.png');",
			want:  map[string]string{"background": "url('http://example.com/image.png')"},
		},
		{
			name:  "empty property or value ignored",
			style: ": red; color:; color: blue;",
			want:  map[string]string{"color": "blue"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseStyleProperties(tt.style)
			if len(got) != len(tt.want) {
				t.Errorf("parseStyleProperties(%q) returned %d properties, want %d", tt.style, len(got), len(tt.want))
			}
			for prop, expectedValue := range tt.want {
				if actualValue, exists := got[prop]; !exists {
					t.Errorf("parseStyleProperties(%q) missing property %q", tt.style, prop)
				} else if actualValue != expectedValue {
					t.Errorf("parseStyleProperties(%q) property %q = %q, want %q", tt.style, prop, actualValue, expectedValue)
				}
			}
			for prop := range got {
				if _, exists := tt.want[prop]; !exists {
					t.Errorf("parseStyleProperties(%q) unexpected property %q = %q", tt.style, prop, got[prop])
				}
			}
		})
	}
}

// Benchmark tests to ensure performance is reasonable
func BenchmarkStylesEqual(b *testing.B) {
	style1 := "font-size:0px;padding:20px;word-break:break-word;color:red;background:white;border:1px solid black;"
	style2 := "color:red;font-size:0px;border:1px solid black;word-break:break-word;background:white;padding:20px;"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		StylesEqual(style1, style2)
	}
}

func BenchmarkParseStyleProperties(b *testing.B) {
	style := "font-size:0px;padding:20px;word-break:break-word;color:red;background:white;border:1px solid black;"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		parseStyleProperties(style)
	}
}

// TestCompareStylesPreciseConsistency verifies that CompareStylesPrecise and StylesEqual
// behave consistently - if StylesEqual returns true, CompareStylesPrecise should not log any differences
func TestCompareStylesPreciseConsistency(t *testing.T) {
	testCases := []struct {
		name        string
		expected    string
		actual      string
		stylesEqual bool
	}{
		{
			name:        "identical styles",
			expected:    `<div style="color: red; font-size: 12px;">content</div>`,
			actual:      `<div style="color: red; font-size: 12px;">content</div>`,
			stylesEqual: true,
		},
		{
			name:        "reordered CSS properties",
			expected:    `<div style="color: red; font-size: 12px;">content</div>`,
			actual:      `<div style="font-size: 12px; color: red;">content</div>`,
			stylesEqual: true,
		},
		{
			name:        "different whitespace in styles",
			expected:    `<div style="color:red;font-size:12px;">content</div>`,
			actual:      `<div style="color: red; font-size: 12px;">content</div>`,
			stylesEqual: true,
		},
		{
			name:        "extra semicolons and whitespace",
			expected:    `<div style="color: red;  font-size: 12px;;">content</div>`,
			actual:      `<div style="font-size: 12px; color: red">content</div>`,
			stylesEqual: true,
		},
		{
			name:        "different style values should differ",
			expected:    `<div style="color: red;">content</div>`,
			actual:      `<div style="color: blue;">content</div>`,
			stylesEqual: false,
		},
		{
			name:        "missing properties should differ",
			expected:    `<div style="color: red; font-size: 12px;">content</div>`,
			actual:      `<div style="color: red;">content</div>`,
			stylesEqual: false,
		},
		{
			name:        "complex real-world styles",
			expected:    `<table style="font-size:0px;padding:20px;word-break:break-word;"><tr><td style="border:0;display:block;outline:none;text-decoration:none;height:auto;width:100%;font-size:13px;">content</td></tr></table>`,
			actual:      `<table style="font-size:0px;word-break:break-word;padding:20px;"><tr><td style="border:0;display:block;outline:none;text-decoration:none;height:auto;width:100%;font-size:13px;">content</td></tr></table>`,
			stylesEqual: true,
		},
		{
			name: "multiple elements with mixed differences",
			expected: `<div style="color: red;">
				<p style="font-size: 12px; margin: 10px;">text</p>
				<span style="background: white;">span</span>
			</div>`,
			actual: `<div style="color: red;">
				<p style="margin: 10px; font-size: 12px;">text</p>
				<span style="background: black;">span</span>
			</div>`,
			stylesEqual: false, // span has different background
		},
		{
			name:        "no styled elements",
			expected:    `<div>content</div>`,
			actual:      `<div>content</div>`,
			stylesEqual: true,
		},
		{
			name:        "empty style attributes",
			expected:    `<div style="">content</div>`,
			actual:      `<div style="">content</div>`,
			stylesEqual: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Run CompareStylesPrecise and get structured result
			result := CompareStylesPrecise(tc.expected, tc.actual)

			if result.ParseError != nil {
				t.Fatalf("Failed to parse HTML: %v", result.ParseError)
			}

			// Check consistency between CompareStylesPrecise and StylesEqual for each element
			var inconsistencies []string

			for _, element := range result.Elements {
				// Only check elements that have both expected and actual styles
				if element.Status == ElementIdentical || element.Status == ElementDifferent {
					stylesEqualResult := StylesEqual(element.Expected, element.Actual)
					compareStylesPreciseFoundDiff := (element.Status == ElementDifferent)

					// If StylesEqual says they're equal, CompareStylesPrecise should mark as identical
					if stylesEqualResult && compareStylesPreciseFoundDiff {
						inconsistencies = append(inconsistencies,
							fmt.Sprintf("Element[%d]: StylesEqual returned true but CompareStylesPrecise found differences. Expected: %q, Actual: %q",
								element.Index, element.Expected, element.Actual))
					}

					// If StylesEqual says they differ, CompareStylesPrecise should mark as different
					if !stylesEqualResult && !compareStylesPreciseFoundDiff {
						inconsistencies = append(inconsistencies,
							fmt.Sprintf("Element[%d]: StylesEqual returned false but CompareStylesPrecise found no differences. Expected: %q, Actual: %q",
								element.Index, element.Expected, element.Actual))
					}
				}
			}

			if len(inconsistencies) > 0 {
				t.Errorf("Inconsistencies between StylesEqual and CompareStylesPrecise:\n%s", strings.Join(inconsistencies, "\n"))
				t.Logf("CompareStylesPrecise result: HasDifferences=%v, Elements=%d", result.HasDifferences, len(result.Elements))
				for i, elem := range result.Elements {
					t.Logf("  Element[%d]: Status=%d, Expected=%q, Actual=%q", i, elem.Status, elem.Expected, elem.Actual)
				}
			}
		})
	}
}
