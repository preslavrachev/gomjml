package mjml

import (
	"strings"
	"testing"
)

// helper
func containsUbuntuBlock(html string) bool {
	return strings.Contains(html, "https://fonts.googleapis.com/css?family=Ubuntu:300,400,500,700") &&
		strings.Contains(html, "@import url(https://fonts.googleapis.com/css?family=Ubuntu:300,400,500,700);") &&
		strings.Contains(html, "<!--[if !mso]><!-->") &&
		strings.Contains(html, "<!--<![endif]-->")
}

// (If you have a RenderOptions/RenderOpts, swap in accordingly.)
func renderMJML(t *testing.T, mjml string) string {
	t.Helper()
	out, err := Render(mjml)
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}
	return out
}

// TestFontImportTriggerMatrix verifies that the MJML renderer correctly includes or omits
// the Ubuntu font import block in the generated HTML <head> based on the presence and type
// of MJML components in the input markup. It tests various scenarios, including default
// text components, buttons, social elements, navbar links, accordion elements, hero headings,
// tables, images, and font-family overrides, ensuring that the font import logic matches
// MJML's upstream behavior and respects explicit font-family settings.
func TestFontImportTriggerMatrix(t *testing.T) {
	tests := []struct {
		name         string
		mjml         string
		expectUbuntu bool
		reason       string
	}{
		{
			name:         "Text component triggers",
			mjml:         `<mjml><mj-body><mj-section><mj-column><mj-text>Hello</mj-text></mj-column></mj-section></mj-body></mjml>`,
			expectUbuntu: true,
			reason:       "mj-text default font stack includes Ubuntu",
		},
		{
			name:         "Button only triggers",
			mjml:         `<mjml><mj-body><mj-section><mj-column><mj-button>Click</mj-button></mj-column></mj-section></mj-body></mjml>`,
			expectUbuntu: true,
			reason:       "mj-button default stack includes Ubuntu",
		},
		{
			name:         "Social triggers",
			mjml:         `<mjml><mj-body><mj-section><mj-column><mj-social><mj-social-element name="twitter">Tw</mj-social-element></mj-social></mj-column></mj-section></mj-body></mjml>`,
			expectUbuntu: true,
			reason:       "mj-social-element text uses default stack",
		},
		{
			name:         "Navbar triggers",
			mjml:         `<mjml><mj-body><mj-section><mj-column><mj-navbar><mj-navbar-link href="https://x">X</mj-navbar-link></mj-navbar></mj-column></mj-section></mj-body></mjml>`,
			expectUbuntu: true,
			reason:       "mj-navbar-link default stack includes Ubuntu",
		},
		{
			name:         "Accordion triggers",
			mjml:         `<mjml><mj-body><mj-section><mj-column><mj-accordion><mj-accordion-element><mj-accordion-title>Q</mj-accordion-title><mj-accordion-text>A</mj-accordion-text></mj-accordion-element></mj-accordion></mj-column></mj-section></mj-body></mjml>`,
			expectUbuntu: true,
			reason:       "accordion title/text default stack includes Ubuntu",
		},
		{
			name:         "Hero heading triggers (text inside hero)",
			mjml:         `<mjml><mj-body><mj-hero mode="fixed-height" height="200px"><mj-text>Hero</mj-text></mj-hero></mj-body></mjml>`,
			expectUbuntu: true,
			reason:       "mj-text nested in hero still a text component",
		},
		{
			name:         "Table triggers",
			mjml:         `<mjml><mj-body><mj-section><mj-column><mj-table><tr><td>Cell</td></tr></mj-table></mj-column></mj-section></mj-body></mjml>`,
			expectUbuntu: true,
			reason:       "mj-table default font-family includes Ubuntu",
		},
		{
			name:         "Image only does NOT trigger",
			mjml:         `<mjml><mj-body><mj-section><mj-column><mj-image src="x.png" /></mj-column></mj-section></mj-body></mjml>`,
			expectUbuntu: false,
			reason:       "No text-like component; upstream MJML omits font block",
		},
		{
			name:         "System font override removes Ubuntu",
			mjml:         `<mjml><mj-head><mj-attributes><mj-all font-family="Arial, Helvetica, sans-serif" /></mj-attributes></mj-head><mj-body><mj-section><mj-column><mj-text>Override</mj-text></mj-column></mj-section></mj-body></mjml>`,
			expectUbuntu: false,
			reason:       "Override removes Google font token from first position",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			html := renderMJML(t, tc.mjml)
			got := containsUbuntuBlock(html)
			if got != tc.expectUbuntu {
				t.Errorf("Ubuntu block presence=%v want %v (%s)\nHTML(head snippet)=%s",
					got, tc.expectUbuntu, tc.reason, headSnippet(html, 500))
			}
		})
	}
}

// headSnippet extracts first n chars from <head> for debugging
func headSnippet(html string, n int) string {
	if len(html) < n {
		return html
	}
	return html[:n]
}

// TestFontsBlockOrderingRelativeToMediaQueries verifies that the conditional fonts block
// (specifically for Ubuntu fonts) is present in the rendered HTML and that it appears
// before the first responsive @media rule. This ensures correct ordering of font and
// media query blocks for proper email client rendering.
func TestFontsBlockOrderingRelativeToMediaQueries(t *testing.T) {
	html := renderMJML(t,
		`<mjml><mj-body><mj-section><mj-column><mj-text>Hello</mj-text></mj-column></mj-section></mj-body></mjml>`)

	// Basic presence
	if !containsUbuntuBlock(html) {
		t.Fatalf("expected Ubuntu fonts block")
	}

	fontBlockIdx := strings.Index(html, "<!--[if !mso]><!-->")
	mediaIdx := strings.Index(html, "@media only screen")
	if mediaIdx < 0 {
		t.Fatalf("expected a responsive @media style")
	}
	if !(fontBlockIdx >= 0 && fontBlockIdx < mediaIdx) {
		t.Errorf("expected fonts conditional block to precede first @media rule: fontBlockIdx=%d mediaIdx=%d", fontBlockIdx, mediaIdx)
	}
}
