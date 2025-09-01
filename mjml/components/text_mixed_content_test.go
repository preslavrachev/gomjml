package components

import (
	"strings"
	"testing"

	"github.com/preslavrachev/gomjml/mjml/options"
	"github.com/preslavrachev/gomjml/parser"
)

// TestMJTextMixedContentWithBR verifies that the MJText component correctly reconstructs mixed content
// containing <br /> tags within <mj-text> elements. It checks various scenarios including multiple <br /> tags
// in the middle, at the end, and in different patterns, ensuring that the output matches the expected
// HTML-like string representation. The test parses the MJML input, locates the <mj-text> node, constructs
// the component, and asserts that the reconstructed mixed content matches the expected result.
func TestMJTextMixedContentWithBR(t *testing.T) {
	testCases := []struct {
		name     string
		mjml     string
		expected string
	}{
		{
			name:     "BR tags in middle of text",
			mjml:     `<mj-text>Hi John,<br /><br />Welcome to this week's newsletter!</mj-text>`,
			expected: "Hi John,<br /><br />Welcome to this week's newsletter!",
		},
		{
			name:     "Single BR tag in text",
			mjml:     `<mj-text>First line<br />Second line</mj-text>`,
			expected: "First line<br />Second line",
		},
		{
			name:     "BR tags at end",
			mjml:     `<mj-text>Some text<br /><br /></mj-text>`,
			expected: "Some text<br /><br />",
		},
		{
			name:     "Multiple BR patterns",
			mjml:     `<mj-text>• Feature releases<br />• Industry insights<br />• Success stories</mj-text>`,
			expected: "• Feature releases<br />• Industry insights<br />• Success stories",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Parse the MJML
			ast, err := parser.ParseMJML(tc.mjml)
			if err != nil {
				t.Fatalf("Failed to parse MJML: %v", err)
			}

			// Find the mj-text node
			var textNode *parser.MJMLNode
			findTextNode(ast, &textNode)
			if textNode == nil {
				t.Fatal("Could not find mj-text node in parsed AST")
			}

			// Create the component
			component := NewMJTextComponent(textNode, &options.RenderOpts{})

			// Test the actual rendered output by capturing what gets written
			var output strings.Builder
			err = component.writeRawInnerHTML(&output)
			if err != nil {
				t.Fatalf("Failed to write raw inner HTML: %v", err)
			}
			actual := output.String()

			if actual != tc.expected {
				t.Errorf("Mixed content mismatch:\nExpected: %s\nActual:   %s", tc.expected, actual)

				// Additional debugging
				t.Logf("Node.Text: %q", textNode.Text)
				t.Logf("Children count: %d", len(textNode.Children))
				for i, child := range textNode.Children {
					t.Logf("Child[%d]: tag=%s, text=%q", i, child.XMLName.Local, child.Text)
				}
			}
		})
	}
}

// findTextNode recursively finds the first mj-text node in the AST
func findTextNode(node *parser.MJMLNode, result **parser.MJMLNode) {
	if node.XMLName.Local == "mj-text" {
		*result = node
		return
	}
	for _, child := range node.Children {
		findTextNode(child, result)
		if *result != nil {
			return
		}
	}
}
