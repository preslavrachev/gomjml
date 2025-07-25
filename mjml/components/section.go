package components

import (
	"fmt"
	"strings"

	"github.com/preslavrachev/gomjml/mjml/html"
	"github.com/preslavrachev/gomjml/parser"
)

// MJSectionComponent represents mj-section
type MJSectionComponent struct {
	*BaseComponent
}

// NewMJSectionComponent creates a new mj-section component
func NewMJSectionComponent(node *parser.MJMLNode) *MJSectionComponent {
	return &MJSectionComponent{
		BaseComponent: NewBaseComponent(node),
	}
}

func (c *MJSectionComponent) Render() (string, error) {
	var output strings.Builder

	// Helper function to get attribute with default
	getAttr := func(name string) string {
		if attr := c.GetAttribute(name); attr != nil {
			return *attr
		}
		return c.GetDefaultAttribute(name)
	}

	// Get section attributes
	backgroundColor := getAttr("background-color")
	padding := getAttr("padding")
	direction := getAttr("direction")
	textAlign := getAttr("text-align")

	// MSO conditional comment - table wrapper for Outlook
	msoTable := html.NewTableTag().
		AddAttribute("align", "center").
		AddAttribute("width", fmt.Sprintf("%d", c.GetEffectiveWidth())).
		AddStyle("width", c.GetEffectiveWidthString())

	if backgroundColor != "" {
		msoTable.AddAttribute("bgcolor", backgroundColor)
	}

	msoTd := html.NewHTMLTag("td").
		AddStyle("line-height", "0px").
		AddStyle("font-size", "0px").
		AddStyle("mso-line-height-rule", "exactly")

	output.WriteString(html.RenderMSOConditional(
		msoTable.RenderOpen() + "<tr>" + msoTd.RenderOpen()))

	// Main section div with styles
	sectionDiv := html.NewHTMLTag("div")

	// Apply background styles first to match MRML order
	c.ApplyBackgroundStyles(sectionDiv)

	// Then add layout styles
	sectionDiv.AddStyle("margin", "0px auto").
		AddStyle("max-width", c.GetEffectiveWidthString())

	output.WriteString(sectionDiv.RenderOpen())

	// Inner table with styles
	innerTable := html.NewTableTag().
		AddAttribute("align", "center")

	// Apply background styles first to match MRML order
	c.ApplyBackgroundStyles(innerTable)

	// Then add width
	innerTable.AddStyle("width", "100%")

	output.WriteString(innerTable.RenderOpen())
	output.WriteString("<tbody><tr>")

	// TD with padding and text alignment
	tdTag := html.NewHTMLTag("td").
		AddStyle("direction", direction).
		AddStyle("font-size", "0px").
		AddStyle("padding", padding).
		AddStyle("text-align", textAlign)

	output.WriteString(tdTag.RenderOpen())

	// Render child columns (section provides MSO TR, columns provide MSO TDs)
	for _, child := range c.Children {
		// Pass the effective width to the child
		child.SetContainerWidth(c.GetEffectiveWidth())

		// Generate MSO conditional TD for each column (following MRML's render_wrapped_children pattern)
		if columnComp, ok := child.(*MJColumnComponent); ok {
			msoStyles := columnComp.GetMSOTDStyles()

			msoTable := html.NewTableTag()

			msoTr := html.NewHTMLTag("tr")

			msoTd := html.NewHTMLTag("td")
			for property, value := range msoStyles {
				msoTd.AddStyle(property, value)
			}

			output.WriteString(html.RenderMSOConditional(
				msoTable.RenderOpen() + msoTr.RenderOpen() + msoTd.RenderOpen()))
		}

		childHTML, err := child.Render()
		if err != nil {
			return "", err
		}
		output.WriteString(childHTML)

		// Close MSO conditional TD/TR/TABLE for columns
		if _, ok := child.(*MJColumnComponent); ok {
			output.WriteString(html.RenderMSOConditional("</td></tr></table>"))
		}
	}

	output.WriteString(tdTag.RenderClose())
	output.WriteString("</tr></tbody>")
	output.WriteString(innerTable.RenderClose())
	output.WriteString(sectionDiv.RenderClose())

	// Close MSO conditional
	output.WriteString(html.RenderMSOConditional(msoTd.RenderClose() + "</tr>" + msoTable.RenderClose()))

	return output.String(), nil
}

func (c *MJSectionComponent) GetTagName() string {
	return "mj-section"
}

func (c *MJSectionComponent) GetDefaultAttribute(name string) string {
	switch name {
	case "background-repeat":
		return "repeat"
	case "background-size":
		return "auto"
	case "direction":
		return "ltr"
	case "padding":
		return "20px 0"
	case "text-align":
		return "center"
	default:
		return ""
	}
}
