package html

import "testing"

func TestEscapeXMLAttr(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		// Basic escaping
		{"hello", "hello"},
		{"", ""},

		// Double quotes
		{"value=\"test\"", "value=&quot;test&quot;"},

		// Single quotes
		{"value='test'", "value=&#39;test&#39;"},

		// Ampersand
		{"Tom & Jerry", "Tom &amp; Jerry"},

		// Less than and greater than
		{"<script>alert('xss')</script>", "&lt;script&gt;alert(&#39;xss&#39;)&lt;/script&gt;"},

		// Complex injection attempt
		{"onclick=\"alert('XSS')\"", "onclick=&quot;alert(&#39;XSS&#39;)&quot;"},

		// Multiple entities
		{"A & B \"quoted\" <tag> 'value'", "A &amp; B &quot;quoted&quot; &lt;tag&gt; &#39;value&#39;"},

		// Edge cases
		{"&amp;", "&amp;amp;"},                   // Already escaped ampersand
		{"&quot;", "&amp;quot;"},                 // Already escaped quote
		{"&lt;test&gt;", "&amp;lt;test&amp;gt;"}, // Already escaped brackets
	}

	for _, test := range tests {
		result := EscapeXMLAttr(test.input)
		if result != test.expected {
			t.Errorf("EscapeXMLAttr(%q) = %q, expected %q", test.input, result, test.expected)
		}
	}
}
