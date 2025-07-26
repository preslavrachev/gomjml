package components

import (
	"fmt"
	"strings"

	"github.com/preslavrachev/gomjml/mjml/html"
	"github.com/preslavrachev/gomjml/mjml/options"
	"github.com/preslavrachev/gomjml/parser"
)

// MJGroupComponent represents mj-group - horizontal grouping of columns
type MJGroupComponent struct {
	*BaseComponent
}

// NewMJGroupComponent creates a new mj-group component
func NewMJGroupComponent(node *parser.MJMLNode, opts *options.RenderOpts) *MJGroupComponent {
	return &MJGroupComponent{
		BaseComponent: NewBaseComponent(node, opts),
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
	backgroundColor := c.getAttribute("background-color")

	// Calculate CSS class based on width (following MRML pattern)
	width := c.getAttribute("width")
	var cssClass string
	if width == "" || width == "100%" {
		cssClass = "mj-column-per-100"
	} else {
		// For other widths, create appropriate class
		cssClass = fmt.Sprintf("mj-column-per-%s", strings.ReplaceAll(width, "%", ""))
	}

	// Root div wrapper (following MRML set_style_root_div)
	// Note: Class order should be "mj-column-per-100 mj-outlook-group-fix" to match MRML
	rootDiv := html.NewHTMLTag("div").
		AddAttribute("class", fmt.Sprintf("%s mj-outlook-group-fix", cssClass)).
		AddStyle("font-size", "0"). // Note: "0" not "0px" to match MRML
		AddStyle("line-height", "0").
		AddStyle("text-align", "left").
		AddStyle("display", "inline-block").
		AddStyle("width", "100%").
		AddStyle("direction", direction)

	if verticalAlign != "" {
		rootDiv.AddStyle("vertical-align", verticalAlign)
	}
	if backgroundColor != "" {
		rootDiv.AddStyle("background-color", backgroundColor)
	}

	output.WriteString(rootDiv.RenderOpen())

	// MSO conditional table structure
	output.WriteString(html.RenderMSOConditional(
		"<table border=\"0\" cellpadding=\"0\" cellspacing=\"0\" role=\"presentation\"><tr>"))

	// Render each column in the group
	for _, child := range c.Children {
		if columnComp, ok := child.(*MJColumnComponent); ok {
			// Set container width for the column
			columnComp.SetContainerWidth(c.GetContainerWidth())

			// MSO conditional TD for each column (following MRML render_children pattern)
			output.WriteString(html.RenderMSOConditional(
				fmt.Sprintf("<td style=\"vertical-align:%s;width:%s;\">", verticalAlign, columnComp.GetEffectiveWidthString())))

			// Render column content with additional table wrapper for group context
			childHTML, err := child.Render()
			if err != nil {
				return "", err
			}

			// Wrap column in additional table with padding support
			// This matches MRML's group column structure
			output.WriteString(`<table border="0" cellpadding="0" cellspacing="0" role="presentation" width="100%"><tbody><tr>`)

			// TD with padding support - extract padding from column attributes
			tdStyle := fmt.Sprintf("vertical-align:%s", verticalAlign)
			if paddingRightAttr := columnComp.GetAttribute("padding-right"); paddingRightAttr != nil {
				tdStyle += fmt.Sprintf(";padding-right:%s", *paddingRightAttr)
			}
			if paddingAttr := columnComp.GetAttribute("padding"); paddingAttr != nil {
				tdStyle += fmt.Sprintf(";padding:%s", *paddingAttr)
			}

			output.WriteString(fmt.Sprintf(`<td style="%s;">`, tdStyle))
			output.WriteString(childHTML)
			output.WriteString(`</td></tr></tbody></table>`)

			// Close MSO conditional TD
			output.WriteString(html.RenderMSOConditional("</td>"))
		}
	}

	// Close MSO conditional table
	output.WriteString(html.RenderMSOConditional("</tr></table>"))

	// Close root div
	output.WriteString(rootDiv.RenderClose())

	return output.String(), nil
}

func (c *MJGroupComponent) GetTagName() string {
	return "mj-group"
}
