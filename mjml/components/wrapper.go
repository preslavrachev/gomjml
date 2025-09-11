package components

import (
	"io"
	"strconv"

	"github.com/preslavrachev/gomjml/mjml/constants"
	"github.com/preslavrachev/gomjml/mjml/html"
	"github.com/preslavrachev/gomjml/mjml/options"
	"github.com/preslavrachev/gomjml/mjml/styles"
	"github.com/preslavrachev/gomjml/parser"
)

// MJWrapperComponent represents mj-wrapper
type MJWrapperComponent struct {
	*BaseComponent
}

// NewMJWrapperComponent creates a new mj-wrapper component
func NewMJWrapperComponent(node *parser.MJMLNode, opts *options.RenderOpts) *MJWrapperComponent {
	return &MJWrapperComponent{
		BaseComponent: NewBaseComponent(node, opts),
	}
}

func (c *MJWrapperComponent) GetDefaultAttribute(name string) string {
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

func (c *MJWrapperComponent) getAttribute(name string) string {
	return c.GetAttributeWithDefault(c, name)
}

// getBorderWidth calculates total horizontal border width taking into account
// shorthand border, border-left, and border-right overrides.
func (c *MJWrapperComponent) getBorderWidth() int {
	left, right := c.getBorderLRWidths()
	return left + right
}

// getBorderLRWidths returns individual left and right border widths in pixels.
func (c *MJWrapperComponent) getBorderLRWidths() (int, int) {
	var left, right int
	if border := c.getAttribute("border"); border != "" {
		w := styles.ParseBorderWidth(border)
		left, right = w, w
	}
	if bl := c.getAttribute("border-left"); bl != "" {
		if w := styles.ParseBorderWidth(bl); w > 0 {
			left = w
		}
	}
	if br := c.getAttribute("border-right"); br != "" {
		if w := styles.ParseBorderWidth(br); w > 0 {
			right = w
		}
	}
	return left, right
}

// getEffectiveWidth calculates width minus border width
func (c *MJWrapperComponent) getEffectiveWidth() int {
	baseWidth := GetDefaultBodyWidthPixels()
	borderLeft, borderRight := c.getBorderLRWidths()
	effectiveWidth := baseWidth - borderLeft - borderRight

	// AIDEV-NOTE: wrapper-width-flow; wrapper padding reduces child containerWidth
	// Subtract horizontal padding (handle both shorthand and individual properties)
	// This ensures child sections receive reduced containerWidth accounting for wrapper padding
	padding := c.getAttribute("padding")
	if padding != "" {
		if sp, err := styles.ParseSpacing(padding); err == nil && sp != nil {
			effectiveWidth -= int(sp.Left + sp.Right)
		}
	}

	// Individual properties override shorthand
	if pl := c.getAttribute(constants.MJMLPaddingLeft); pl != "" {
		if px, err := styles.ParsePixel(pl); err == nil && px != nil {
			// If we already subtracted from shorthand, add it back first
			if padding != "" {
				if sp, err := styles.ParseSpacing(padding); err == nil && sp != nil {
					effectiveWidth += int(sp.Left)
				}
			}
			effectiveWidth -= int(px.Value)
		}
	}
	if pr := c.getAttribute(constants.MJMLPaddingRight); pr != "" {
		if px, err := styles.ParsePixel(pr); err == nil && px != nil {
			// If we already subtracted from shorthand, add it back first
			if padding != "" {
				if sp, err := styles.ParseSpacing(padding); err == nil && sp != nil {
					effectiveWidth += int(sp.Right)
				}
			}
			effectiveWidth -= int(px.Value)
		}
	}

	return effectiveWidth
}

func (c *MJWrapperComponent) isFullWidth() bool {
	// Full width only if explicitly set
	return c.getAttribute("full-width") == "full-width"
}

// Render implements optimized Writer-based rendering for MJWrapperComponent
func (c *MJWrapperComponent) Render(w io.StringWriter) error {
	if c.isFullWidth() {
		return c.renderFullWidthToWriter(w)
	}
	return c.renderSimpleToWriter(w)
}

// renderFullWidthToWriter writes full-width wrapper directly to Writer
func (c *MJWrapperComponent) renderFullWidthToWriter(w io.StringWriter) error {
	// Get wrapper attributes
	padding := c.getAttribute("padding")
	textAlign := c.getAttribute("text-align")
	direction := c.getAttribute("direction")

	// Calculate effective content width by subtracting horizontal padding and border widths
	effectiveWidth := GetDefaultBodyWidthPixels() - c.getBorderWidth()
	if pl := c.getAttribute(constants.MJMLPaddingLeft); pl != "" {
		if px, err := styles.ParsePixel(pl); err == nil && px != nil {
			effectiveWidth -= int(px.Value)
		}
	}
	if pr := c.getAttribute(constants.MJMLPaddingRight); pr != "" {
		if px, err := styles.ParsePixel(pr); err == nil && px != nil {
			effectiveWidth -= int(px.Value)
		}
	}

	// Outer full-width table (MRML pattern)
	outerTable := html.NewHTMLTag("table").
		AddAttribute("border", "0").
		AddAttribute("cellpadding", "0").
		AddAttribute("cellspacing", "0").
		AddAttribute("role", "presentation").
		AddAttribute("align", "center")

	if cssClass := c.getAttribute("css-class"); cssClass != "" {
		outerTable.AddAttribute("class", cssClass)
	}

	// Apply background styles to outer table and add width:100%
	c.ApplyBackgroundStyles(outerTable, c)
	outerTable.AddStyle("width", "100%")

	if err := outerTable.RenderOpen(w); err != nil {
		return err
	}
	if _, err := w.WriteString("<tbody><tr><td>"); err != nil {
		return err
	}

	// MSO conditional for inner container
	msoTable := html.NewHTMLTag("table").
		AddAttribute("border", "0").
		AddAttribute("cellpadding", "0").
		AddAttribute("cellspacing", "0").
		AddAttribute("role", "presentation")

	// Add bgcolor to MSO table if background-color is set
	if bgColor := c.getAttribute("background-color"); bgColor != "" {
		msoTable.AddAttribute(constants.AttrBgcolor, bgColor)
	}

	msoTable.AddAttribute("align", "center").
		AddAttribute("width", strconv.Itoa(GetDefaultBodyWidthPixels())).
		AddStyle("width", GetDefaultBodyWidth())

	// Add css-class support for MSO table (MRML adds -outlook suffix)
	if cssClass := c.getAttribute("css-class"); cssClass != "" {
		msoTable.AddAttribute("class", cssClass+"-outlook")
	}

	msoTd := html.NewHTMLTag("td").
		AddStyle("line-height", "0px").
		AddStyle("font-size", "0px").
		AddStyle("mso-line-height-rule", "exactly")

	if err := html.RenderMSOTableOpenConditional(w, msoTable, msoTd); err != nil {
		return err
	}

	// Inner constrained div (standard MRML pattern)
	innerDiv := html.NewHTMLTag("div").
		AddStyle("margin", "0px auto").
		AddStyle("max-width", GetDefaultBodyWidth())

	if err := innerDiv.RenderOpen(w); err != nil {
		return err
	}

	// Inner table with content
	innerTable := html.NewHTMLTag("table").
		AddAttribute("border", "0").
		AddAttribute("cellpadding", "0").
		AddAttribute("cellspacing", "0").
		AddAttribute("role", "presentation").
		AddAttribute("align", "center").
		AddStyle("width", "100%")

	if err := innerTable.RenderOpen(w); err != nil {
		return err
	}
	if _, err := w.WriteString("<tbody><tr>"); err != nil {
		return err
	}

	// Inner TD
	innerTd := html.NewHTMLTag("td")

	if border := c.getAttribute("border"); border != "" {
		innerTd.AddStyle("border", border)
	}
	if borderBottom := c.getAttribute("border-bottom"); borderBottom != "" {
		innerTd.AddStyle("border-bottom", borderBottom)
	}
	if borderLeft := c.getAttribute("border-left"); borderLeft != "" {
		innerTd.AddStyle("border-left", borderLeft)
	}
	if borderRight := c.getAttribute("border-right"); borderRight != "" {
		innerTd.AddStyle("border-right", borderRight)
	}
	if borderTop := c.getAttribute("border-top"); borderTop != "" {
		innerTd.AddStyle("border-top", borderTop)
	}

	innerTd.AddStyle("direction", direction).
		AddStyle("font-size", "0px").
		AddStyle("padding", padding)

	// Add individual padding properties after shorthand to match MRML order:
	// bottom, left, right, then top.
	if paddingBottom := c.getAttribute(constants.MJMLPaddingBottom); paddingBottom != "" {
		innerTd.AddStyle(constants.CSSPaddingBottom, paddingBottom)
	}
	if paddingLeft := c.getAttribute(constants.MJMLPaddingLeft); paddingLeft != "" {
		innerTd.AddStyle(constants.CSSPaddingLeft, paddingLeft)
	}
	if paddingRight := c.getAttribute(constants.MJMLPaddingRight); paddingRight != "" {
		innerTd.AddStyle(constants.CSSPaddingRight, paddingRight)
	}
	if paddingTop := c.getAttribute(constants.MJMLPaddingTop); paddingTop != "" {
		innerTd.AddStyle(constants.CSSPaddingTop, paddingTop)
	}

	innerTd.AddStyle("text-align", textAlign)

	if err := innerTd.RenderOpen(w); err != nil {
		return err
	}

	// MSO conditional for wrapper content
	if err := html.RenderMSOWrapperTableOpen(w, effectiveWidth); err != nil {
		return err
	}

	// Render children with standard body width
	// Add MSO section transitions between section children (like MRML does)
	for i, child := range c.Children {
		if child.IsRawElement() {
			if err := html.RenderMSOSectionTransition(w, GetDefaultBodyWidthPixels()); err != nil {
				return err
			}
			if err := child.Render(w); err != nil {
				return err
			}
			continue
		}

		// Add MSO section transition between successive sections
		if i > 0 && child.GetTagName() == "mj-section" {
			if err := html.RenderMSOSectionTransition(w, GetDefaultBodyWidthPixels()); err != nil {
				return err
			}
		}

		// AIDEV-NOTE: width-flow-parent-to-child; pass reduced width to child (accounts for wrapper padding)
		child.SetContainerWidth(effectiveWidth)
		if err := child.Render(w); err != nil {
			return err
		}
	}

	if err := html.RenderMSOWrapperTableClose(w); err != nil {
		return err
	}

	if err := innerTd.RenderClose(w); err != nil {
		return err
	}
	if _, err := w.WriteString("</tr></tbody>"); err != nil {
		return err
	}
	if err := innerTable.RenderClose(w); err != nil {
		return err
	}
	if err := innerDiv.RenderClose(w); err != nil {
		return err
	}

	// Close MSO conditional
	if err := html.RenderMSOTableCloseConditional(w, msoTd, msoTable); err != nil {
		return err
	}

	// Close outer table
	if _, err := w.WriteString("</td></tr></tbody>"); err != nil {
		return err
	}
	if err := outerTable.RenderClose(w); err != nil {
		return err
	}

	return nil
}

// renderSimpleToWriter writes simple wrapper directly to Writer
func (c *MJWrapperComponent) renderSimpleToWriter(w io.StringWriter) error {
	// Get wrapper attributes
	padding := c.getAttribute("padding")
	textAlign := c.getAttribute("text-align")
	direction := c.getAttribute("direction")
	effectiveWidth := c.getEffectiveWidth()

	// MSO conditional table wrapper (should use full default body width, not effective width)
	msoTable := html.NewHTMLTag("table").
		AddAttribute("border", "0").
		AddAttribute("cellpadding", "0").
		AddAttribute("cellspacing", "0").
		AddAttribute("role", "presentation")

	// Add bgcolor to MSO table if background-color is set
	if bgColor := c.getAttribute("background-color"); bgColor != "" {
		msoTable.AddAttribute(constants.AttrBgcolor, bgColor)
	}

	msoTable.AddAttribute("align", "center").
		AddAttribute("width", strconv.Itoa(GetDefaultBodyWidthPixels())).
		AddStyle("width", GetDefaultBodyWidth())

	// Add css-class support for MSO table (MRML adds -outlook suffix)
	if cssClass := c.getAttribute("css-class"); cssClass != "" {
		msoTable.AddAttribute("class", cssClass+"-outlook")
	}

	msoTd := html.NewHTMLTag("td").
		AddStyle("line-height", "0px").
		AddStyle("font-size", "0px").
		AddStyle("mso-line-height-rule", "exactly")

	if err := html.RenderMSOTableOpenConditional(w, msoTable, msoTd); err != nil {
		return err
	}

	// Main wrapper div (match MRML property order: background first, then margin, border-radius, max-width)
	wrapperDiv := html.NewHTMLTag("div")
	c.AddDebugAttribute(wrapperDiv, "wrapper")

	// Apply background styles first to match MRML order
	c.ApplyBackgroundStyles(wrapperDiv, c)

	wrapperDiv.AddStyle("margin", "0px auto")

	// Add css-class support for wrapper div
	if cssClass := c.getAttribute("css-class"); cssClass != "" {
		wrapperDiv.AddAttribute("class", cssClass)
	}

	// Add border-radius before max-width to match MRML order
	if borderRadius := c.getAttribute("border-radius"); borderRadius != "" {
		wrapperDiv.AddStyle("border-radius", borderRadius)
	}

	wrapperDiv.AddStyle("max-width", GetDefaultBodyWidth())

	if err := wrapperDiv.RenderOpen(w); err != nil {
		return err
	}

	// Inner table (match MRML order: background first, then width, border-radius)
	innerTable := html.NewHTMLTag("table").
		AddAttribute("border", "0").
		AddAttribute("cellpadding", "0").
		AddAttribute("cellspacing", "0").
		AddAttribute("role", "presentation").
		AddAttribute("align", "center")

		// Apply background styles first to match MRML order
	c.ApplyBackgroundStyles(innerTable, c)

	innerTable.AddStyle("width", "100%")

	// Add border-radius after width to match MRML order
	if borderRadius := c.getAttribute("border-radius"); borderRadius != "" {
		innerTable.AddStyle("border-radius", borderRadius)
	}

	if err := innerTable.RenderOpen(w); err != nil {
		return err
	}
	if _, err := w.WriteString("<tbody><tr>"); err != nil {
		return err
	}

	// Main TD with wrapper styles (match MRML order: border first, then other properties)
	mainTd := html.NewHTMLTag("td")

	// Add border first to match MRML order
	if border := c.getAttribute("border"); border != "" {
		mainTd.AddStyle("border", border)
	}

	if borderBottom := c.getAttribute("border-bottom"); borderBottom != "" {
		mainTd.AddStyle("border-bottom", borderBottom)
	}
	if borderLeft := c.getAttribute("border-left"); borderLeft != "" {
		mainTd.AddStyle("border-left", borderLeft)
	}
	if borderRight := c.getAttribute("border-right"); borderRight != "" {
		mainTd.AddStyle("border-right", borderRight)
	}
	if borderTop := c.getAttribute("border-top"); borderTop != "" {
		mainTd.AddStyle("border-top", borderTop)
	}

	mainTd.AddStyle("direction", direction).
		AddStyle("font-size", "0px").
		AddStyle("padding", padding)

	// Add individual padding properties after shorthand to match MRML order
	if paddingBottom := c.getAttribute(constants.MJMLPaddingBottom); paddingBottom != "" {
		mainTd.AddStyle(constants.CSSPaddingBottom, paddingBottom)
	}
	if paddingLeft := c.getAttribute(constants.MJMLPaddingLeft); paddingLeft != "" {
		mainTd.AddStyle(constants.CSSPaddingLeft, paddingLeft)
	}
	if paddingRight := c.getAttribute(constants.MJMLPaddingRight); paddingRight != "" {
		mainTd.AddStyle(constants.CSSPaddingRight, paddingRight)
	}
	if paddingTop := c.getAttribute(constants.MJMLPaddingTop); paddingTop != "" {
		mainTd.AddStyle(constants.CSSPaddingTop, paddingTop)
	}

	mainTd.AddStyle("text-align", textAlign)

	if err := mainTd.RenderOpen(w); err != nil {
		return err
	}

	// For basic wrapper, we need a specific MSO conditional pattern
	// that matches MRML's output more closely - use original body width for wrapper MSO
	if err := html.RenderMSOWrapperTableOpen(w, GetDefaultBodyWidthPixels()); err != nil {
		return err
	}

	// Render children - pass the effective width (600px - border width)
	// Add MSO section transitions between section children (like MRML does)
	for i, child := range c.Children {
		if child.IsRawElement() {
			if err := html.RenderMSOSectionTransition(w, GetDefaultBodyWidthPixels()); err != nil {
				return err
			}
			if err := child.Render(w); err != nil {
				return err
			}
			continue
		}

		// Add MSO section transition between sections (but not before the first section)
		if i > 0 && child.GetTagName() == "mj-section" {
			if err := html.RenderMSOSectionTransition(w, GetDefaultBodyWidthPixels()); err != nil {
				return err
			}
		}

		// AIDEV-NOTE: width-flow-parent-to-child; pass reduced width to child (accounts for wrapper padding)
		child.SetContainerWidth(effectiveWidth)
		if err := child.Render(w); err != nil {
			return err
		}
	}

	if err := html.RenderMSOWrapperTableClose(w); err != nil {
		return err
	}

	if err := mainTd.RenderClose(w); err != nil {
		return err
	}
	if _, err := w.WriteString("</tr></tbody>"); err != nil {
		return err
	}
	if err := innerTable.RenderClose(w); err != nil {
		return err
	}
	if err := wrapperDiv.RenderClose(w); err != nil {
		return err
	}

	// Close MSO conditional
	if err := html.RenderMSOTableCloseConditional(w, msoTd, msoTable); err != nil {
		return err
	}

	return nil
}

func (c *MJWrapperComponent) GetTagName() string {
	return "mj-wrapper"
}
