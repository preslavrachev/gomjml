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
	case "align":
		return "center"
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
	align := c.getAttribute("align")

	// Calculate margin based on alignment (matching MRML logic)
	var margin string
	switch align {
	case "left":
		margin = "0px"
	case "right":
		margin = "0px 0px 0px auto"
	default:
		margin = "0px auto"
	}

	// Create paragraph with border styles matching MRML exact order
	p := html.NewHTMLTag("p")
	c.AddDebugAttribute(p, "divider")
	p.
		AddStyle("border-top", borderStyle+" "+borderWidth+" "+borderColor).
		AddStyle("font-size", "1px").
		AddStyle("margin", margin)

	// Add width (MRML includes default width of 100%)
	width := c.getAttribute("width")
	p = p.AddStyle("width", width)

	// Table cell with padding and center alignment
	td := html.NewHTMLTag("td").
		AddAttribute("align", "center").
		AddStyle("font-size", "0px").
		AddStyle("padding", padding).
		AddStyle("word-break", "break-word")

	// MSO conditional comment for Outlook compatibility
	msoTable := `<!--[if mso | IE]><table border="0" cellpadding="0" cellspacing="0" role="presentation" align="center" width="550px" style="border-top:` + borderStyle + ` ` + borderWidth + ` ` + borderColor + `;font-size:1px;margin:0px auto;width:550px;"><tr><td style="height:0;line-height:0;">&nbsp;</td></tr></table><![endif]-->`

	// Render complete table row - paragraph must be empty, not self-closing to match MRML
	return "<tr>" + td.RenderOpen() + p.RenderOpen() + p.RenderClose() + msoTable + td.RenderClose() + "</tr>", nil
}

func (c *MJDividerComponent) GetTagName() string {
	return "mj-divider"
}
