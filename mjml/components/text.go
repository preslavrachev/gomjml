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

	// Get raw inner HTML content (preserve HTML tags and formatting)
	textContent := c.getRawInnerHTML()

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
		AddStyle("padding", padding)
	
	// Add specific padding overrides if they exist (following MRML/section pattern)
	if paddingTopAttr := c.GetAttribute("padding-top"); paddingTopAttr != nil {
		tdTag.AddStyle("padding-top", *paddingTopAttr)
	}
	if paddingBottomAttr := c.GetAttribute("padding-bottom"); paddingBottomAttr != nil {
		tdTag.AddStyle("padding-bottom", *paddingBottomAttr)
	}
	if paddingLeftAttr := c.GetAttribute("padding-left"); paddingLeftAttr != nil {
		tdTag.AddStyle("padding-left", *paddingLeftAttr)
	}
	if paddingRightAttr := c.GetAttribute("padding-right"); paddingRightAttr != nil {
		tdTag.AddStyle("padding-right", *paddingRightAttr)
	}
	
	tdTag.AddStyle("word-break", "break-word")

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

// getRawInnerHTML reconstructs the original inner HTML content of the mj-text element
// This is needed because our parser splits content, but mj-text needs to preserve HTML
func (c *MJTextComponent) getRawInnerHTML() string {
	// If we have children (HTML elements), we need to reconstruct the original HTML
	if len(c.Node.Children) > 0 {
		var html strings.Builder
		
		// Add any text content before children
		if c.Node.Text != "" {
			html.WriteString(c.Node.Text)
		}
		
		// Add children as HTML elements
		for _, child := range c.Node.Children {
			html.WriteString(c.reconstructHTMLElement(child))
		}
		
		return html.String()
	}
	
	// If no children, just return the text content
	return c.Node.Text
}

// reconstructHTMLElement reconstructs an HTML element from a parsed node
func (c *MJTextComponent) reconstructHTMLElement(node *parser.MJMLNode) string {
	var html strings.Builder
	
	// Opening tag
	html.WriteString("<")
	html.WriteString(node.XMLName.Local)
	
	// Attributes
	for _, attr := range node.Attrs {
		html.WriteString(" ")
		html.WriteString(attr.Name.Local)
		html.WriteString(`="`)
		html.WriteString(attr.Value)
		html.WriteString(`"`)
	}
	
	html.WriteString(">")
	
	// Content (text + children)
	if node.Text != "" {
		html.WriteString(node.Text)
	}
	
	for _, child := range node.Children {
		html.WriteString(c.reconstructHTMLElement(child))
	}
	
	// Closing tag
	html.WriteString("</")
	html.WriteString(node.XMLName.Local)
	html.WriteString(">")
	
	return html.String()
}
