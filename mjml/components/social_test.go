package components

import (
	"strings"
	"testing"

	"github.com/preslavrachev/gomjml/mjml/options"
	"github.com/preslavrachev/gomjml/parser"
)

func TestSocialElementTwitterURL(t *testing.T) {
	tests := []struct {
		name        string
		mjml        string
		expectedURL string
		description string
	}{
		{
			name:        "explicit_twitter_url",
			mjml:        `<mj-social><mj-social-element name="twitter" href="https://twitter.com/myhandle" src="test.png">Follow us</mj-social-element></mj-social>`,
			expectedURL: `https://twitter.com/myhandle`,
			description: "Should preserve explicit Twitter URLs",
		},
		{
			name:        "hashtag_placeholder",
			mjml:        `<mj-social><mj-social-element name="twitter" href="#" src="test.png">Follow us</mj-social-element></mj-social>`,
			expectedURL: `https://twitter.com/home?status=#`,
			description: "Should expand # to default Twitter share URL",
		},
		{
			name:        "no_href_renders_without_link",
			mjml:        `<mj-social><mj-social-element name="twitter" src="test.png">Follow us</mj-social-element></mj-social>`,
			expectedURL: ``,
			description: "Should render as non-clickable when no href provided",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse MJML
			doc, err := parser.ParseMJML(tt.mjml)
			if err != nil {
				t.Fatalf("Failed to parse MJML: %v", err)
			}

			// Find social element
			socialElement := findMJMLElement(doc, "mj-social-element")
			if socialElement == nil {
				t.Fatal("mj-social-element not found")
			}

			// Create component
			socialComp := NewMJSocialElementComponent(socialElement, &options.RenderOpts{})

			// Render and check URL
			var buf strings.Builder
			err = socialComp.Render(&buf)
			if err != nil {
				t.Fatalf("Render failed: %v", err)
			}

			output := buf.String()

			if tt.expectedURL == "" {
				// Should not contain any href attribute
				if strings.Contains(output, `href="`) {
					t.Errorf("%s: Expected no href attribute in output, but found one in:\n%s", tt.description, output)
				}
			} else {
				// Should contain the expected href
				if !strings.Contains(output, `href="`+tt.expectedURL+`"`) {
					t.Errorf("%s: Expected href=\"%s\" in output, got:\n%s", tt.description, tt.expectedURL, output)
				}
			}
		})
	}
}

// Helper function to find MJML elements in the parsed document
func findMJMLElement(node *parser.MJMLNode, tagName string) *parser.MJMLNode {
	if node.XMLName.Local == tagName {
		return node
	}
	for _, child := range node.Children {
		if result := findMJMLElement(child, tagName); result != nil {
			return result
		}
	}
	return nil
}
