package components

import (
	"fmt"
	"io"
	"strings"

	"github.com/preslavrachev/gomjml/mjml/html"
	"github.com/preslavrachev/gomjml/mjml/options"
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

// getBorderWidth calculates the total border width from border attribute
func (c *MJWrapperComponent) getBorderWidth() int {
	border := c.getAttribute("border")
	if border == "" {
		return 0
	}

	// Parse border width (e.g., "2px solid #333" -> 2)
	// Simple parsing for the common case
	if strings.Contains(border, "1px") {
		return 2 // 1px on each side
	}
	if strings.Contains(border, "2px") {
		return 4 // 2px on each side
	}
	if strings.Contains(border, "3px") {
		return 6 // 3px on each side
	}
	return 0
}

// getEffectiveWidth calculates width minus border width
func (c *MJWrapperComponent) getEffectiveWidth() int {
	baseWidth := GetDefaultBodyWidthPixels()
	borderWidth := c.getBorderWidth()
	return baseWidth - borderWidth
}

func (c *MJWrapperComponent) isFullWidth() bool {
	// Full width only if explicitly set
	return c.getAttribute("full-width") == "full-width"
}

func (c *MJWrapperComponent) renderFullWidth() (string, error) {
	var output strings.Builder

	// Get wrapper attributes
	padding := c.getAttribute("padding")
	textAlign := c.getAttribute("text-align")
	direction := c.getAttribute("direction")

	// Outer full-width table (MRML pattern)
	outerTable := html.NewHTMLTag("table").
		AddAttribute("border", "0").
		AddAttribute("cellpadding", "0").
		AddAttribute("cellspacing", "0").
		AddAttribute("role", "presentation").
		AddAttribute("align", "center")

	// Apply background styles to outer table and add width:100%
	c.ApplyBackgroundStyles(outerTable)
	outerTable.AddStyle("width", "100%")

	output.WriteString(outerTable.RenderOpen())
	output.WriteString("<tbody><tr><td>")

	// MSO conditional for inner container
	msoTable := html.NewHTMLTag("table").
		AddAttribute("border", "0").
		AddAttribute("cellpadding", "0").
		AddAttribute("cellspacing", "0").
		AddAttribute("role", "presentation")

	// Add bgcolor to MSO table if background-color is set
	if bgColor := c.getAttribute("background-color"); bgColor != "" {
		msoTable.AddAttribute("bgcolor", bgColor)
	}

	msoTable.AddAttribute("align", "center").
		AddAttribute("width", fmt.Sprintf("%d", GetDefaultBodyWidthPixels())).
		AddStyle("width", GetDefaultBodyWidth())

	// Add css-class support for MSO table (MRML adds -outlook suffix)
	if cssClass := c.getAttribute("css-class"); cssClass != "" {
		msoTable.AddAttribute("class", cssClass+"-outlook")
	}

	msoTd := html.NewHTMLTag("td").
		AddStyle("line-height", "0px").
		AddStyle("font-size", "0px").
		AddStyle("mso-line-height-rule", "exactly")

	output.WriteString(html.RenderMSOConditional(
		msoTable.RenderOpen() + "<tr>" + msoTd.RenderOpen()))

	// Inner constrained div (standard MRML pattern)
	innerDiv := html.NewHTMLTag("div").
		AddStyle("margin", "0px auto").
		AddStyle("max-width", GetDefaultBodyWidth())

	// Add css-class support for inner div
	if cssClass := c.getAttribute("css-class"); cssClass != "" {
		innerDiv.AddAttribute("class", cssClass)
	}

	output.WriteString(innerDiv.RenderOpen())

	// Inner table with content
	innerTable := html.NewHTMLTag("table").
		AddAttribute("border", "0").
		AddAttribute("cellpadding", "0").
		AddAttribute("cellspacing", "0").
		AddAttribute("role", "presentation").
		AddAttribute("align", "center").
		AddStyle("width", "100%")

	output.WriteString(innerTable.RenderOpen())
	output.WriteString("<tbody><tr>")

	// Inner TD
	innerTd := html.NewHTMLTag("td").
		AddStyle("direction", direction).
		AddStyle("font-size", "0px").
		AddStyle("padding", padding)

	// Add individual padding properties after shorthand to match MRML order (bottom first, then top)
	if paddingBottom := c.getAttribute("padding-bottom"); paddingBottom != "" {
		innerTd.AddStyle("padding-bottom", paddingBottom)
	}
	if paddingTop := c.getAttribute("padding-top"); paddingTop != "" {
		innerTd.AddStyle("padding-top", paddingTop)
	}

	innerTd.AddStyle("text-align", textAlign)

	output.WriteString(innerTd.RenderOpen())

	// MSO conditional for wrapper content
	output.WriteString(html.RenderMSOConditional(
		fmt.Sprintf(
			"<table border=\"0\" cellpadding=\"0\" cellspacing=\"0\" role=\"presentation\"><tr><td width=\"%dpx\">",
			GetDefaultBodyWidthPixels(),
		),
	))

	// Render children with standard body width
	for _, child := range c.Children {
		child.SetContainerWidth(GetDefaultBodyWidthPixels())
		if err := child.Render(&output); err != nil {
			return "", err
		}
	}

	output.WriteString(html.RenderMSOConditional("</td></tr></table>"))

	output.WriteString(innerTd.RenderClose())
	output.WriteString("</tr></tbody>")
	output.WriteString(innerTable.RenderClose())
	output.WriteString(innerDiv.RenderClose())

	// Close MSO conditional
	output.WriteString(html.RenderMSOConditional(msoTd.RenderClose() + "</tr>" + msoTable.RenderClose()))

	// Close outer table
	output.WriteString("</td></tr></tbody>")
	output.WriteString(outerTable.RenderClose())

	return output.String(), nil
}

func (c *MJWrapperComponent) renderSimple() (string, error) {
	var output strings.Builder

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
		msoTable.AddAttribute("bgcolor", bgColor)
	}

	msoTable.AddAttribute("align", "center").
		AddAttribute("width", fmt.Sprintf("%d", GetDefaultBodyWidthPixels())).
		AddStyle("width", GetDefaultBodyWidth())

	// Add css-class support for MSO table (MRML adds -outlook suffix)
	if cssClass := c.getAttribute("css-class"); cssClass != "" {
		msoTable.AddAttribute("class", cssClass+"-outlook")
	}

	msoTd := html.NewHTMLTag("td").
		AddStyle("line-height", "0px").
		AddStyle("font-size", "0px").
		AddStyle("mso-line-height-rule", "exactly")

	output.WriteString(html.RenderMSOConditional(
		msoTable.RenderOpen() + "<tr>" + msoTd.RenderOpen()))

	// Main wrapper div (match MRML property order: background first, then margin, border-radius, max-width)
	wrapperDiv := html.NewHTMLTag("div")
	c.AddDebugAttribute(wrapperDiv, "wrapper")

	// Apply background styles first to match MRML order
	c.ApplyBackgroundStyles(wrapperDiv)

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

	output.WriteString(wrapperDiv.RenderOpen())

	// Inner table (match MRML order: background first, then width, border-radius)
	innerTable := html.NewHTMLTag("table").
		AddAttribute("border", "0").
		AddAttribute("cellpadding", "0").
		AddAttribute("cellspacing", "0").
		AddAttribute("role", "presentation").
		AddAttribute("align", "center")

	// Apply background styles first to match MRML order
	c.ApplyBackgroundStyles(innerTable)

	innerTable.AddStyle("width", "100%")

	// Add border-radius after width to match MRML order
	if borderRadius := c.getAttribute("border-radius"); borderRadius != "" {
		innerTable.AddStyle("border-radius", borderRadius)
	}

	output.WriteString(innerTable.RenderOpen())
	output.WriteString("<tbody><tr>")

	// Main TD with wrapper styles (match MRML order: border first, then other properties)
	mainTd := html.NewHTMLTag("td")

	// Add border first to match MRML order
	if border := c.getAttribute("border"); border != "" {
		mainTd.AddStyle("border", border)
	}

	mainTd.AddStyle("direction", direction).
		AddStyle("font-size", "0px").
		AddStyle("padding", padding)

	// Add individual padding properties after shorthand to match MRML order (bottom first, then top)
	if paddingBottom := c.getAttribute("padding-bottom"); paddingBottom != "" {
		mainTd.AddStyle("padding-bottom", paddingBottom)
	}
	if paddingTop := c.getAttribute("padding-top"); paddingTop != "" {
		mainTd.AddStyle("padding-top", paddingTop)
	}

	mainTd.AddStyle("text-align", textAlign)

	output.WriteString(mainTd.RenderOpen())

	// For basic wrapper, we need a specific MSO conditional pattern
	// that matches MRML's output more closely - use original body width for wrapper MSO
	output.WriteString(html.RenderMSOConditional(
		fmt.Sprintf(
			"<table border=\"0\" cellpadding=\"0\" cellspacing=\"0\" role=\"presentation\"><tr><td width=\"%dpx\">",
			GetDefaultBodyWidthPixels(),
		),
	))

	// Render children - pass the effective width (600px - border width)
	for _, child := range c.Children {
		child.SetContainerWidth(effectiveWidth)
		if err := child.Render(&output); err != nil {
			return "", err
		}
	}

	output.WriteString(html.RenderMSOConditional("</td></tr></table>"))

	output.WriteString(mainTd.RenderClose())
	output.WriteString("</tr></tbody>")
	output.WriteString(innerTable.RenderClose())
	output.WriteString(wrapperDiv.RenderClose())

	// Close MSO conditional
	output.WriteString(html.RenderMSOConditional(msoTd.RenderClose() + "</tr>" + msoTable.RenderClose()))

	return output.String(), nil
}

// Render implements optimized Writer-based rendering for MJWrapperComponent
func (c *MJWrapperComponent) Render(w io.Writer) error {
	if c.isFullWidth() {
		return c.renderFullWidthToWriter(w)
	}
	return c.renderSimpleToWriter(w)
}

// renderFullWidthToWriter writes full-width wrapper directly to Writer
func (c *MJWrapperComponent) renderFullWidthToWriter(w io.Writer) error {
	// Get wrapper attributes
	padding := c.getAttribute("padding")
	textAlign := c.getAttribute("text-align")
	direction := c.getAttribute("direction")

	// Outer full-width table (MRML pattern)
	outerTable := html.NewHTMLTag("table").
		AddAttribute("border", "0").
		AddAttribute("cellpadding", "0").
		AddAttribute("cellspacing", "0").
		AddAttribute("role", "presentation").
		AddAttribute("align", "center")

	// Apply background styles to outer table and add width:100%
	c.ApplyBackgroundStyles(outerTable)
	outerTable.AddStyle("width", "100%")

	if _, err := w.Write([]byte(outerTable.RenderOpen())); err != nil {
		return err
	}
	if _, err := w.Write([]byte("<tbody><tr><td>")); err != nil {
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
		msoTable.AddAttribute("bgcolor", bgColor)
	}

	msoTable.AddAttribute("align", "center").
		AddAttribute("width", fmt.Sprintf("%d", GetDefaultBodyWidthPixels())).
		AddStyle("width", GetDefaultBodyWidth())

	// Add css-class support for MSO table (MRML adds -outlook suffix)
	if cssClass := c.getAttribute("css-class"); cssClass != "" {
		msoTable.AddAttribute("class", cssClass+"-outlook")
	}

	msoTd := html.NewHTMLTag("td").
		AddStyle("line-height", "0px").
		AddStyle("font-size", "0px").
		AddStyle("mso-line-height-rule", "exactly")

	if _, err := w.Write([]byte(html.RenderMSOConditional(
		msoTable.RenderOpen() + "<tr>" + msoTd.RenderOpen()))); err != nil {
		return err
	}

	// Inner constrained div (standard MRML pattern)
	innerDiv := html.NewHTMLTag("div").
		AddStyle("margin", "0px auto").
		AddStyle("max-width", GetDefaultBodyWidth())

	// Add css-class support for inner div
	if cssClass := c.getAttribute("css-class"); cssClass != "" {
		innerDiv.AddAttribute("class", cssClass)
	}

	if _, err := w.Write([]byte(innerDiv.RenderOpen())); err != nil {
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

	if _, err := w.Write([]byte(innerTable.RenderOpen())); err != nil {
		return err
	}
	if _, err := w.Write([]byte("<tbody><tr>")); err != nil {
		return err
	}

	// Inner TD
	innerTd := html.NewHTMLTag("td").
		AddStyle("direction", direction).
		AddStyle("font-size", "0px").
		AddStyle("padding", padding)

	// Add individual padding properties after shorthand to match MRML order (bottom first, then top)
	if paddingBottom := c.getAttribute("padding-bottom"); paddingBottom != "" {
		innerTd.AddStyle("padding-bottom", paddingBottom)
	}
	if paddingTop := c.getAttribute("padding-top"); paddingTop != "" {
		innerTd.AddStyle("padding-top", paddingTop)
	}

	innerTd.AddStyle("text-align", textAlign)

	if _, err := w.Write([]byte(innerTd.RenderOpen())); err != nil {
		return err
	}

	// MSO conditional for wrapper content
	if _, err := w.Write([]byte(html.RenderMSOConditional(
		fmt.Sprintf("<table border=\"0\" cellpadding=\"0\" cellspacing=\"0\" role=\"presentation\"><tr><td width=\"%dpx\">", GetDefaultBodyWidthPixels())))); err != nil {
		return err
	}

	// Render children with standard body width
	for _, child := range c.Children {
		child.SetContainerWidth(GetDefaultBodyWidthPixels())
		if err := child.Render(w); err != nil {
			return err
		}
	}

	if _, err := w.Write([]byte(html.RenderMSOConditional("</td></tr></table>"))); err != nil {
		return err
	}

	if _, err := w.Write([]byte(innerTd.RenderClose())); err != nil {
		return err
	}
	if _, err := w.Write([]byte("</tr></tbody>")); err != nil {
		return err
	}
	if _, err := w.Write([]byte(innerTable.RenderClose())); err != nil {
		return err
	}
	if _, err := w.Write([]byte(innerDiv.RenderClose())); err != nil {
		return err
	}

	// Close MSO conditional
	if _, err := w.Write([]byte(html.RenderMSOConditional(msoTd.RenderClose() + "</tr>" + msoTable.RenderClose()))); err != nil {
		return err
	}

	// Close outer table
	if _, err := w.Write([]byte("</td></tr></tbody>")); err != nil {
		return err
	}
	if _, err := w.Write([]byte(outerTable.RenderClose())); err != nil {
		return err
	}

	return nil
}

// renderSimpleToWriter writes simple wrapper directly to Writer
func (c *MJWrapperComponent) renderSimpleToWriter(w io.Writer) error {
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
		msoTable.AddAttribute("bgcolor", bgColor)
	}

	msoTable.AddAttribute("align", "center").
		AddAttribute("width", fmt.Sprintf("%d", GetDefaultBodyWidthPixels())).
		AddStyle("width", GetDefaultBodyWidth())

	// Add css-class support for MSO table (MRML adds -outlook suffix)
	if cssClass := c.getAttribute("css-class"); cssClass != "" {
		msoTable.AddAttribute("class", cssClass+"-outlook")
	}

	msoTd := html.NewHTMLTag("td").
		AddStyle("line-height", "0px").
		AddStyle("font-size", "0px").
		AddStyle("mso-line-height-rule", "exactly")

	if _, err := w.Write([]byte(html.RenderMSOConditional(
		msoTable.RenderOpen() + "<tr>" + msoTd.RenderOpen()))); err != nil {
		return err
	}

	// Main wrapper div (match MRML property order: background first, then margin, border-radius, max-width)
	wrapperDiv := html.NewHTMLTag("div")
	c.AddDebugAttribute(wrapperDiv, "wrapper")

	// Apply background styles first to match MRML order
	c.ApplyBackgroundStyles(wrapperDiv)

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

	if _, err := w.Write([]byte(wrapperDiv.RenderOpen())); err != nil {
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
	c.ApplyBackgroundStyles(innerTable)

	innerTable.AddStyle("width", "100%")

	// Add border-radius after width to match MRML order
	if borderRadius := c.getAttribute("border-radius"); borderRadius != "" {
		innerTable.AddStyle("border-radius", borderRadius)
	}

	if _, err := w.Write([]byte(innerTable.RenderOpen())); err != nil {
		return err
	}
	if _, err := w.Write([]byte("<tbody><tr>")); err != nil {
		return err
	}

	// Main TD with wrapper styles (match MRML order: border first, then other properties)
	mainTd := html.NewHTMLTag("td")

	// Add border first to match MRML order
	if border := c.getAttribute("border"); border != "" {
		mainTd.AddStyle("border", border)
	}

	mainTd.AddStyle("direction", direction).
		AddStyle("font-size", "0px").
		AddStyle("padding", padding)

	// Add individual padding properties after shorthand to match MRML order (bottom first, then top)
	if paddingBottom := c.getAttribute("padding-bottom"); paddingBottom != "" {
		mainTd.AddStyle("padding-bottom", paddingBottom)
	}
	if paddingTop := c.getAttribute("padding-top"); paddingTop != "" {
		mainTd.AddStyle("padding-top", paddingTop)
	}

	mainTd.AddStyle("text-align", textAlign)

	if _, err := w.Write([]byte(mainTd.RenderOpen())); err != nil {
		return err
	}

	// For basic wrapper, we need a specific MSO conditional pattern
	// that matches MRML's output more closely - use original body width for wrapper MSO
	if _, err := w.Write([]byte(html.RenderMSOConditional(
		fmt.Sprintf("<table border=\"0\" cellpadding=\"0\" cellspacing=\"0\" role=\"presentation\"><tr><td width=\"%dpx\">", GetDefaultBodyWidthPixels())))); err != nil {
		return err
	}

	// Render children - pass the effective width (600px - border width)
	for _, child := range c.Children {
		child.SetContainerWidth(effectiveWidth)
		if err := child.Render(w); err != nil {
			return err
		}
	}

	if _, err := w.Write([]byte(html.RenderMSOConditional("</td></tr></table>"))); err != nil {
		return err
	}

	if _, err := w.Write([]byte(mainTd.RenderClose())); err != nil {
		return err
	}
	if _, err := w.Write([]byte("</tr></tbody>")); err != nil {
		return err
	}
	if _, err := w.Write([]byte(innerTable.RenderClose())); err != nil {
		return err
	}
	if _, err := w.Write([]byte(wrapperDiv.RenderClose())); err != nil {
		return err
	}

	// Close MSO conditional
	if _, err := w.Write([]byte(html.RenderMSOConditional(msoTd.RenderClose() + "</tr>" + msoTable.RenderClose()))); err != nil {
		return err
	}

	return nil
}

func (c *MJWrapperComponent) GetTagName() string {
	return "mj-wrapper"
}
