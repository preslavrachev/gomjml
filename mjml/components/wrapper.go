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

// getChildAlign returns the align attribute for a section child if specified.
// Only mj-section children can provide alignment for MSO wrapper tables.
func getChildAlign(child Component) string {
	if sec, ok := child.(*MJSectionComponent); ok {
		return sec.GetAttributeWithDefault(sec, "align")
	}
	return ""
}

func (c *MJWrapperComponent) hasRenderableChildren() bool {
	for _, child := range c.Children {
		if child.IsRawElement() {
			if raw, ok := child.(*MJRawComponent); ok {
				if strings.TrimSpace(raw.Content) == "" {
					continue
				}
			}
			return true
		}
		return true
	}
	return false
}

// shouldUseOuterOnlyMSOWrapper reports whether the wrapper's Outlook fallback should
// emit only the outer table, leaving the inner wrapper table to be handled by the
// section itself. This mirrors MJML's behavior for wrappers that exclusively contain
// full-width sections with background images—those sections open (and close) the
// inner MSO table inside their VML markup to ensure the background renders correctly
// in Outlook. Mixing such sections with standard sections would require both tables,
// so we only take this path when every non-raw child consumes the inner wrapper
// table.
func (c *MJWrapperComponent) shouldUseOuterOnlyMSOWrapper() bool {
	hasSection := false
	anyConsumer := false
	allConsumers := true

	for _, child := range c.Children {
		if child.IsRawElement() {
			continue
		}

		section, ok := child.(*MJSectionComponent)
		if !ok {
			// Wrappers are only expected to contain sections (and raw nodes).
			// If we encounter anything else, fall back to the standard
			// wrapper table structure for safety.
			return false
		}

		hasSection = true

		fullWidth := section.GetAttributeWithDefault(section, "full-width")
		backgroundURL := section.GetAttributeWithDefault(section, constants.MJMLBackgroundUrl)
		consumesWrapperTable := fullWidth != "" && backgroundURL != ""

		anyConsumer = anyConsumer || consumesWrapperTable
		if !consumesWrapperTable {
			allConsumers = false
		}
	}

	return hasSection && anyConsumer && allConsumers
}

func (c *MJWrapperComponent) hasFullWidthSectionChild() bool {
	for _, child := range c.Children {
		if child.IsRawElement() {
			continue
		}
		if section, ok := child.(*MJSectionComponent); ok {
			if section.GetAttributeWithDefault(section, "full-width") != "" {
				return true
			}
		}
	}
	return false
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
	cssClass := c.getAttribute("css-class")
	wrapperBgColor := c.getAttribute("background-color")

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

	if cssClass != "" {
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
	msoTable := html.NewHTMLTag("table")

	// Attribute order matches MJML output: align, border, cellpadding, cellspacing, class, role, style, width, bgcolor.
	msoTable.AddAttribute("align", "center")
	msoTable.AddAttribute("border", "0")
	msoTable.AddAttribute("cellpadding", "0")
	msoTable.AddAttribute("cellspacing", "0")

	if cssClass != "" {
		msoTable.AddAttribute("class", cssClass+"-outlook")
	} else {
		msoTable.AddAttribute("class", "")
	}

	msoTable.AddAttribute("role", "presentation")
	msoTable.AddAttribute("style", "width:"+GetDefaultBodyWidth()+";")
	msoTable.AddAttribute("width", strconv.Itoa(GetDefaultBodyWidthPixels()))

	// Add bgcolor to MSO table if background-color is set (after width to match expected order)
	if wrapperBgColor != "" {
		msoTable.AddAttribute(constants.AttrBgcolor, wrapperBgColor)
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
	firstAlign := ""
	firstBgColor := ""
	for _, ch := range c.Children {
		if ch.IsRawElement() {
			continue
		}
		firstAlign = getChildAlign(ch)
		if section, ok := ch.(*MJSectionComponent); ok {
			firstBgColor = section.GetAttributeWithDefault(section, "background-color")
		}
		break
	}

	splitMSOWrapper := c.hasFullWidthSectionChild()
	msoBgColor := firstBgColor
	if msoBgColor == "" && splitMSOWrapper {
		msoBgColor = c.getAttribute("background-color")
	}

	useOuterOnlyMSO := c.shouldUseOuterOnlyMSOWrapper()
	hasRenderableChildren := c.hasRenderableChildren()
	msoWrapperOpened := false
	delegatedWrapperBackground := false
	wrapperWidth := c.GetEffectiveWidth()
	if !hasRenderableChildren {
		if err := html.RenderMSOEmptyWrapperPlaceholder(w); err != nil {
			return err
		}
	} else if useOuterOnlyMSO {
		if err := html.RenderMSOWrapperOuterOpen(w, wrapperWidth, firstAlign, msoBgColor); err != nil {
			return err
		}
		msoWrapperOpened = true
	} else if splitMSOWrapper && msoBgColor != "" {
		if err := html.RenderMSOWrapperOuterOpen(w, wrapperWidth, firstAlign, ""); err != nil {
			return err
		}
		msoWrapperOpened = true
		delegatedWrapperBackground = true
	} else {
		if err := html.RenderMSOWrapperTableOpenWithWidths(w, wrapperWidth, effectiveWidth, firstAlign, firstBgColor); err != nil {
			return err
		}
		msoWrapperOpened = true
	}

	// Render children with standard body width
	// Add MSO section transitions between section children (like MRML does)
	wrapperMSOClosedByChild := false

	for i, child := range c.Children {
		if child.IsRawElement() {
			// Inject raw content inside the MSO transition block so Outlook maintains table structure
			if err := html.RenderMSOSectionTransitionWithContent(w, GetDefaultBodyWidthPixels(), effectiveWidth, "", func(sw io.StringWriter) error {
				return child.Render(sw)
			}); err != nil {
				return err
			}
			continue
		}

		// Add MSO section transition between successive sections
		if i > 0 && child.GetTagName() == "mj-section" && !c.Children[i-1].IsRawElement() {
			if err := html.RenderMSOSectionTransition(w, GetDefaultBodyWidthPixels(), getChildAlign(child)); err != nil {
				return err
			}
		}

		// AIDEV-NOTE: width-flow-parent-to-child; pass reduced width to child (accounts for wrapper padding)
		child.SetContainerWidth(effectiveWidth)

		if delegatedWrapperBackground {
			if sectionComp, ok := child.(*MJSectionComponent); ok {
				if sectionComp.GetAttributeWithDefault(sectionComp, "full-width") != "" && sectionComp.GetAttributeWithDefault(sectionComp, constants.MJMLBackgroundUrl) == "" {
					sectionBg := sectionComp.GetAttributeWithDefault(sectionComp, "background-color")
					if sectionBg == "" {
						sectionBg = wrapperBgColor
					}
					if sectionBg != "" {
						alignForSection := getChildAlign(sectionComp)
						sectionComp.SetWrapperMSOBackground(sectionBg, alignForSection)
					}
				}
			}
		}
		if err := child.Render(w); err != nil {
			return err
		}
		if consumer, ok := child.(interface{ ConsumedWrapperMSOTable() bool }); ok && consumer.ConsumedWrapperMSOTable() {
			wrapperMSOClosedByChild = true
		}
	}

	if wrapperMSOClosedByChild {
		if err := html.RenderMSOConditional(w, "</td></tr></table>"); err != nil {
			return err
		}
	} else if msoWrapperOpened {
		if useOuterOnlyMSO || delegatedWrapperBackground {
			if err := html.RenderMSOConditional(w, "</td></tr></table>"); err != nil {
				return err
			}
		} else {
			if err := html.RenderMSOWrapperTableClose(w); err != nil {
				return err
			}
		}
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
	cssClass := c.getAttribute("css-class")
	borderRadius := c.getAttribute("border-radius")
	wrapperBgColor := c.getAttribute("background-color")
	effectiveWidth := c.getEffectiveWidth()

	hasBorder := false
	if c.getAttribute("border") != "" || c.getAttribute("border-left") != "" || c.getAttribute("border-right") != "" ||
		c.getAttribute("border-top") != "" || c.getAttribute("border-bottom") != "" {
		hasBorder = true
	}

	// MSO conditional table wrapper (should use full default body width, not effective width)
	msoTable := html.NewHTMLTag("table")

	// Attribute order matches MJML output: align, border, cellpadding, cellspacing, class, role, style, width, bgcolor.
	msoTable.AddAttribute("align", "center")
	msoTable.AddAttribute("border", "0")
	msoTable.AddAttribute("cellpadding", "0")
	msoTable.AddAttribute("cellspacing", "0")

	if cssClass != "" {
		msoTable.AddAttribute("class", cssClass+"-outlook")
	} else {
		msoTable.AddAttribute("class", "")
	}

	msoTable.AddAttribute("role", "presentation")
	msoTable.AddAttribute("style", "width:"+GetDefaultBodyWidth()+";")
	msoTable.AddAttribute("width", strconv.Itoa(GetDefaultBodyWidthPixels()))

	// Add bgcolor to MSO table if background-color is set (after width to match expected order)
	if wrapperBgColor != "" {
		msoTable.AddAttribute(constants.AttrBgcolor, wrapperBgColor)
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

	// Add css-class support for wrapper div
	if cssClass != "" {
		wrapperDiv.AddAttribute("class", cssClass)
		c.ApplyInlineStyles(wrapperDiv, cssClass)
	}

	wrapperDiv.AddStyle("margin", "0px auto")

	// Order styles to match MJML output: margin -> max-width -> border-radius -> overflow
	wrapperDiv.AddStyle("max-width", GetDefaultBodyWidth())

	if borderRadius != "" {
		wrapperDiv.AddStyle("border-radius", borderRadius)
		wrapperDiv.AddStyle("overflow", "hidden")
	}

	if err := wrapperDiv.RenderOpen(w); err != nil {
		return err
	}

	// Inner table (match MJML order: background first, then width and border handling)
	innerTable := html.NewHTMLTag("table").
		AddAttribute("border", "0").
		AddAttribute("cellpadding", "0").
		AddAttribute("cellspacing", "0").
		AddAttribute("role", "presentation").
		AddAttribute("align", "center")

		// Apply background styles first to match MRML order
	c.ApplyBackgroundStyles(innerTable, c)

	innerTable.AddStyle("width", "100%")
	if hasBorder || borderRadius != "" {
		innerTable.AddStyle(constants.CSSBorderCollapse, constants.BorderCollapseSeparate)
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

	if borderRadius != "" {
		mainTd.AddStyle("border-radius", borderRadius)
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

	firstAlign := ""
	firstBgColor := ""
	for _, ch := range c.Children {
		if ch.IsRawElement() {
			continue
		}
		firstAlign = getChildAlign(ch)
		if section, ok := ch.(*MJSectionComponent); ok {
			firstBgColor = section.GetAttributeWithDefault(section, "background-color")
		}
		break
	}

	splitMSOWrapper := c.hasFullWidthSectionChild()
	msoBgColor := firstBgColor
	if msoBgColor == "" && splitMSOWrapper {
		msoBgColor = c.getAttribute("background-color")
	}

	// For basic wrapper, we need a specific MSO conditional pattern
	// that matches MRML's output more closely - use the outer container width
	outerWidth := c.GetEffectiveWidth()
	useOuterOnlyMSO := c.shouldUseOuterOnlyMSOWrapper()
	hasRenderableChildren := c.hasRenderableChildren()
	msoWrapperOpened := false
	delegatedWrapperBackground := false
	if !hasRenderableChildren {
		if err := html.RenderMSOEmptyWrapperPlaceholder(w); err != nil {
			return err
		}
	} else if useOuterOnlyMSO {
		if err := html.RenderMSOWrapperOuterOpen(w, outerWidth, firstAlign, msoBgColor); err != nil {
			return err
		}
		msoWrapperOpened = true
	} else if splitMSOWrapper && msoBgColor != "" {
		if err := html.RenderMSOWrapperOuterOpen(w, outerWidth, firstAlign, ""); err != nil {
			return err
		}
		msoWrapperOpened = true
		delegatedWrapperBackground = true
	} else {
		if err := html.RenderMSOWrapperTableOpenWithWidths(w, outerWidth, effectiveWidth, firstAlign, firstBgColor); err != nil {
			return err
		}
		msoWrapperOpened = true
	}

	// Render children - pass the effective width (600px - border width)
	// Add MSO section transitions between section children (like MJML does)

	wrapperMSOClosedByChild := false

	for i, child := range c.Children {
		if child.IsRawElement() {
			if err := html.RenderMSOSectionTransitionWithContent(w, outerWidth, effectiveWidth, "", func(sw io.StringWriter) error {
				return child.Render(sw)
			}); err != nil {
				return err
			}
			continue
		}

		// Add MSO section transition between sections (but not before the first section)
		if i > 0 && child.GetTagName() == "mj-section" && !c.Children[i-1].IsRawElement() {
			if err := html.RenderMSOSectionTransition(w, outerWidth, getChildAlign(child)); err != nil {
				return err
			}
		}

		// AIDEV-NOTE: width-flow-parent-to-child; pass reduced width to child (accounts for wrapper padding)
		child.SetContainerWidth(effectiveWidth)

		if delegatedWrapperBackground {
			if sectionComp, ok := child.(*MJSectionComponent); ok {
				if sectionComp.GetAttributeWithDefault(sectionComp, "full-width") != "" && sectionComp.GetAttributeWithDefault(sectionComp, constants.MJMLBackgroundUrl) == "" {
					sectionBg := sectionComp.GetAttributeWithDefault(sectionComp, "background-color")
					if sectionBg == "" {
						sectionBg = wrapperBgColor
					}
					if sectionBg != "" {
						alignForSection := getChildAlign(sectionComp)
						sectionComp.SetWrapperMSOBackground(sectionBg, alignForSection)
					}
				}
			}
		}
		if err := child.Render(w); err != nil {
			return err
		}
		if consumer, ok := child.(interface{ ConsumedWrapperMSOTable() bool }); ok && consumer.ConsumedWrapperMSOTable() {
			wrapperMSOClosedByChild = true
		}
	}

	if wrapperMSOClosedByChild {
		if err := html.RenderMSOConditional(w, "</td></tr></table>"); err != nil {
			return err
		}
	} else if msoWrapperOpened {
		if useOuterOnlyMSO || delegatedWrapperBackground {
			if err := html.RenderMSOConditional(w, "</td></tr></table>"); err != nil {
				return err
			}
		} else {
			if err := html.RenderMSOWrapperTableClose(w); err != nil {
				return err
			}
		}
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
