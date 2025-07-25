package components

import (
	"github.com/preslavrachev/gomjml/mjml/html"
	"github.com/preslavrachev/gomjml/parser"
)

// MJDividerComponent represents mj-divider
type MJDividerComponent struct {
	*BaseComponent
}

// NewMJDividerComponent creates a new mj-divider component
func NewMJDividerComponent(node *parser.MJMLNode) *MJDividerComponent {
	return &MJDividerComponent{
		BaseComponent: NewBaseComponent(node),
	}
}

func (c *MJDividerComponent) GetDefaultAttribute(name string) string {
	switch name {
	case "border-color":
		return "#000000"
	case "border-style":
		return "solid"
	case "border-width":
		return "4px"
	case "container-background-color":
		return "transparent"
	case "padding":
		return "10px 25px"
	case "width":
		return "100%"
	default:
		return ""
	}
}

func (c *MJDividerComponent) getAttribute(name string) string {
	return c.GetAttributeWithDefault(c, name)
}

func (c *MJDividerComponent) Render() (string, error) {
	padding := c.getAttribute("padding")
	borderColor := c.getAttribute("border-color")
	borderStyle := c.getAttribute("border-style")
	borderWidth := c.getAttribute("border-width")
	width := c.getAttribute("width")

	// Create paragraph with border styles
	p := html.NewHTMLTag("p").
		AddStyle("border-top", borderWidth+" "+borderStyle+" "+borderColor).
		AddStyle("font-size", "1px").
		AddStyle("margin", "0px auto").
		AddStyle("width", width)

	// Outer table cell with padding
	td := html.NewHTMLTag("td").
		AddStyle("font-size", "0px").
		AddStyle("padding", padding).
		AddStyle("word-break", "break-word")

	return td.RenderOpen() + p.RenderSelfClosing() + td.RenderClose(), nil
}

func (c *MJDividerComponent) GetTagName() string {
	return "mj-divider"
}