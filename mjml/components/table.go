package components

import (
	"io"
	"strings"

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
	if bgColor := c.GetAttribute(constants.MJMLContainerBackgroundColor); bgColor != nil {
		tdTag.AddStyle(constants.CSSBackground, *bgColor)
	}

	// Add css-class if present
	if cssClass := c.BuildClassAttribute(); cssClass != "" {
		tdTag.AddAttribute(constants.AttrClass, cssClass)
	}

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
	borderValue := c.GetAttributeWithDefault(c, constants.MJMLBorder)
	htmlBorderValue := "0" // HTML border attribute should always be "0"
	if borderValue != "none" {
		htmlBorderValue = "0" // Even with CSS border, HTML border should be "0"
	}

	tableTag := html.NewHTMLTag("table").
		AddAttribute(constants.AttrBorder, htmlBorderValue).
		AddAttribute(constants.AttrCellPadding, c.GetAttributeWithDefault(c, "cellpadding")).
		AddAttribute(constants.AttrCellSpacing, c.GetAttributeWithDefault(c, "cellspacing")).
		AddAttribute(constants.AttrWidth, c.GetAttributeWithDefault(c, constants.MJMLWidth)).
		AddStyle(constants.CSSColor, c.GetAttributeWithDefault(c, constants.MJMLColor)).
		AddStyle(constants.CSSFontFamily, c.GetAttributeWithDefault(c, constants.MJMLFontFamily)).
		AddStyle(constants.CSSFontSize, c.GetAttributeWithDefault(c, constants.MJMLFontSize)).
		AddStyle(constants.CSSLineHeight, c.GetAttributeWithDefault(c, constants.MJMLLineHeight)).
		AddStyle(constants.CSSTableLayout, c.GetAttributeWithDefault(c, "table-layout")).
		AddStyle(constants.CSSWidth, c.GetAttributeWithDefault(c, constants.MJMLWidth)).
		AddStyle(constants.CSSBorder, borderValue) // Use the actual border value for CSS

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

// writeInnerTableContent writes the inner HTML content (TR, TH, TD elements) to the writer
func (c *MJTableComponent) writeInnerTableContent(w io.StringWriter) error {
	// If we have children (HTML elements), we need to reconstruct the original HTML
	if len(c.Node.Children) > 0 {
		// Add children as HTML elements (skip text content to avoid extra whitespace)
		for _, child := range c.Node.Children {
			if err := c.reconstructHTMLElement(child, w); err != nil {
				return err
			}
		}
		return nil
	}

	// If no children, write the text content with whitespace trimmed
	if c.Node.Text != "" {
		_, err := w.WriteString(strings.TrimSpace(c.Node.Text))
		return err
	}

	return nil
}

// reconstructHTMLElement reconstructs an HTML element from a parsed node
func (c *MJTableComponent) reconstructHTMLElement(node *parser.MJMLNode, w io.StringWriter) error {
	tagName := node.XMLName.Local

	// Check if this is a void element (self-closing)
	isVoidElement := isVoidHTMLElement(tagName)

	// Opening tag
	if _, err := w.WriteString("<"); err != nil {
		return err
	}
	if _, err := w.WriteString(tagName); err != nil {
		return err
	}

	// Attributes
	for _, attr := range node.Attrs {
		if _, err := w.WriteString(" "); err != nil {
			return err
		}
		if _, err := w.WriteString(attr.Name.Local); err != nil {
			return err
		}
		if _, err := w.WriteString(`="`); err != nil {
			return err
		}
		if _, err := w.WriteString(attr.Value); err != nil {
			return err
		}
		if _, err := w.WriteString(`"`); err != nil {
			return err
		}
	}

	if isVoidElement {
		// Self-closing tag
		_, err := w.WriteString(" />")
		return err
	}

	if _, err := w.WriteString(">"); err != nil {
		return err
	}

	// Content (text + children) - trim whitespace to match MRML behavior
	if node.Text != "" {
		trimmedText := strings.TrimSpace(node.Text)
		if trimmedText != "" {
			if _, err := w.WriteString(trimmedText); err != nil {
				return err
			}
		}
	}

	for _, child := range node.Children {
		if err := c.reconstructHTMLElement(child, w); err != nil {
			return err
		}
	}

	// Closing tag
	if _, err := w.WriteString("</"); err != nil {
		return err
	}
	if _, err := w.WriteString(tagName); err != nil {
		return err
	}
	_, err := w.WriteString(">")
	return err
}
