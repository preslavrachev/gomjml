package components

import (
	"io"

	"github.com/preslavrachev/gomjml/mjml/constants"
	"github.com/preslavrachev/gomjml/mjml/fonts"
	"github.com/preslavrachev/gomjml/mjml/html"
	"github.com/preslavrachev/gomjml/mjml/options"
	"github.com/preslavrachev/gomjml/parser"
)

// MJTableComponent represents the mj-table component
type MJTableComponent struct {
	*BaseComponent
}

func NewMJTableComponent(node *parser.MJMLNode, opts *options.RenderOpts) *MJTableComponent {
	return &MJTableComponent{
		BaseComponent: NewBaseComponent(node, opts),
	}
}

func (c *MJTableComponent) Render(w io.StringWriter) error {
	// Get attributes
	align := c.GetAttributeWithDefault(c, constants.MJMLAlign)
	padding := c.GetAttributeWithDefault(c, constants.MJMLPadding)
	containerBg := c.GetAttributeFast(c, constants.MJMLContainerBackgroundColor)

	// Create TR element
	if _, err := w.WriteString("<tr>"); err != nil {
		return err
	}

	// Create TD with alignment and base styles
	tdTag := html.NewHTMLTag("td").
		AddAttribute(constants.AttrAlign, align).
		AddStyle(constants.CSSFontSize, "0px").
		AddStyle(constants.CSSPadding, padding).
		AddStyle(constants.CSSWordBreak, "break-word")

	// Add container background color if specified
	if containerBg != "" {
		tdTag.AddStyle(constants.CSSBackground, containerBg)
	}

	// Add css-class if present
	c.SetClassAttribute(tdTag)

	// Add specific padding overrides if they exist
	if paddingTop := c.GetAttribute(constants.MJMLPaddingTop); paddingTop != nil {
		tdTag.AddStyle(constants.CSSPaddingTop, *paddingTop)
	}
	if paddingBottom := c.GetAttribute(constants.MJMLPaddingBottom); paddingBottom != nil {
		tdTag.AddStyle(constants.CSSPaddingBottom, *paddingBottom)
	}
	if paddingLeft := c.GetAttribute(constants.MJMLPaddingLeft); paddingLeft != nil {
		tdTag.AddStyle(constants.CSSPaddingLeft, *paddingLeft)
	}
	if paddingRight := c.GetAttribute(constants.MJMLPaddingRight); paddingRight != nil {
		tdTag.AddStyle(constants.CSSPaddingRight, *paddingRight)
	}

	if err := tdTag.RenderOpen(w); err != nil {
		return err
	}

	// Create table element with styles
	tableTag := html.NewHTMLTag("table").
		// HTML border attribute should always be "0", even with CSS border
		AddAttribute(constants.AttrBorder, "0").
		AddAttribute(constants.AttrCellPadding, c.GetAttributeWithDefault(c, "cellpadding")).
		AddAttribute(constants.AttrCellSpacing, c.GetAttributeWithDefault(c, "cellspacing")).
		AddAttribute(constants.AttrWidth, c.GetAttributeWithDefault(c, constants.MJMLWidth)).
		AddStyle(constants.CSSColor, c.GetAttributeWithDefault(c, constants.MJMLColor)).
		AddStyle(constants.CSSFontFamily, c.GetAttributeWithDefault(c, constants.MJMLFontFamily)).
		AddStyle(constants.CSSFontSize, c.GetAttributeWithDefault(c, constants.MJMLFontSize)).
		AddStyle(constants.CSSLineHeight, c.GetAttributeWithDefault(c, constants.MJMLLineHeight)).
		AddStyle(constants.CSSTableLayout, c.GetAttributeWithDefault(c, "table-layout")).
		AddStyle(constants.CSSWidth, c.GetAttributeWithDefault(c, constants.MJMLWidth)).
		AddStyle(constants.CSSBorder, c.GetAttributeWithDefault(c, constants.MJMLBorder)) // Use the actual border value for CSS

	if err := tableTag.RenderOpen(w); err != nil {
		return err
	}

	// Write the inner HTML content (TR, TH, TD elements)
	if err := c.writeInnerTableContent(w); err != nil {
		return err
	}

	if err := tableTag.RenderClose(w); err != nil {
		return err
	}

	if err := tdTag.RenderClose(w); err != nil {
		return err
	}

	if _, err := w.WriteString("</tr>"); err != nil {
		return err
	}

	return nil
}

func (c *MJTableComponent) GetTagName() string {
	return "mj-table"
}

func (c *MJTableComponent) GetDefaultAttribute(name string) string {
	switch name {
	case "align":
		return "left"
	case constants.MJMLBorder:
		return "none"
	case "cellpadding":
		return "0"
	case "cellspacing":
		return "0"
	case "color":
		return "#000000"
	case "font-family":
		return fonts.DefaultFontStack
	case "font-size":
		return "13px"
	case "line-height":
		return "22px"
	case "padding":
		return "10px 25px"
	case "table-layout":
		return "auto"
	case "width":
		return "100%"
	default:
		return ""
	}
}

// writeInnerTableContent writes the content (TR, TH, TD elements) inside the table
func (c *MJTableComponent) writeInnerTableContent(w io.StringWriter) error {
	_, err := w.WriteString(endingTagHTML(c.Node))
	return err
}
