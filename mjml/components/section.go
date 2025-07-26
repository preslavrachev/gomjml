package components

import (
	"fmt"
	"strings"

	"github.com/preslavrachev/gomjml/mjml/html"
	"github.com/preslavrachev/gomjml/mjml/options"
	"github.com/preslavrachev/gomjml/parser"
)

// MJSectionComponent represents mj-section
type MJSectionComponent struct {
	*BaseComponent
}

// NewMJSectionComponent creates a new mj-section component
func NewMJSectionComponent(node *parser.MJMLNode, opts *options.RenderOpts) *MJSectionComponent {
	return &MJSectionComponent{
		BaseComponent: NewBaseComponent(node, opts),
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
	fullWidth := getAttr("full-width")

	// For full-width sections with background, add outer table wrapper (like MRML does)
	if backgroundColor != "" && fullWidth != "" {
		outerTable := html.NewTableTag().
			AddAttribute("align", "center")

		// Apply background styles in MRML order: background, background-color, width
		c.ApplyBackgroundStyles(outerTable)
		outerTable.AddStyle("width", "100%")

		output.WriteString(outerTable.RenderOpen())
		output.WriteString("<tbody><tr><td>")
	}

	// MSO conditional comment - table wrapper for Outlook
	msoTable := html.NewTableTag()

	// Add attributes in MRML order: bgcolor, align, width
	if backgroundColor != "" {
		msoTable.AddAttribute("bgcolor", backgroundColor)
	}

	msoTable.AddAttribute("align", "center").
		AddAttribute("width", fmt.Sprintf("%d", c.GetEffectiveWidth())).
		AddStyle("width", c.GetEffectiveWidthString())

	msoTd := html.NewHTMLTag("td").
		AddStyle("line-height", "0px").
		AddStyle("font-size", "0px").
		AddStyle("mso-line-height-rule", "exactly")

	output.WriteString(html.RenderMSOConditional(
		msoTable.RenderOpen() + "<tr>" + msoTd.RenderOpen()))

	// Main section div with styles
	sectionDiv := html.NewHTMLTag("div")
	c.AddDebugAttribute(sectionDiv, "section")

	// For non-full-width background sections, apply background to the div (like MRML)
	if backgroundColor != "" && fullWidth == "" {
		c.ApplyBackgroundStyles(sectionDiv)
	}

	// Add layout styles
	sectionDiv.AddStyle("margin", "0px auto").
		AddStyle("max-width", c.GetEffectiveWidthString())

	output.WriteString(sectionDiv.RenderOpen())

	// Inner table with styles
	innerTable := html.NewTableTag().
		AddAttribute("align", "center")

	// Apply background styles to inner table
	// - Always for no-background sections
	// - Also for non-full-width background sections (MRML puts background on both div and table)
	if backgroundColor == "" || (backgroundColor != "" && fullWidth == "") {
		c.ApplyBackgroundStyles(innerTable)
	}

	// Then add width
	innerTable.AddStyle("width", "100%")

	output.WriteString(innerTable.RenderOpen())
	output.WriteString("<tbody><tr>")

	// TD with padding and text alignment
	tdTag := html.NewHTMLTag("td").
		AddStyle("direction", direction).
		AddStyle("font-size", "0px").
		AddStyle("padding", padding)

	// Add specific padding overrides in MRML order: left, right, top, bottom
	if paddingLeftAttr := c.GetAttribute("padding-left"); paddingLeftAttr != nil {
		tdTag.AddStyle("padding-left", *paddingLeftAttr)
	}
	if paddingRightAttr := c.GetAttribute("padding-right"); paddingRightAttr != nil {
		tdTag.AddStyle("padding-right", *paddingRightAttr)
	}
	if paddingTopAttr := c.GetAttribute("padding-top"); paddingTopAttr != nil {
		tdTag.AddStyle("padding-top", *paddingTopAttr)
	}
	if paddingBottomAttr := c.GetAttribute("padding-bottom"); paddingBottomAttr != nil {
		tdTag.AddStyle("padding-bottom", *paddingBottomAttr)
	}

	tdTag.AddStyle("text-align", textAlign)

	output.WriteString(tdTag.RenderOpen())

	// Calculate sibling counts for width calculations (following MRML logic)
	siblings := len(c.Children)
	rawSiblings := 0
	for _, child := range c.Children {
		// Count raw siblings (components that don't participate in width calculations)
		// For now, all our components are non-raw, but this matches MRML structure
		if child.GetTagName() == "mj-raw" {
			rawSiblings++
		}
	}

	// Render child columns and groups (section provides MSO TR, columns provide MSO TDs)
	for _, child := range c.Children {
		// Pass the effective width and sibling counts to the child
		child.SetContainerWidth(c.GetEffectiveWidth())
		child.SetSiblings(siblings)
		child.SetRawSiblings(rawSiblings)

		// Generate MSO conditional TD for each column (following MRML's render_wrapped_children pattern)
		if columnComp, ok := child.(*MJColumnComponent); ok {
			msoTable := html.NewTableTag()

			msoTr := html.NewHTMLTag("tr")

			msoTd := html.NewHTMLTag("td")
			// Add styles in MRML insertion order: vertical-align first, then width
			getAttr := func(name string) string {
				if attr := columnComp.GetAttribute(name); attr != nil {
					return *attr
				}
				return columnComp.GetDefaultAttribute(name)
			}
			msoTd.AddStyle("vertical-align", getAttr("vertical-align"))
			msoTd.AddStyle("width", columnComp.GetWidthAsPixel())

			output.WriteString(html.RenderMSOConditional(
				msoTable.RenderOpen() + msoTr.RenderOpen() + msoTd.RenderOpen()))
		} else if groupComp, ok := child.(*MJGroupComponent); ok {
			// Groups handle their own MSO structure, just set container width
			groupComp.SetContainerWidth(c.GetEffectiveWidth())
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

	// Close outer table if we added one for full-width background
	if backgroundColor != "" && fullWidth != "" {
		output.WriteString("</td></tr></tbody></table>")
	}

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
