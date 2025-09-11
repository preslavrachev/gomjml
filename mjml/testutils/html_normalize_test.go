package testutils

import (
	"strings"
	"testing"
)

func TestNormalizeHTMLAttributes(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple attribute reordering",
			input:    `<div class="foo" id="bar">`,
			expected: `<div class="foo" id="bar">`,
		},
		{
			name:     "reverse attribute order",
			input:    `<div id="bar" class="foo">`,
			expected: `<div class="foo" id="bar">`,
		},
		{
			name:     "table tag with multiple attributes - original order",
			input:    `<table border="0" cellpadding="0" cellspacing="0" role="presentation" bgcolor="#ffffff" align="center" width="560" style="width:560px;">`,
			expected: `<table align="center" bgcolor="#ffffff" border="0" cellpadding="0" cellspacing="0" role="presentation" style="width:560px;" width="560">`,
		},
		{
			name:     "table tag with multiple attributes - different order",
			input:    `<table border="0" cellpadding="0" cellspacing="0" role="presentation" align="center" width="560" bgcolor="#ffffff" style="width:560px;">`,
			expected: `<table align="center" bgcolor="#ffffff" border="0" cellpadding="0" cellspacing="0" role="presentation" style="width:560px;" width="560">`,
		},
		{
			name:     "self-closing tag",
			input:    `<img alt="test" src="image.jpg" />`,
			expected: `<img alt="test" src="image.jpg" />`,
		},
		{
			name:     "self-closing tag reverse order",
			input:    `<img src="image.jpg" alt="test" />`,
			expected: `<img alt="test" src="image.jpg" />`,
		},
		{
			name:     "tag with no attributes",
			input:    `<div>`,
			expected: `<div>`,
		},
		{
			name:     "multiple tags in one line",
			input:    `<div id="outer" class="container"><span class="inner" id="text">content</span></div>`,
			expected: `<div class="container" id="outer"><span class="inner" id="text">content</span></div>`,
		},
		{
			name:     "attributes with spaces in values",
			input:    `<div style="color: red; font-size: 14px" class="test">`,
			expected: `<div class="test" style="color:red;font-size:14px">`,
		},
		{
			name:     "mixed quote types",
			input:    `<div id='single' class="double">`,
			expected: `<div class="double" id='single'>`,
		},
		{
			name:     "complex style attribute with CSS reordering",
			input:    `<td style="padding:10px;background:#fff;border:1px solid #ccc" class="cell">`,
			expected: `<td class="cell" style="background:#fff;border:1px solid #ccc;padding:10px">`,
		},
		{
			name:     "line with no HTML tags",
			input:    `This is plain text without any HTML tags.`,
			expected: `This is plain text without any HTML tags.`,
		},
		{
			name:     "empty string",
			input:    ``,
			expected: ``,
		},
		{
			name:     "malformed tag (no closing bracket)",
			input:    `<div class="test"`,
			expected: `<div class="test"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizeHTMLAttributes(tt.input)
			if result != tt.expected {
				t.Errorf("NormalizeHTMLAttributes() = %q, expected %q", result, tt.expected)
			}
		})
	}
}

func TestNormalizeHTMLAttributesIdenticalResults(t *testing.T) {
	// Test cases where different input should produce identical normalized output
	identicalTests := []struct {
		name   string
		input1 string
		input2 string
	}{
		{
			name:   "table attributes different order",
			input1: `<table border="0" cellpadding="0" cellspacing="0" role="presentation" bgcolor="#ffffff" align="center" width="560" style="width:560px;">`,
			input2: `<table border="0" cellpadding="0" cellspacing="0" role="presentation" align="center" width="560" bgcolor="#ffffff" style="width:560px;">`,
		},
		{
			name:   "div attributes reversed",
			input1: `<div class="container" id="main" style="color:red">`,
			input2: `<div style="color:red" id="main" class="container">`,
		},
		{
			name:   "img tag attributes shuffled",
			input1: `<img src="test.jpg" alt="Test" width="100" height="50" />`,
			input2: `<img height="50" alt="Test" src="test.jpg" width="100" />`,
		},
		{
			name:   "multiple tags with reordered attributes",
			input1: `<div class="outer" id="container"><span role="text" class="inner">text</span></div>`,
			input2: `<div id="container" class="outer"><span class="inner" role="text">text</span></div>`,
		},
	}

	for _, tt := range identicalTests {
		t.Run(tt.name, func(t *testing.T) {
			result1 := NormalizeHTMLAttributes(tt.input1)
			result2 := NormalizeHTMLAttributes(tt.input2)
			if result1 != result2 {
				t.Errorf("NormalizeHTMLAttributes() produced different results:\n  input1: %q -> %q\n  input2: %q -> %q",
					tt.input1, result1, tt.input2, result2)
			}
		})
	}
}

func TestParseHTMLTagContent(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "simple tag with attributes",
			input:    `div class="test" id="main"`,
			expected: []string{"div", `class="test"`, `id="main"`},
		},
		{
			name:     "tag name only",
			input:    `div`,
			expected: []string{"div"},
		},
		{
			name:     "attributes with spaces in values",
			input:    `span style="color: red; font-size: 14px" class="text"`,
			expected: []string{"span", `style="color: red; font-size: 14px"`, `class="text"`},
		},
		{
			name:     "mixed quote types",
			input:    `input type="text" name='username' placeholder="Enter username"`,
			expected: []string{"input", `type="text"`, `name='username'`, `placeholder="Enter username"`},
		},
		{
			name:     "extra spaces",
			input:    `  div   class="test"    id="main"  `,
			expected: []string{"div", `class="test"`, `id="main"`},
		},
		{
			name:     "empty input",
			input:    ``,
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseHTMLTagContent(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("parseHTMLTagContent() returned %d parts, expected %d: %v vs %v",
					len(result), len(tt.expected), result, tt.expected)
				return
			}
			for i, part := range result {
				if part != tt.expected[i] {
					t.Errorf("parseHTMLTagContent()[%d] = %q, expected %q", i, part, tt.expected[i])
				}
			}
		})
	}
}

func TestNormalizeCSSProperties(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "basic CSS properties",
			input:    `color:red;font-size:14px`,
			expected: `color:red;font-size:14px`,
		},
		{
			name:     "reverse order CSS properties",
			input:    `font-size:14px;color:red`,
			expected: `color:red;font-size:14px`,
		},
		{
			name:     "CSS with spaces",
			input:    `padding: 10px; margin: 5px; background: #fff`,
			expected: `background:#fff;margin:5px;padding:10px`,
		},
		{
			name:     "CSS with trailing semicolon",
			input:    `color:red;font-size:14px;`,
			expected: `color:red;font-size:14px;`,
		},
		{
			name:     "CSS with quoted wrapper",
			input:    `"color:red;font-size:14px"`,
			expected: `color:red;font-size:14px`,
		},
		{
			name:     "empty CSS",
			input:    ``,
			expected: ``,
		},
		{
			name:     "single property",
			input:    `color:red`,
			expected: `color:red`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizeCSSProperties(tt.input)
			if result != tt.expected {
				t.Errorf("NormalizeCSSProperties() = %q, expected %q", result, tt.expected)
			}
		})
	}
}

func TestSortTagAttributes(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "basic opening tag",
			input:    `<div id="test" class="container">`,
			expected: `<div class="container" id="test">`,
		},
		{
			name:     "self-closing tag",
			input:    `<img src="test.jpg" alt="Test" />`,
			expected: `<img alt="Test" src="test.jpg" />`,
		},
		{
			name:     "tag with no attributes",
			input:    `<div>`,
			expected: `<div>`,
		},
		{
			name:     "self-closing tag no attributes",
			input:    `<br />`,
			expected: `<br />`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sortTagAttributes(tt.input)
			if result != tt.expected {
				t.Errorf("sortTagAttributes() = %q, expected %q", result, tt.expected)
			}
		})
	}
}

// Benchmark tests
func BenchmarkNormalizeHTMLAttributes(b *testing.B) {
	input := `<table border="0" cellpadding="0" cellspacing="0" role="presentation" bgcolor="#ffffff" align="center" width="560" style="width:560px;">`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		NormalizeHTMLAttributes(input)
	}
}

func BenchmarkParseHTMLTagContent(b *testing.B) {
	input := `div class="container" id="main" style="color: red; font-size: 14px" data-value="test"`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		parseHTMLTagContent(input)
	}
}

// Test the exact false positive case from the issue
func TestSpecificFalsePositiveCases(t *testing.T) {
	tests := []struct {
		name  string
		line1 string
		line2 string
	}{
		{
			name:  "HTML attribute ordering",
			line1: `<table border="0" cellpadding="0" cellspacing="0" role="presentation" bgcolor="#ffffff" align="center" width="560" style="width:560px;">`,
			line2: `<table border="0" cellpadding="0" cellspacing="0" role="presentation" align="center" width="560" bgcolor="#ffffff" style="width:560px;">`,
		},
		{
			name:  "CSS property ordering",
			line1: `<img height="auto" src="https://placehold.co/150x60/E3F2FD/1976D2?font=playfair-display&text=LOGO" width="150" style="border:none;border-radius:0px;display:block;outline:none;text-decoration:none;height:auto;width:100%;font-size:13px;" />`,
			line2: `<img height="auto" src="https://placehold.co/150x60/E3F2FD/1976D2?font=playfair-display&text=LOGO" width="150" style="border:none;display:block;outline:none;text-decoration:none;height:auto;width:100%;font-size:13px;border-radius:0px;" />`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			normalized1 := NormalizeHTMLAttributes(tt.line1)
			normalized2 := NormalizeHTMLAttributes(tt.line2)

			if normalized1 != normalized2 {
				t.Errorf("False positive case still failing:\n  line1 -> %q\n  line2 -> %q", normalized1, normalized2)
			}
		})
	}
}

func TestEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "malformed HTML - unclosed quote",
			input:    `<div class="test id="main">`,
			expected: `<div class="test id="main">`, // Should not crash, preserve as-is
		},
		{
			name:     "HTML entities in attributes",
			input:    `<div title="Tom &amp; Jerry" class="cartoon">`,
			expected: `<div class="cartoon" title="Tom &amp; Jerry">`,
		},
		{
			name:     "empty attribute value",
			input:    `<input type="text" value="" required>`,
			expected: `<input required type="text" value="">`,
		},
		{
			name:     "boolean attributes",
			input:    `<input type="checkbox" checked disabled>`,
			expected: `<input checked disabled type="checkbox">`,
		},
		{
			name:     "very long attribute value",
			input:    `<div data-config="` + strings.Repeat("x", 1000) + `" class="test">`,
			expected: `<div class="test" data-config="` + strings.Repeat("x", 1000) + `">`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizeHTMLAttributes(tt.input)
			if result != tt.expected {
				t.Errorf("NormalizeHTMLAttributes() = %q, expected %q", result, tt.expected)
			}
		})
	}
}
