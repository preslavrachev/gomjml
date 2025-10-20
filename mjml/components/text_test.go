package components

import (
	"strings"
	"testing"

	"github.com/preslavrachev/gomjml/mjml/options"
	"github.com/preslavrachev/gomjml/parser"
)

// TestMJTextRenderHTMLAsIs tests that mj-text renders HTML content as-is without parsing or correction.
// This test guards the spec requirement that mj-text is an "ending tag" component where all content
// is passed through directly to the output HTML without any additional MJML parsing, escaping, or correction.
func TestMJTextRenderHTMLAsIs(t *testing.T) {
	testCases := []struct {
		name     string
		mjml     string
		expected string
		desc     string
	}{
		{
			name:     "Standard unclosed br tag",
			mjml:     `<mj-text>Hello<br>World!</mj-text>`,
			expected: "Hello<br>World!",
			desc:     "Unclosed <br> tags should be preserved as-is",
		},
		{
			name:     "Self-closing br tag",
			mjml:     `<mj-text>Hello<br/>World!</mj-text>`,
			expected: "Hello<br />World!",
			desc:     "Self-closing <br/> tags are normalized to <br /> with space",
		},
		{
			name:     "Self-closing br with space",
			mjml:     `<mj-text>Hello<br />World!</mj-text>`,
			expected: "Hello<br />World!",
			desc:     "Self-closing <br /> tags with space should be preserved as-is",
		},
		{
			name:     "Malformed br tag",
			mjml:     `<mj-text>Hello<br bar>World!</mj-text>`,
			expected: "Hello<br bar>World!",
			desc:     "Malformed HTML tags should be passed through unchanged",
		},
		{
			name:     "Mixed br formats",
			mjml:     `<mj-text>Line1<br>Line2<br/>Line3<br />Line4</mj-text>`,
			expected: "Line1<br>Line2<br />Line3<br />Line4",
			desc:     "Mixed formats: <br> preserved, <br/> normalized to <br />",
		},
		{
			name:     "Complex HTML tags",
			mjml:     `<mj-text>Hello<span style="color: red">colored</span> text</mj-text>`,
			expected: "Hello<span style=\"color: red\">colored</span> text",
			desc:     "Complex HTML tags should be preserved with attributes",
		},
		{
			name:     "Unclosed HTML tag",
			mjml:     `<mj-text>Hello<strong>bold text</mj-text>`,
			expected: "Hello<strong>bold text",
			desc:     "Unclosed HTML tags should be passed through as-is",
		},
		{
			name:     "Unclosed img tag",
			mjml:     `<mj-text>Check this <img src="test.jpg"> image</mj-text>`,
			expected: "Check this <img src=\"test.jpg\"> image",
			desc:     "Unclosed <img> tags should be preserved as-is",
		},
		{
			name:     "Self-closing img tag",
			mjml:     `<mj-text>Check this <img src="test.jpg"/> image</mj-text>`,
			expected: "Check this <img src=\"test.jpg\" /> image",
			desc:     "Self-closing <img/> tags are normalized to <img /> with space",
		},
		{
			name:     "Self-closing img with space",
			mjml:     `<mj-text>Check this <img src="test.jpg" /> image</mj-text>`,
			expected: "Check this <img src=\"test.jpg\" /> image",
			desc:     "Self-closing <img /> tags with space should be preserved as-is",
		},
		{
			name:     "Mixed void elements",
			mjml:     `<mj-text>Line1<br>Image: <img src="test.jpg"><br/>End</mj-text>`,
			expected: "Line1<br>Image: <img src=\"test.jpg\"><br />End",
			desc:     "Mixed void elements: unclosed preserved, self-closing normalized",
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

			// Test the actual rendered inner HTML content
			var output strings.Builder
			err = component.writeRawInnerHTML(&output)
			if err != nil {
				t.Fatalf("Failed to write raw inner HTML: %v", err)
			}
			actual := strings.TrimSpace(output.String())

			if actual != tc.expected {
				t.Errorf("%s\nExpected: %q\nActual:   %q", tc.desc, tc.expected, actual)

				// Additional debugging information
				t.Logf("Raw node text: %q", textNode.Text)
				t.Logf("Mixed content parts: %d", len(textNode.MixedContent))
				for i, part := range textNode.MixedContent {
					if part.Node != nil {
						t.Logf("Part[%d]: Node <%s>", i, part.Node.XMLName.Local)
					} else {
						t.Logf("Part[%d]: Text %q", i, part.Text)
					}
				}
			}
		})
	}
}

// TestMJTextInvalidXMLShouldFail tests that malformed MJML XML should fail parsing
func TestMJTextInvalidXMLShouldFail(t *testing.T) {
	invalidMJML := `<mj-text>Hello<br>World!` // Missing closing tag

	_, err := parser.ParseMJML(invalidMJML)
	if err == nil {
		t.Error("Expected parsing to fail for invalid XML, but it succeeded")
	}
}

func TestNormalizeVoidHTMLTags(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Self-closing br drops slash",
			input:    "Hello<br />World",
			expected: "Hello<br>World",
		},
		{
			name:     "Uppercase BR drops slash",
			input:    "Hello<BR />World",
			expected: "Hello<BR>World",
		},
		{
			name:     "Other void tags remain self-closed",
			input:    "Check <img src=\"test.jpg\"/> image",
			expected: "Check <img src=\"test.jpg\" /> image",
		},
		{
			name:     "Existing spacing preserved",
			input:    "<meta name=\"viewport\" content=\"width=device-width\" />",
			expected: "<meta name=\"viewport\" content=\"width=device-width\" />",
		},
		{
			name:     "Spaces around br removed",
			input:    "Austin, TX <br /> <span>-</span>",
			expected: "Austin, TX<br><span>-</span>",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := normalizeVoidHTMLTags(tc.input)
			if actual != tc.expected {
				t.Errorf("normalizeVoidHTMLTags mismatch\nexpected: %q\nactual:   %q", tc.expected, actual)
			}
		})
	}
}
