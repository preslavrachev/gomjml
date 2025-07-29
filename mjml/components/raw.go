package components

import (
	"strings"

	"github.com/preslavrachev/gomjml/mjml/options"
	"github.com/preslavrachev/gomjml/parser"
)

// MJRawComponent represents an mj-raw component
type MJRawComponent struct {
	*BaseComponent
}

// NewMJRawComponent creates a new mj-raw component
func NewMJRawComponent(node *parser.MJMLNode, opts *options.RenderOpts) *MJRawComponent {
	return &MJRawComponent{
		BaseComponent: NewBaseComponent(node, opts),
	}
}

// GetTagName returns the component's tag name
func (c *MJRawComponent) GetTagName() string {
	return "mj-raw"
}

// GetDefaultAttribute returns default values for the component's attributes
func (c *MJRawComponent) GetDefaultAttribute(name string) string {
	// mj-raw has no default attributes
	return ""
}

// Render renders the mj-raw component to HTML
func (c *MJRawComponent) Render() (string, error) {
	// mj-raw simply outputs its content as-is, without any wrapping HTML
	// The content includes both text and child nodes as raw HTML
	return c.getRawContent(), nil
}

// getRawContent returns the raw content of the mj-raw component
// This includes both text content and serialized child elements
func (c *MJRawComponent) getRawContent() string {
	// For mj-raw, we need to reconstruct the original HTML content
	// since the XML parser has already parsed it into nodes
	return c.reconstructRawHTML()
}

// reconstructRawHTML reconstructs the original HTML from the parsed XML nodes
func (c *MJRawComponent) reconstructRawHTML() string {
	if len(c.Node.Children) == 0 {
		return c.Node.Text
	}

	var builder strings.Builder
	builder.Grow(len(c.Node.Text) + len(c.Node.Children)*32) // Estimate content size

	builder.WriteString(c.Node.Text)

	// Add any child nodes as raw HTML
	for _, child := range c.Node.Children {
		builder.WriteString(c.nodeToHTML(child))
	}

	return builder.String()
}

// nodeToHTML converts an XML node back to HTML string
func (c *MJRawComponent) nodeToHTML(node *parser.MJMLNode) string {
	if node == nil {
		return ""
	}

	tagName := node.XMLName.Local
	var builder strings.Builder

	// Pre-allocate reasonable capacity to reduce reallocations
	builder.Grow(64 + len(tagName)*2 + len(node.Text))

	// Handle self-closing tags
	if len(node.Children) == 0 && node.Text == "" {
		builder.WriteByte('<')
		builder.WriteString(tagName)

		for _, attr := range node.Attrs {
			builder.WriteByte(' ')
			builder.WriteString(attr.Name.Local)
			builder.WriteString(`="`)
			builder.WriteString(attr.Value)
			builder.WriteByte('"')
		}

		// For HTML5 void elements, use self-closing syntax only for XML compatibility
		// But for HTML output, canvas should have closing tags
		if tagName == "img" || tagName == "br" || tagName == "hr" || tagName == "meta" || tagName == "input" {
			builder.WriteString(" />")
		} else {
			// Canvas and other elements should have closing tags in HTML
			builder.WriteString("></")
			builder.WriteString(tagName)
			builder.WriteByte('>')
		}
		return builder.String()
	}

	// Handle tags with content
	builder.WriteByte('<')
	builder.WriteString(tagName)

	for _, attr := range node.Attrs {
		builder.WriteByte(' ')
		builder.WriteString(attr.Name.Local)
		builder.WriteString(`="`)
		builder.WriteString(attr.Value)
		builder.WriteByte('"')
	}
	builder.WriteByte('>')

	// Add text content
	builder.WriteString(node.Text)

	// Add child nodes
	for _, child := range node.Children {
		builder.WriteString(c.nodeToHTML(child))
	}

	builder.WriteString("</")
	builder.WriteString(tagName)
	builder.WriteByte('>')
	return builder.String()
}
