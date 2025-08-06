package components

import (
	"io"

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

// Render implements optimized Writer-based rendering for MJRawComponent
func (c *MJRawComponent) RenderHTML(w io.StringWriter) error {
	// mj-raw simply outputs its content as-is, without any wrapping HTML
	// The content includes both text and child nodes as raw HTML
	content := c.getRawContent()
	if _, err := w.WriteString(content); err != nil {
		return err
	}
	return nil
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
	content := c.Node.Text

	// Add any child nodes as raw HTML
	for _, child := range c.Node.Children {
		content += c.nodeToHTML(child)
	}

	return content
}

// nodeToHTML converts an XML node back to HTML string
func (c *MJRawComponent) nodeToHTML(node *parser.MJMLNode) string {
	if node == nil {
		return ""
	}

	tagName := node.XMLName.Local

	// Handle self-closing tags
	if len(node.Children) == 0 && node.Text == "" {
		html := "<" + tagName
		for _, attr := range node.Attrs {
			html += " " + attr.Name.Local + `="` + attr.Value + `"`
		}

		// For HTML5 void elements, use self-closing syntax only for XML compatibility
		// But for HTML output, canvas should have closing tags
		if tagName == "img" || tagName == "br" || tagName == "hr" || tagName == "meta" || tagName == "input" {
			html += " />"
		} else {
			// Canvas and other elements should have closing tags in HTML
			html += "></" + tagName + ">"
		}
		return html
	}

	// Handle tags with content
	html := "<" + tagName
	for _, attr := range node.Attrs {
		html += " " + attr.Name.Local + `="` + attr.Value + `"`
	}
	html += ">"

	// Add text content
	html += node.Text

	// Add child nodes
	for _, child := range node.Children {
		html += c.nodeToHTML(child)
	}

	html += "</" + tagName + ">"
	return html
}

func (c *MJRawComponent) RenderMJML(w io.StringWriter) error {
	return &NotImplementedError{ComponentName: "mj-raw"}
}
