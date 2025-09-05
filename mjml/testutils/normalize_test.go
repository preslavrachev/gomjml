package testutils

import (
	"testing"
)

func TestNormalizeForComparison(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "basic_whitespace_between_tags",
			input:    "<div>  <span>text</span>  </div>",
			expected: "<div><span>text</span></div>",
		},
		{
			name:     "multiple_whitespace_normalization",
			input:    "<div>text   with    multiple     spaces</div>",
			expected: "<div>text with multiple spaces</div>",
		},
		{
			name:     "whitespace_only_content_removal",
			input:    "<div>   </div><span>content</span>",
			expected: "<div></div><span>content</span>",
		},
		{
			name:     "preserve_entities",
			input:    "<div>&nbsp;&amp;&lt;&gt;</div>",
			expected: "<div>&nbsp;&amp;&lt;&gt;</div>",
		},
		{
			name:     "css_style_sorting",
			input:    `<div style="color: red; padding: 10px; margin: 5px">content</div>`,
			expected: `<div style="color:red;margin:5px;padding:10px">content</div>`,
		},
		{
			name:     "css_style_whitespace_normalization",
			input:    `<div style="color :  red  ;  padding  :  10px ">content</div>`,
			expected: `<div style="color:red;padding:10px">content</div>`,
		},
		{
			name:     "empty_style_attribute",
			input:    `<div style="">content</div>`,
			expected: `<div style="">content</div>`,
		},
		{
			name:     "style_with_semicolon_at_end",
			input:    `<div style="color: red; padding: 10px;">content</div>`,
			expected: `<div style="color:red;padding:10px">content</div>`,
		},
		{
			name: "complex_html_structure",
			input: `<!doctype html>
<html>
  <head>
    <title>Test</title>
  </head>
  <body style="margin: 0; padding: 10px">
    <div   style="color: blue;  font-size:  14px"  >
      Content with   multiple   spaces
    </div>
  </body>
</html>`,
			expected: `<!doctype html><html><head><title>Test</title></head><body style="margin:0;padding:10px"><div style="color:blue;font-size:14px" >Content with multiple spaces</div></body></html>`,
		},
		{
			name:     "newlines_and_tabs_normalization",
			input:    "<div>\n\t<span>\n\t\tcontent\n\t</span>\n</div>",
			expected: "<div><span>content</span></div>",
		},
		{
			name:     "social_media_href_preservation",
			input:    `<a href="https://twitter.com/handle" style="text-decoration: none; color: blue">Twitter</a>`,
			expected: `<a href="https://twitter.com/handle" style="color:blue;text-decoration:none">Twitter</a>`,
		},
		{
			name: "table_structure_normalization",
			input: `<table border="0">
  <tr>
    <td style="padding: 5px;  border:  1px solid #000 ">
      Cell content
    </td>
  </tr>
</table>`,
			expected: `<table border="0"><tr><td style="border:1px solid #000;padding:5px">Cell content</td></tr></table>`,
		},
		{
			name:     "mso_conditional_comments_preservation",
			input:    `<!--[if mso]><div>Outlook content</div><![endif]-->`,
			expected: `<!--[if mso]><div>Outlook content</div><![endif]-->`,
		},
		{
			name: "mixed_content_with_attributes",
			input: `<div class="container" style="font-family: Arial; font-size: 14px">
  <p style="margin: 0; text-align: center">
    Hello   World!
  </p>
</div>`,
			expected: `<div class="container" style="font-family:Arial;font-size:14px"><p style="margin:0;text-align:center">Hello World!</p></div>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizeForComparison(tt.input)
			if result != tt.expected {
				t.Errorf("NormalizeForComparison() failed\nInput:\n%s\nExpected:\n%s\nGot:\n%s", tt.input, tt.expected, result)
			}
		})
	}
}

func TestNormalizeForComparison_RealWorldScenarios(t *testing.T) {
	// Test scenarios based on actual MJML output differences
	tests := []struct {
		name        string
		html1       string
		html2       string
		shouldMatch bool
		description string
	}{
		{
			name: "whitespace_only_differences",
			html1: `<div style="color: red; padding: 10px">
  <span>Content</span>
</div>`,
			html2:       `<div style="padding: 10px; color: red"><span>Content</span></div>`,
			shouldMatch: true,
			description: "Should match when only whitespace and CSS order differ",
		},
		{
			name:        "real_content_differences",
			html1:       `<a href="https://twitter.com/handle">Twitter</a>`,
			html2:       `<a href="https://facebook.com/handle">Twitter</a>`,
			shouldMatch: false,
			description: "Should NOT match when href attributes differ",
		},
		{
			name:        "twitter_url_expansion_equivalence",
			html1:       `<a href="https://twitter.com/home?status=#">Twitter</a>`,
			html2:       `<a href="https://twitter.com/home?status=#">Twitter</a>`,
			shouldMatch: true,
			description: "Should match when Twitter URLs are identical after expansion",
		},
		{
			name:        "css_property_order_differences",
			html1:       `<div style="margin: 0; padding: 10px; color: blue; font-size: 14px">Content</div>`,
			html2:       `<div style="color: blue; font-size: 14px; margin: 0; padding: 10px">Content</div>`,
			shouldMatch: true,
			description: "Should match when CSS properties are in different order but same values",
		},
		{
			name:        "whitespace_in_css_values",
			html1:       `<div style="font-family: Arial, sans-serif; padding: 10px 5px">Content</div>`,
			html2:       `<div style="font-family:Arial, sans-serif;padding:10px 5px">Content</div>`,
			shouldMatch: true,
			description: "Should match when CSS has different whitespace around colons",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			norm1 := NormalizeForComparison(tt.html1)
			norm2 := NormalizeForComparison(tt.html2)

			matches := norm1 == norm2
			if matches != tt.shouldMatch {
				t.Errorf("%s\nHTML1 normalized: %s\nHTML2 normalized: %s\nExpected match: %v, got match: %v",
					tt.description, norm1, norm2, tt.shouldMatch, matches)
			}
		})
	}
}

func TestNormalizeForComparison_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty_string",
			input:    "",
			expected: "",
		},
		{
			name:     "whitespace_only",
			input:    "   \n\t  ",
			expected: "",
		},
		{
			name:     "single_tag",
			input:    "<div>content</div>",
			expected: "<div>content</div>",
		},
		{
			name:     "self_closing_tag",
			input:    "<br />",
			expected: "<br />",
		},
		{
			name:     "malformed_css_no_colon",
			input:    `<div style="color">content</div>`,
			expected: `<div style="">content</div>`,
		},
		{
			name:     "css_with_urls",
			input:    `<div style="background-image: url(https://example.com/image.jpg); color: red">content</div>`,
			expected: `<div style="background-image:url(https://example.com/image.jpg);color:red">content</div>`,
		},
		{
			name:     "nested_quotes_in_css",
			input:    `<div style='font-family: "Arial", sans-serif; color: red'>content</div>`,
			expected: `<div style='color:red;font-family:"Arial", sans-serif'>content</div>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizeForComparison(tt.input)
			if result != tt.expected {
				t.Errorf("NormalizeForComparison() edge case failed\nInput:\n%s\nExpected:\n%s\nGot:\n%s", tt.input, tt.expected, result)
			}
		})
	}
}

// Benchmark the normalization function to ensure it's performant
func BenchmarkNormalizeForComparison(b *testing.B) {
	complexHTML := `<!doctype html>
<html>
  <head>
    <title>Complex Test</title>
    <style>body { margin: 0; padding: 0; }</style>
  </head>
  <body style="font-family: Arial, sans-serif;   color: #333;   margin: 0;   padding: 20px">
    <div style="max-width:   600px;   margin:   0 auto;   background: white;   border: 1px solid #ddd">
      <table border="0" cellpadding="0" cellspacing="0" style="width: 100%;   border-collapse: collapse">
        <tr>
          <td style="padding:   20px;   text-align:   center;   background:   #f5f5f5">
            <h1 style="margin:   0;   color:   #333;   font-size:   24px">Welcome</h1>
          </td>
        </tr>
        <tr>
          <td style="padding:   20px">
            <p style="margin:   0 0 15px 0;   line-height:   1.6">
              This   is   a   complex   HTML   structure   with   multiple   styles   and   whitespace.
            </p>
          </td>
        </tr>
      </table>
    </div>
  </body>
</html>`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		NormalizeForComparison(complexHTML)
	}
}
