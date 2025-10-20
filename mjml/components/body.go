package components

import (
	"io"
	"strings"

	"github.com/preslavrachev/gomjml/mjml/constants"
	"github.com/preslavrachev/gomjml/mjml/html"
	"github.com/preslavrachev/gomjml/mjml/options"
	"github.com/preslavrachev/gomjml/mjml/styles"
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

// GetEffectiveWidth returns the body width in pixels, allowing the width
// attribute to override the default 600px container width. This ensures
// section children inherit the correct width.
func (c *MJBodyComponent) GetEffectiveWidth() int {
	if c.ContainerWidth > 0 {
		return c.ContainerWidth
	}
	if widthAttr := c.GetAttributeWithDefault(c, "width"); widthAttr != "" {
		if size, err := styles.ParseSize(widthAttr); err == nil && size.IsPixel() {
			return int(size.Value())
		}
	}
	return GetDefaultBodyWidthPixels()
}

// GetEffectiveWidthString returns the body width as a pixel string, honoring
// the width attribute when provided.
func (c *MJBodyComponent) GetEffectiveWidthString() string {
	if c.ContainerWidth > 0 {
		return getPixelWidthString(c.ContainerWidth)
	}
	if widthAttr := c.GetAttributeWithDefault(c, "width"); widthAttr != "" {
		if size, err := styles.ParseSize(widthAttr); err == nil && size.IsPixel() {
			return getPixelWidthString(int(size.Value()))
		}
	}
	return GetDefaultBodyWidth()
}

// Render implements optimized Writer-based rendering for MJBodyComponent
func (c *MJBodyComponent) Render(w io.StringWriter) error {
	backgroundColor := c.GetAttribute("background-color")
	langAttr := c.RenderOpts.Lang

	// Build class attribute: just use the user's css-class if present
	classAttr := c.BuildClassAttribute("")
	bodyDiv := html.NewHTMLTag("div")
	bodyDiv.AddAttribute("aria-roledescription", "email").
		AddAttribute("role", "article")

	if langAttr != "" {
		bodyDiv.AddAttribute("lang", langAttr).
			AddAttribute("dir", constants.DirAuto)
	}

	if title := strings.TrimSpace(c.RenderOpts.Title); title != "" {
		bodyDiv.AddAttribute(constants.AttrAriaLabel, title)
	}

	if classAttr != "" {
		bodyDiv.AddAttribute("class", classAttr)
		c.ApplyInlineStyles(bodyDiv, classAttr)
	}

	if backgroundColor != nil && *backgroundColor != "" {
		bodyDiv.AddStyle("background-color", *backgroundColor)
	}

	if err := bodyDiv.RenderOpen(w); err != nil {
		return err
	}

	if c.RenderOpts != nil {
		c.RenderOpts.PendingMSOSectionClose = false
	}

	// Track how many Outlook-sensitive blocks remain (mj-section and mj-wrapper)
	// so conditional comments can be chained correctly across mixed content.
	remainingBlocks := 0
	for _, child := range c.Children {
		switch child.(type) {
		case *MJSectionComponent, *MJWrapperComponent:
			remainingBlocks++
		}
	}

	for _, child := range c.Children {
		switch comp := child.(type) {
		case *MJSectionComponent:
			remainingBlocks--
			if c.RenderOpts != nil {
				c.RenderOpts.RemainingBodySections = remainingBlocks
			}
			if err := comp.Render(w); err != nil {
				return err
			}
			continue
		case *MJWrapperComponent:
			remainingBlocks--
			if c.RenderOpts != nil {
				c.RenderOpts.RemainingBodySections = remainingBlocks
			}
			if err := comp.Render(w); err != nil {
				return err
			}
			continue
		}

		if err := child.Render(w); err != nil {
			return err
		}
	}

	if c.RenderOpts != nil {
		c.RenderOpts.PendingMSOSectionClose = false
		c.RenderOpts.RemainingBodySections = 0
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
