package components

import (
	"io"

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
func (c *MJBodyComponent) Render(w io.Writer) error {
	// Apply background-color to div if specified (matching MRML's set_body_style)
	backgroundColor := c.GetAttribute("background-color")

	if backgroundColor != nil && *backgroundColor != "" {
		if _, err := w.Write([]byte(`<div style="background-color:` + *backgroundColor + `;">`)); err != nil {
			return err
		}
	} else {
		if _, err := w.Write([]byte(`<div>`)); err != nil {
			return err
		}
	}

	for _, child := range c.Children {
		if err := child.Render(w); err != nil {
			return err
		}
	}

	if _, err := w.Write([]byte(`</div>`)); err != nil {
		return err
	}

	return nil
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
