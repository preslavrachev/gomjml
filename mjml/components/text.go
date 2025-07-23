package components

import (
	"strings"

	"github.com/preslavrachev/gomjml/mjml/html"
	"github.com/preslavrachev/gomjml/parser"
)

// MJTextComponent represents mj-text
type MJTextComponent struct {
	*BaseComponent
}

// NewMJTextComponent creates a new mj-text component
func NewMJTextComponent(node *parser.MJMLNode) *MJTextComponent {
	return &MJTextComponent{
		BaseComponent: NewBaseComponent(node),
	}
}

func (c *MJTextComponent) Render() (string, error) {
	var output strings.Builder

	// Get raw text content (preserve original formatting)
	textContent := c.Node.Text

	// Helper function to get attribute with default
	getAttr := func(name string) string {
		if attr := c.GetAttribute(name); attr != nil {
			return *attr
		}
		return c.GetDefaultAttribute(name)
	}

	// Get attributes
	align := getAttr("align")
	padding := getAttr("padding")

	// Create TR element
	output.WriteString("<tr>")

	// Create TD with alignment and base styles
	tdTag := html.NewHTMLTag("td").
		AddAttribute("align", align).
		AddStyle("font-size", "0px").
		AddStyle("padding", padding).
		AddStyle("word-break", "break-word")

	output.WriteString(tdTag.RenderOpen())

	// Create inner div with font styling
	divTag := html.NewHTMLTag("div")

	// Apply font styles using the proper interface method
	fontFamily := c.GetAttributeWithDefault(c, "font-family")
	fontSize := c.GetAttributeWithDefault(c, "font-size")
	fontWeight := c.GetAttributeWithDefault(c, "font-weight")
	fontStyle := c.GetAttributeWithDefault(c, "font-style")
	color := c.GetAttributeWithDefault(c, "color")
	lineHeight := c.GetAttributeWithDefault(c, "line-height")
	textAlign := c.GetAttributeWithDefault(c, "align")
	textDecoration := c.GetAttributeWithDefault(c, "text-decoration")

	// Apply styles in the order expected by MRML
	if fontFamily != "" {
		divTag.AddStyle("font-family", fontFamily)
	}
	if fontSize != "" {
		divTag.AddStyle("font-size", fontSize)
	}
	if lineHeight != "" {
		divTag.AddStyle("line-height", lineHeight)
	}
	if textAlign != "" {
		divTag.AddStyle("text-align", textAlign)
	}
	if color != "" {
		divTag.AddStyle("color", color)
	}
	if fontWeight != "" {
		divTag.AddStyle("font-weight", fontWeight)
	}
	if fontStyle != "" {
		divTag.AddStyle("font-style", fontStyle)
	}
	if textDecoration != "" {
		divTag.AddStyle("text-decoration", textDecoration)
	}

	output.WriteString(divTag.RenderOpen())
	output.WriteString(textContent)
	output.WriteString(divTag.RenderClose())
	output.WriteString(tdTag.RenderClose())
	output.WriteString("</tr>")

	return output.String(), nil
}

func (c *MJTextComponent) GetTagName() string {
	return "mj-text"
}

func (c *MJTextComponent) GetDefaultAttribute(name string) string {
	switch name {
	case "font-size":
		return "13px"
	case "color":
		return "#000000"
	case "align":
		return "left"
	case "font-family":
		return "Ubuntu, Helvetica, Arial, sans-serif"
	case "line-height":
		return "1"
	case "padding":
		return "10px 25px"
	default:
		return ""
	}
}
