package mjml

import (
	"strings"
	"testing"
)

// TestRenderPreservesEscapedXMLEntities verifies that escaped markup in
// mj-text and mj-title is rendered as text, never as real HTML tags.
//
// Bug: Previously, "&lt;1&gt;" in mj-text was rendered as a raw "<1>" tag,
// which produced invalid HTML and allowed escaped user content such as
// "&lt;script&gt;" to be injected as markup.
func TestRenderPreservesEscapedXMLEntities(t *testing.T) {
	tests := []struct {
		name          string
		mjml          string
		expected      string
		shouldNotHave string
	}{
		{
			name:          "mj-text escaped pseudo tag",
			mjml:          `<mjml><mj-body><mj-section><mj-column><mj-text>hello Q&amp;A &lt;1&gt;</mj-text></mj-column></mj-section></mj-body></mjml>`,
			expected:      `hello Q&amp;A &lt;1&gt;`,
			shouldNotHave: `<1>`,
		},
		{
			name:          "mj-text escaped script tag",
			mjml:          `<mjml><mj-body><mj-section><mj-column><mj-text>&lt;script&gt;alert(1)&lt;/script&gt;</mj-text></mj-column></mj-section></mj-body></mjml>`,
			expected:      `&lt;script&gt;alert(1)&lt;/script&gt;`,
			shouldNotHave: `<script>`,
		},
		{
			name:          "mj-title double escaped",
			mjml:          `<mjml><mj-head><mj-title>Q&amp;amp;A &amp;lt;1&amp;gt; Sujet</mj-title></mj-head><mj-body></mj-body></mjml>`,
			expected:      `<title>Q&amp;A &lt;1&gt; Sujet</title>`,
			shouldNotHave: `<1>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			html, err := Render(tt.mjml)
			if err != nil {
				t.Fatalf("Render() error = %v", err)
			}

			if !strings.Contains(html, tt.expected) {
				t.Errorf("expected output to contain %q", tt.expected)
			}
			if strings.Contains(html, tt.shouldNotHave) {
				t.Errorf("expected output not to contain %q", tt.shouldNotHave)
			}
		})
	}
}
