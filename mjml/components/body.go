package components

import (
	"io"
	"strings"

	"github.com/preslavrachev/gomjml/mjml/options"
	"github.com/preslavrachev/gomjml/parser"
)

// Email layout constants following MRML's architecture where mj-body defines the default width.
// In MRML, only mj_body/render.rs:74 defines the default "width" => Some("600px").
const (
	// DefaultBodyWidth is the default width of the email body in string format with units
	DefaultBodyWidth = "600px"

	// DefaultBodyWidthPixels is the default width of the email body as integer pixels
	DefaultBodyWidthPixels = 600
)

// MJBodyComponent represents mj-body
type MJBodyComponent struct {
	*BaseComponent
}

// NewMJBodyComponent creates a new mj-body component
func NewMJBodyComponent(node *parser.MJMLNode, opts *options.RenderOpts) *MJBodyComponent {
	return &MJBodyComponent{
		BaseComponent: NewBaseComponent(node, opts),
	}
}

func (c *MJBodyComponent) GetTagName() string {
	return "mj-body"
}

// Render implements optimized Writer-based rendering for MJBodyComponent
func (c *MJBodyComponent) Render(w io.StringWriter) error {
	backgroundColor := c.GetAttribute("background-color")
	langAttr := c.RenderOpts.Lang

	// Build class attribute: just use the user's css-class if present
	classAttr := c.BuildClassAttribute("")

	var b strings.Builder
	b.WriteString("<div")
	if langAttr != "" {
		b.WriteString(` lang="`)
		b.WriteString(langAttr)
		b.WriteString(`"`)
	}
	if classAttr != "" {
		b.WriteString(` class="`)
		b.WriteString(classAttr)
		b.WriteString(`"`)
	}
	if backgroundColor != nil && *backgroundColor != "" {
		b.WriteString(` style="background-color:`)
		b.WriteString(*backgroundColor)
		b.WriteString(`;"`)
	}
	b.WriteString(">")

	if _, err := w.WriteString(b.String()); err != nil {
		return err
	}

	for _, child := range c.Children {
		if err := child.Render(w); err != nil {
			return err
		}
	}

	_, err := w.WriteString("</div>")
	return err
}

func (c *MJBodyComponent) GetDefaultAttribute(name string) string {
	switch name {
	case "width":
		return DefaultBodyWidth
	default:
		return ""
	}
}

// GetDefaultBodyWidth returns the default body width as a string with units
func GetDefaultBodyWidth() string {
	return DefaultBodyWidth
}

// GetDefaultBodyWidthPixels returns the default body width as integer pixels
func GetDefaultBodyWidthPixels() int {
	return DefaultBodyWidthPixels
}
