package mjml

import (
	"strings"
	"testing"
)

// TestButtonWithHTMLContent verifies that mj-button correctly renders content
// containing HTML tags like <strong>, <em>, etc. Per the MJML specification,
// mj-button is an "ending tag" that can contain HTML code.
//
// Bug: Previously, buttons with HTML content would fall back to displaying
// "Button" (the default text) because c.Node.Text was empty when content
// was wrapped in HTML tags.
func TestButtonWithHTMLContent(t *testing.T) {
	tests := []struct {
		name          string
		mjml          string
		expectedText  string
		shouldNotHave string
	}{
		{
			name:          "plain text button",
			mjml:          `<mjml><mj-body><mj-section><mj-column><mj-button href="#">Click</mj-button></mj-column></mj-section></mj-body></mjml>`,
			expectedText:  "Click",
			shouldNotHave: ">Button<",
		},
		{
			name:          "button with strong tag",
			mjml:          `<mjml><mj-body><mj-section><mj-column><mj-button href="#"><strong>Click Here</strong></mj-button></mj-column></mj-section></mj-body></mjml>`,
			expectedText:  "Click Here",
			shouldNotHave: ">Button<",
		},
		{
			name:          "button with em tag",
			mjml:          `<mjml><mj-body><mj-section><mj-column><mj-button href="#"><em>Important</em></mj-button></mj-column></mj-section></mj-body></mjml>`,
			expectedText:  "Important",
			shouldNotHave: ">Button<",
		},
		{
			name:          "button with nested formatting",
			mjml:          `<mjml><mj-body><mj-section><mj-column><mj-button href="#"><strong><em>Bold Italic</em></strong></mj-button></mj-column></mj-section></mj-body></mjml>`,
			expectedText:  "Bold Italic",
			shouldNotHave: ">Button<",
		},
		{
			name:          "button with mixed text and HTML",
			mjml:          `<mjml><mj-body><mj-section><mj-column><mj-button href="#">Start <strong>Now</strong>!</mj-button></mj-column></mj-section></mj-body></mjml>`,
			expectedText:  "Now",
			shouldNotHave: ">Button<",
		},
		{
			name:          "button with span styling",
			mjml:          `<mjml><mj-body><mj-section><mj-column><mj-button href="#"><span style="color: red;">Red Text</span></mj-button></mj-column></mj-section></mj-body></mjml>`,
			expectedText:  "Red Text",
			shouldNotHave: ">Button<",
		},
		{
			name:          "button with br tag only",
			mjml:          `<mjml><mj-body><mj-section><mj-column><mj-button href="#">Line1<br/>Line2</mj-button></mj-column></mj-section></mj-body></mjml>`,
			expectedText:  "Line1",
			shouldNotHave: ">Button<",
		},
		{
			name:          "empty button falls back to default",
			mjml:          `<mjml><mj-body><mj-section><mj-column><mj-button href="#"></mj-button></mj-column></mj-section></mj-body></mjml>`,
			expectedText:  "Button",
			shouldNotHave: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			html, err := Render(tt.mjml)
			if err != nil {
				t.Fatalf("Render failed: %v", err)
			}

			if !strings.Contains(html, tt.expectedText) {
				t.Errorf("Expected button to contain %q, but it didn't.\nHTML output:\n%s", tt.expectedText, html)
			}

			if tt.shouldNotHave != "" && strings.Contains(html, tt.shouldNotHave) {
				t.Errorf("Button should NOT contain %q (default fallback), but it did.\nThis indicates the bug where HTML content causes fallback to 'Button'.\nHTML output:\n%s", tt.shouldNotHave, html)
			}
		})
	}
}

// TestButtonHTMLContentPreserved verifies that HTML formatting inside buttons
// is actually preserved in the output (not just that the text is extracted).
func TestButtonHTMLContentPreserved(t *testing.T) {
	tests := []struct {
		name         string
		mjml         string
		expectedHTML string
	}{
		{
			name:         "strong tag preserved",
			mjml:         `<mjml><mj-body><mj-section><mj-column><mj-button href="#"><strong>Bold</strong></mj-button></mj-column></mj-section></mj-body></mjml>`,
			expectedHTML: "<strong>Bold</strong>",
		},
		{
			name:         "em tag preserved",
			mjml:         `<mjml><mj-body><mj-section><mj-column><mj-button href="#"><em>Italic</em></mj-button></mj-column></mj-section></mj-body></mjml>`,
			expectedHTML: "<em>Italic</em>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			html, err := Render(tt.mjml)
			if err != nil {
				t.Fatalf("Render failed: %v", err)
			}

			if !strings.Contains(html, tt.expectedHTML) {
				t.Errorf("Expected HTML formatting %q to be preserved in button, but it wasn't.\nHTML output:\n%s", tt.expectedHTML, html)
			}
		})
	}
}
