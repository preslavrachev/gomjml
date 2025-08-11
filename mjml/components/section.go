package components

import (
	"io"
	"strconv"
	"strings"

	"github.com/preslavrachev/gomjml/mjml/constants"
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

func (c *MJSectionComponent) GetTagName() string {
	return "mj-section"
}

// Render implements optimized Writer-based rendering for MJSectionComponent
func (c *MJSectionComponent) Render(w io.StringWriter) error {
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

		if err := outerTable.RenderOpen(w); err != nil {
			return err
		}
		if _, err := w.WriteString("<tbody><tr><td>"); err != nil {
			return err
		}
	}

	// MSO conditional comment - table wrapper for Outlook
	msoTable := html.NewTableTag()

	// Add attributes in MRML order: bgcolor, align, width
	if backgroundColor != "" {
		msoTable.AddAttribute(constants.AttrBgcolor, backgroundColor)
	}

	msoTable.AddAttribute("align", "center").
		AddAttribute("width", strconv.Itoa(c.GetEffectiveWidth())).
		AddStyle("width", c.GetEffectiveWidthString())

	// Add css-class-outlook if present
	if cssClass := c.GetCSSClass(); cssClass != "" {
		msoTable.AddAttribute("class", cssClass+"-outlook")
	}

	msoTd := html.NewHTMLTag("td").
		AddStyle("line-height", "0px").
		AddStyle("font-size", "0px").
		AddStyle("mso-line-height-rule", "exactly")

	if err := html.RenderMSOTableOpenConditional(w, msoTable, msoTd); err != nil {
		return err
	}

	// Main section div with styles
	sectionDiv := html.NewHTMLTag("div")
	c.AddDebugAttribute(sectionDiv, "section")

	// Add css-class if present
	if cssClass := c.BuildClassAttribute(); cssClass != "" {
		sectionDiv.AddAttribute("class", cssClass)
	}

	// For non-full-width background sections, apply background to the div (like MRML)
	if backgroundColor != "" && fullWidth == "" {
		c.ApplyBackgroundStyles(sectionDiv)
	}

	// Add layout styles
	sectionDiv.AddStyle("margin", "0px auto").
		AddStyle("max-width", c.GetEffectiveWidthString())

	if err := sectionDiv.RenderOpen(w); err != nil {
		return err
	}

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

	if err := innerTable.RenderOpen(w); err != nil {
		return err
	}
	if _, err := w.WriteString("<tbody><tr>"); err != nil {
		return err
	}

	// TD with padding and text alignment
	tdTag := html.NewHTMLTag("td").
		AddStyle("direction", direction).
		AddStyle("font-size", "0px").
		AddStyle("padding", padding)

	// Add specific padding overrides in MRML order: left, right, top, bottom
	if paddingLeftAttr := c.GetAttribute(constants.MJMLPaddingLeft); paddingLeftAttr != nil {
		tdTag.AddStyle(constants.CSSPaddingLeft, *paddingLeftAttr)
	}
	if paddingRightAttr := c.GetAttribute(constants.MJMLPaddingRight); paddingRightAttr != nil {
		tdTag.AddStyle(constants.CSSPaddingRight, *paddingRightAttr)
	}
	if paddingTopAttr := c.GetAttribute(constants.MJMLPaddingTop); paddingTopAttr != nil {
		tdTag.AddStyle(constants.CSSPaddingTop, *paddingTopAttr)
	}
	if paddingBottomAttr := c.GetAttribute(constants.MJMLPaddingBottom); paddingBottomAttr != nil {
		tdTag.AddStyle(constants.CSSPaddingBottom, *paddingBottomAttr)
	}

	tdTag.AddStyle("text-align", textAlign)

	if err := tdTag.RenderOpen(w); err != nil {
		return err
	}

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

			if err := html.RenderMSOTableTrOpenConditional(w, msoTable, msoTr, msoTd); err != nil {
				return err
			}
		} else if groupComp, ok := child.(*MJGroupComponent); ok {
			// Groups also need MSO conditionals like columns
			groupComp.SetContainerWidth(c.GetEffectiveWidth())

			msoTable := html.NewTableTag()
			msoTr := html.NewHTMLTag("tr")
			msoTd := html.NewHTMLTag("td")

			// Use group's specific width if it has one, otherwise use section's effective width
			groupWidth := "100%" // default
			if groupComp.GetAttribute("width") != nil {
				groupWidth = *groupComp.GetAttribute("width")
			}

			if strings.HasSuffix(groupWidth, "px") {
				// Use the group's pixel width directly
				msoTd.AddStyle("width", groupWidth)
			} else {
				// Use section's effective width for percentage-based groups
				msoTd.AddStyle("width", c.GetEffectiveWidthString())
			}

			if err := html.RenderMSOTableTrOpenConditional(w, msoTable, msoTr, msoTd); err != nil {
				return err
			}
		}

		// Use optimized rendering with fallback to string-based
		if err := child.Render(w); err != nil {
			return err
		}

		// Close MSO conditional TD/TR/TABLE for columns and groups
		if _, ok := child.(*MJColumnComponent); ok {
			if err := html.RenderMSOGroupTableClose(w); err != nil {
				return err
			}
		} else if _, ok := child.(*MJGroupComponent); ok {
			if err := html.RenderMSOGroupTableClose(w); err != nil {
				return err
			}
		}
	}

	if err := tdTag.RenderClose(w); err != nil {
		return err
	}
	if _, err := w.WriteString("</tr></tbody>"); err != nil {
		return err
	}
	if err := innerTable.RenderClose(w); err != nil {
		return err
	}
	if err := sectionDiv.RenderClose(w); err != nil {
		return err
	}

	// Close MSO conditional
	if err := html.RenderMSOTableCloseConditional(w, msoTd, msoTable); err != nil {
		return err
	}

	// Close outer table if we added one for full-width background
	if backgroundColor != "" && fullWidth != "" {
		if _, err := w.WriteString("</td></tr></tbody></table>"); err != nil {
			return err
		}
	}

	return nil
}

func (c *MJSectionComponent) GetDefaultAttribute(name string) string {
	switch name {
	case "background-position":
		return "top center"
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
	case "text-padding":
		return "4px 4px 4px 0"
	default:
		return ""
	}
}
