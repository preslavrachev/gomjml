package components

import (
	"io"

	"github.com/preslavrachev/gomjml/mjml/constants"
	"github.com/preslavrachev/gomjml/mjml/html"
	"github.com/preslavrachev/gomjml/mjml/options"
	"github.com/preslavrachev/gomjml/parser"
)

// MJSpacerComponent represents the mj-spacer component
type MJSpacerComponent struct {
	*BaseComponent
}

func NewMJSpacerComponent(node *parser.MJMLNode, opts *options.RenderOpts) *MJSpacerComponent {
	return &MJSpacerComponent{
		BaseComponent: NewBaseComponent(node, opts),
	}
}

func (c *MJSpacerComponent) Render(w io.StringWriter) error {
	height := c.GetAttributeWithDefault(c, constants.MJMLHeight)
	containerBackgroundColor := c.GetAttributeWithDefault(c, constants.MJMLContainerBackgroundColor)
	padding := c.GetAttributeWithDefault(c, constants.MJMLPadding)
	paddingTop := c.GetAttributeWithDefault(c, constants.MJMLPaddingTop)
	paddingRight := c.GetAttributeWithDefault(c, constants.MJMLPaddingRight)
	paddingBottom := c.GetAttributeWithDefault(c, constants.MJMLPaddingBottom)
	paddingLeft := c.GetAttributeWithDefault(c, constants.MJMLPaddingLeft)
	verticalAlign := c.GetAttributeWithDefault(c, constants.MJMLVerticalAlign)

	// Create table row
	if _, err := w.WriteString("<tr>"); err != nil {
		return err
	}

	// Create table cell with base styles
	td := html.NewHTMLTag("td").
		AddStyle(constants.CSSFontSize, "0px").
		AddStyle(constants.CSSWordBreak, "break-word")

	// Add optional styles
	if containerBackgroundColor != "" {
		td.AddStyle(constants.CSSBackground, containerBackgroundColor)
	}

	if classAttr := c.BuildClassAttribute(); classAttr != "" {
		td.AddAttribute(constants.AttrClass, classAttr)
	}

	// Handle padding attributes
	if padding != "" {
		td.AddStyle(constants.CSSPadding, padding)
	}
	if paddingTop != "" {
		td.AddStyle(constants.CSSPaddingTop, paddingTop)
	}
	if paddingRight != "" {
		td.AddStyle(constants.CSSPaddingRight, paddingRight)
	}
	if paddingBottom != "" {
		td.AddStyle(constants.CSSPaddingBottom, paddingBottom)
	}
	if paddingLeft != "" {
		td.AddStyle(constants.CSSPaddingLeft, paddingLeft)
	}

	// Add vertical-align as attribute (not style) if specified
	if verticalAlign != "" {
		td.AddAttribute(constants.AttrVerticalAlign, verticalAlign)
	}

	// Render table cell opening tag
	if err := td.RenderOpen(w); err != nil {
		return err
	}

	// Create div with height and line-height, containing thin space character
	div := html.NewHTMLTag("div").
		AddStyle(constants.CSSHeight, height).
		AddStyle(constants.CSSLineHeight, height)

	if err := div.RenderOpen(w); err != nil {
		return err
	}

	// Add thin space character (&#8202;)
	if _, err := w.WriteString("&#8202;"); err != nil {
		return err
	}

	if err := div.RenderClose(w); err != nil {
		return err
	}

	// Close table cell and row
	if err := td.RenderClose(w); err != nil {
		return err
	}

	if _, err := w.WriteString("</tr>"); err != nil {
		return err
	}

	return nil
}

func (c *MJSpacerComponent) GetTagName() string {
	return "mj-spacer"
}

func (c *MJSpacerComponent) GetDefaultAttribute(name string) string {
	switch name {
	case constants.MJMLHeight:
		return "20px"
	default:
		return ""
	}
}
