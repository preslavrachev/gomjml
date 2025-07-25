package components

import (
	"fmt"
	"strings"

	"github.com/preslavrachev/gomjml/mjml/html"
	"github.com/preslavrachev/gomjml/parser"
)

// MJGroupComponent represents mj-group - horizontal grouping of columns
type MJGroupComponent struct {
	*BaseComponent
}

// NewMJGroupComponent creates a new mj-group component
func NewMJGroupComponent(node *parser.MJMLNode) *MJGroupComponent {
	return &MJGroupComponent{
		BaseComponent: NewBaseComponent(node),
	}
}

func (c *MJGroupComponent) GetDefaultAttribute(name string) string {
	switch name {
	case "direction":
		return "ltr"
	case "vertical-align":
		return "top"
	case "width":
		return "100%"
	default:
		return ""
	}
}

func (c *MJGroupComponent) getAttribute(name string) string {
	return c.GetAttributeWithDefault(c, name)
}

func (c *MJGroupComponent) Render() (string, error) {
	var output strings.Builder

	direction := c.getAttribute("direction")
	verticalAlign := c.getAttribute("vertical-align")

	// MSO conditional for group
	output.WriteString(html.RenderMSOConditional(
		"<table border=\"0\" cellpadding=\"0\" cellspacing=\"0\" role=\"presentation\"><tr>"))

	// Render each column in the group
	for _, child := range c.Children {
		if columnComp, ok := child.(*MJColumnComponent); ok {
			// Set container width for the column
			columnComp.SetContainerWidth(c.GetContainerWidth())

			// MSO conditional TD for each column  
			output.WriteString(html.RenderMSOConditional(
				fmt.Sprintf("<td style=\"vertical-align:%s;width:%s;\">", verticalAlign, columnComp.GetEffectiveWidthString())))

			// Column div wrapper
			columnDiv := html.NewHTMLTag("div").
				AddAttribute("class", "mj-outlook-group-fix").
				AddStyle("font-size", "0px").
				AddStyle("text-align", "left").
				AddStyle("direction", direction).
				AddStyle("display", "inline-block").
				AddStyle("vertical-align", verticalAlign).
				AddStyle("width", "100%")

			output.WriteString(columnDiv.RenderOpen())

			// Render column content
			childHTML, err := child.Render()
			if err != nil {
				return "", err
			}
			output.WriteString(childHTML)

			output.WriteString(columnDiv.RenderClose())

			// Close MSO conditional TD
			output.WriteString(html.RenderMSOConditional("</td>"))
		}
	}

	// Close MSO conditional table
	output.WriteString(html.RenderMSOConditional("</tr></table>"))

	return output.String(), nil
}

func (c *MJGroupComponent) GetTagName() string {
	return "mj-group"
}