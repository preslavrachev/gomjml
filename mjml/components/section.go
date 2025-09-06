package components

import (
	"io"
	"strconv"
	"strings"

	"github.com/preslavrachev/gomjml/mjml/constants"
	"github.com/preslavrachev/gomjml/mjml/html"
	"github.com/preslavrachev/gomjml/mjml/options"
	"github.com/preslavrachev/gomjml/mjml/styles"
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
	backgroundUrl := getAttr(constants.MJMLBackgroundUrl)
	backgroundPosition := getAttr("background-position")
	backgroundPositionX := getAttr("background-position-x")
	backgroundPositionY := getAttr("background-position-y")
	backgroundRepeat := getAttr("background-repeat")
	backgroundSize := getAttr("background-size")
	padding := getAttr("padding")
	direction := getAttr("direction")
	textAlign := getAttr("text-align")
	fullWidth := getAttr("full-width")

	// Check if we have a background image for VML generation (only for full-width sections)
	hasBackgroundImage := backgroundUrl != "" && fullWidth != ""

	// For full-width sections with any background (color or image), add
	// outer table wrapper (like MRML does). Previously we only checked for
	// background color which skipped cases where only background-url was
	// provided, causing significant diffs in complex templates.
	if fullWidth != "" && (backgroundColor != "" || backgroundUrl != "") {
		outerTable := html.NewTableTag().
			AddAttribute("align", "center")

		// Apply background styles properly for full-width sections
		if backgroundUrl != "" {
			// Use shorthand and explicit longhand properties for full-width background images
			posX, posY := parseBackgroundPosition(backgroundPosition)
			posX, posY = overridePosition(posX, posY, backgroundPositionX, backgroundPositionY)
			shorthandBg := buildBackgroundShorthand(backgroundColor, backgroundUrl, posX, posY, backgroundSize, backgroundRepeat)
			if shorthandBg != "" {
				outerTable.AddStyle("background", shorthandBg)
				outerTable.AddStyle("background-position", posX+" "+posY)
				outerTable.AddStyle("background-repeat", backgroundRepeat)
				outerTable.AddStyle("background-size", backgroundSize)
				// Also add the background attribute for email client compatibility (use same encoding as VML src)
				outerTable.AddAttribute("background", htmlEscape(backgroundUrl))
			}
		} else {
			// Apply background color only
			c.ApplyBackgroundStyles(outerTable)
		}
		c.ApplyBorderStyles(outerTable)
		outerTable.AddStyle("width", "100%")

		if err := outerTable.RenderOpen(w); err != nil {
			return err
		}
		if _, err := w.WriteString("<tbody><tr><td>"); err != nil {
			return err
		}

		// Write VML opening if we have background image (inside full-width outer table TD)
		if hasBackgroundImage {
			// Parse background position
			posX, posY := parseBackgroundPosition(backgroundPosition)
			posX, posY = overridePosition(posX, posY, backgroundPositionX, backgroundPositionY)

			// Compute VML attributes
			vOriginX, vOriginY, vPosX, vPosY := computeVMLPosition(posX, posY, backgroundSize)
			vSizeAttrs, vAspect := computeVMLSize(backgroundSize)
			vmlType := computeVMLType(backgroundRepeat, backgroundSize)

			// Build VML attributes
			sizeFragment := ""
			if vSizeAttrs != "" {
				sizeFragment = " " + vSizeAttrs
			}
			aspectFragment := ""
			if vAspect != "" {
				aspectFragment = ` aspect="` + vAspect + `"`
			}

			// Build VML strings
			colorAttr := "transparent"
			if backgroundColor != "" {
				colorAttr = backgroundColor
			}

			vmlOpen := `<v:rect mso-width-percent="1000" xmlns:v="urn:schemas-microsoft-com:vml" fill="true" stroke="false"><v:fill position="` + vPosX + `, ` + vPosY + `" origin="` + vOriginX + `, ` + vOriginY +
				`" src="` + htmlEscape(backgroundUrl) + `" color="` + colorAttr + `"` +
				sizeFragment + ` type="` + vmlType + `"` +
				aspectFragment + ` /><v:textbox inset="0,0,0,0" style="mso-fit-shape-to-text:true;"><![endif]-->`

			if _, err := w.WriteString(vmlOpen); err != nil {
				return err
			}
		}
	}

	// MSO conditional comment - table wrapper for Outlook (inside VML textbox if present)
	// Compute effective MSO inner width by subtracting horizontal padding
	// from the container width to match MRML (e.g., 600 -> 560)
	effectiveWidth := c.GetEffectiveWidth()
	if padding != "" {
		if sp, err := styles.ParseSpacing(padding); err == nil && sp != nil {
			effectiveWidth -= int(sp.Left + sp.Right)
		}
	}
	if pa := c.GetAttribute(constants.MJMLPaddingLeft); pa != nil && *pa != "" {
		if px, err := styles.ParsePixel(*pa); err == nil && px != nil {
			if sp, err := styles.ParseSpacing(padding); err == nil && sp != nil {
				effectiveWidth += int(sp.Left)
			}
			effectiveWidth -= int(px.Value)
		}
	}
	if pa := c.GetAttribute(constants.MJMLPaddingRight); pa != nil && *pa != "" {
		if px, err := styles.ParsePixel(*pa); err == nil && px != nil {
			if sp, err := styles.ParseSpacing(padding); err == nil && sp != nil {
				effectiveWidth += int(sp.Right)
			}
			effectiveWidth -= int(px.Value)
		}
	}
	if effectiveWidth < 0 {
		effectiveWidth = 0
	}

	// Get align from attributes (including mj-class)
	alignAttr := getAttr("align")
	if alignAttr == "" {
		alignAttr = "center" // default align for MSO table
	}

	msoTable := html.NewTableTag().
		AddAttribute("align", alignAttr).
		AddAttribute("width", strconv.Itoa(effectiveWidth)).
		AddStyle("width", getPixelWidthString(effectiveWidth))

	// Add css-class-outlook if present
	if cssClass := c.GetCSSClass(); cssClass != "" {
		msoTable.AddAttribute("class", cssClass+"-outlook")
	}

	msoTd := html.NewHTMLTag("td").
		AddStyle("line-height", "0px").
		AddStyle("font-size", "0px").
		AddStyle("mso-line-height-rule", "exactly")

	// Custom MSO conditional
	if _, err := w.WriteString("<!--[if mso | IE]>"); err != nil {
		return err
	}
	if err := msoTable.RenderOpen(w); err != nil {
		return err
	}
	if _, err := w.WriteString("<tr>"); err != nil {
		return err
	}
	if err := msoTd.RenderOpen(w); err != nil {
		return err
	}

	if _, err := w.WriteString("<![endif]-->"); err != nil {
		return err
	}

	// Main section div with styles
	sectionDiv := html.NewHTMLTag("div")
	c.AddDebugAttribute(sectionDiv, "section")

	// Add css-class if present
	if cssClass := c.BuildClassAttribute(); cssClass != "" {
		sectionDiv.AddAttribute("class", cssClass)
	}

	// Background on main section div (MRML behavior):
	// - When not full-width and we have a background image, use shorthand background
	//   and explicitly set position/repeat/size (no extra longhands for color/image).
	// - When only background color is present (no image) and not full-width, apply color.
	if fullWidth == "" {
		if backgroundUrl != "" {
			posX, posY := parseBackgroundPosition(backgroundPosition)
			posX, posY = overridePosition(posX, posY, backgroundPositionX, backgroundPositionY)
			shorthandBg := buildBackgroundShorthand(backgroundColor, backgroundUrl, posX, posY, backgroundSize, backgroundRepeat)
			if shorthandBg != "" {
				sectionDiv.AddStyle("background", shorthandBg)
				// Add the explicit longhand properties to match MRML output
				sectionDiv.AddStyle("background-position", posX+" "+posY)
				sectionDiv.AddStyle("background-repeat", backgroundRepeat)
				sectionDiv.AddStyle("background-size", backgroundSize)
			}
		} else if backgroundColor != "" {
			// Color-only background
			c.ApplyBackgroundStyles(sectionDiv)
		}
	}

	// Add layout styles - use actual container width (respects wrapper padding)
	sectionDiv.AddStyle("margin", "0px auto").
		AddStyle("max-width", strconv.Itoa(c.GetContainerWidth())+"px")

	// Add border-radius if specified
	if borderRadius := getAttr("border-radius"); borderRadius != "" {
		sectionDiv.AddStyle("border-radius", borderRadius)
	}

	if err := sectionDiv.RenderOpen(w); err != nil {
		return err
	}

	// Add intermediate div wrapper when we have background image (matches MRML structure)
	var intermediateDiv *html.HTMLTag
	if backgroundUrl != "" {
		intermediateDiv = html.NewHTMLTag("div").
			AddStyle("line-height", "0").
			// Match MRML: font-size should be 0 (unitless), not 0px
			AddStyle("font-size", "0")
		if err := intermediateDiv.RenderOpen(w); err != nil {
			return err
		}
	}

	// Inner table with styles
	innerTable := html.NewTableTag().
		AddAttribute("align", "center")

	// Apply background styles to inner table (only for non-full-width sections)
	if fullWidth == "" && backgroundUrl != "" {
		// Use shorthand and explicit longhand properties (avoid extra background-color/image longhands)
		posX, posY := parseBackgroundPosition(backgroundPosition)
		posX, posY = overridePosition(posX, posY, backgroundPositionX, backgroundPositionY)
		shorthandBg := buildBackgroundShorthand(backgroundColor, backgroundUrl, posX, posY, backgroundSize, backgroundRepeat)
		if shorthandBg != "" {
			innerTable.AddStyle("background", shorthandBg)
			innerTable.AddStyle("background-position", posX+" "+posY)
			innerTable.AddStyle("background-repeat", backgroundRepeat)
			innerTable.AddStyle("background-size", backgroundSize)
			// Also add the background attribute for email client compatibility (use same encoding as VML src)
			innerTable.AddAttribute("background", htmlEscape(backgroundUrl))
		}
	} else if fullWidth == "" {
		// No background image: apply defaults (color-only etc.)
		if backgroundColor == "" || (backgroundColor != "" && fullWidth == "") {
			c.ApplyBackgroundStyles(innerTable)
		}
	}

	// Then add width and border-radius
	innerTable.AddStyle("width", "100%")

	// Add border-radius if specified
	if borderRadius := getAttr("border-radius"); borderRadius != "" {
		innerTable.AddStyle("border-radius", borderRadius)
	}

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

	// Add specific padding overrides in MRML order: left, right, bottom, top
	if paddingLeftAttr := c.GetAttribute(constants.MJMLPaddingLeft); paddingLeftAttr != nil {
		tdTag.AddStyle(constants.CSSPaddingLeft, *paddingLeftAttr)
	}
	if paddingRightAttr := c.GetAttribute(constants.MJMLPaddingRight); paddingRightAttr != nil {
		tdTag.AddStyle(constants.CSSPaddingRight, *paddingRightAttr)
	}
	if paddingBottomAttr := c.GetAttribute(constants.MJMLPaddingBottom); paddingBottomAttr != nil {
		tdTag.AddStyle(constants.CSSPaddingBottom, *paddingBottomAttr)
	}
	if paddingTopAttr := c.GetAttribute(constants.MJMLPaddingTop); paddingTopAttr != nil {
		tdTag.AddStyle(constants.CSSPaddingTop, *paddingTopAttr)
	}

	tdTag.AddStyle("text-align", textAlign)

	if err := tdTag.RenderOpen(w); err != nil {
		return err
	}

	// No inner MSO wrapper here; per MRML, inner column MSO wrappers are handled separately

	// Calculate sibling counts for width calculations (following MRML logic)
	siblings := len(c.Children)
	rawSiblings := 0
	for _, child := range c.Children {
		if child.IsRawElement() {
			rawSiblings++
		}
	}

	// Render child columns and groups (section provides MSO TR, columns provide MSO TDs)
	// AIDEV-NOTE: width-flow-start; section initiates width flow by passing effective width to columns
	for _, child := range c.Children {
		if child.IsRawElement() {
			if err := child.Render(w); err != nil {
				return err
			}
			continue
		}

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

	// Close intermediate div if we added one
	if intermediateDiv != nil {
		if err := intermediateDiv.RenderClose(w); err != nil {
			return err
		}
	}

	if err := sectionDiv.RenderClose(w); err != nil {
		return err
	}

	// Close MSO table structure
	if _, err := w.WriteString("<!--[if mso | IE]></td></tr></table><![endif]-->"); err != nil {
		return err
	}

	// Close outer table if we added one for full-width background
	if fullWidth != "" && (backgroundColor != "" || backgroundUrl != "") {
		// Close VML first if present, then outer table
		if hasBackgroundImage {
			if _, err := w.WriteString("<!--[if mso | IE]></v:textbox></v:rect>"); err != nil {
				return err
			}
		}
		if _, err := w.WriteString("</td></tr></tbody></table>"); err != nil {
			return err
		}
	}

	return nil
}

func (c *MJSectionComponent) GetDefaultAttribute(name string) string {
	switch name {
	case "background-color":
		return ""
	case constants.MJMLBackgroundUrl:
		return ""
	case "background-position":
		return "top center"
	case "background-position-x":
		return ""
	case "background-position-y":
		return ""
	case "background-repeat":
		return "repeat"
	case "background-size":
		return "auto"
	case "direction":
		return "ltr"
	case "full-width":
		return ""
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
